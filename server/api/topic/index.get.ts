import { prisma } from '~~/prisma/prisma'
import { getNSFWCookie } from '~~/server/utils/getNSFWCookie'
import { getTopicSchema } from '~/validations/topic'
import type { TopicCard } from '~/types/api/topic'
import type { Prisma } from '~~/prisma/generated/prisma/client'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getTopicSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { page, limit, sortField, sortOrder, category } = input

  let orderBy: Prisma.topicOrderByWithRelationInput = {}
  if (
    sortField === 'created' ||
    sortField === 'view' ||
    sortField === 'status_update_time'
  ) {
    orderBy = {
      [sortField]: sortOrder
    }
  } else if (
    sortField === 'like' ||
    sortField === 'favorite' ||
    sortField === 'upvote'
  ) {
    orderBy = {
      [sortField]: {
        _count: sortOrder
      }
    }
  }

  const nsfw = getNSFWCookie(event)
  const isSFW = nsfw === 'sfw'

  const data = await prisma.topic.findMany({
    where: {
      category: category === 'all' ? undefined : category,
      status: { not: 1 },
      section: {
        none: {
          topic_section: {
            name: { in: ['g-seeking', 'g-other', 't-help'] }
          }
        }
      },
      ...(isSFW ? { is_nsfw: false } : {})
    },
    skip: (page - 1) * limit,
    take: limit,
    orderBy,
    select: {
      id: true,
      title: true,
      view: true,
      tag: true,
      status: true,
      status_update_time: true,
      upvote_time: true,
      best_answer: true,
      is_nsfw: true,

      user: {
        select: {
          id: true,
          name: true,
          avatar: true
        }
      },

      section: {
        select: {
          topic_section: {
            select: {
              name: true
            }
          }
        }
      },

      _count: {
        select: {
          like: true,
          reply: true,
          comment: true,
          poll: true
        }
      }
    }
  })

  const topics: TopicCard[] = data.map((topic) => ({
    id: topic.id,
    title: topic.title,
    view: topic.view,
    tag: topic.tag,
    user: topic.user as KunUser,
    status: topic.status,
    hasBestAnswer: !!topic.best_answer,
    isPollTopic: !!topic._count.poll,
    isNSFWTopic: topic.is_nsfw,

    likeCount: topic._count.like,
    replyCount: topic._count.reply,
    commentCount: topic._count.comment,

    section: topic.section.map((s) => s.topic_section.name),
    statusUpdateTime: topic.status_update_time,
    upvoteTime: topic.upvote_time
  }))

  return topics
})
