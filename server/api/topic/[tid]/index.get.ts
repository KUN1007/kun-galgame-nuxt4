import { prisma } from '~~/prisma/prisma'
import { getNSFWCookie } from '~~/server/utils/getNSFWCookie'
import { getTopicDetailSchema } from '~/validations/topic'
import type { TopicDetail, TopicBestAnswer } from '~/types/api/topic'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getTopicDetailSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  const uid = userInfo?.uid

  const [data] = await Promise.all([
    prisma.topic.findUnique({
      where: { id: input.topicId },
      include: {
        best_answer: {
          include: {
            user: {
              select: { id: true, name: true, avatar: true, moemoepoint: true }
            },
            target: { select: { content: true } }
          }
        },
        user: {
          select: { id: true, name: true, avatar: true, moemoepoint: true }
        },
        like: { where: { user_id: uid } },
        dislike: { where: { user_id: uid } },
        favorite: { where: { user_id: uid } },
        upvote: { where: { user_id: uid } },
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
            dislike: true,
            favorite: true,
            upvote: true,
            reply: true,
            poll: true
          }
        }
      }
    }),

    prisma.topic.update({
      where: { id: input.topicId },
      data: { view: { increment: 1 } }
    })
  ])
  if (!data) {
    return kunError(event, '未找到该话题')
  }
  const nsfw = getNSFWCookie(event)
  if ((nsfw === 'sfw' && data.is_nsfw) || data.status === 1) {
    return 'banned'
  }

  let bestAnswer: TopicBestAnswer | null = null
  if (data.best_answer) {
    const ba = data.best_answer
    bestAnswer = {
      id: ba.id,
      topicId: ba.topic_id,
      floor: ba.floor,
      user: ba.user,
      edited: ba.edited,
      created: ba.created,
      contentMarkdown: `${ba.content}${ba.target.map((t) => t.content).join('')}`,
      contentHtml: await markdownToHtml(ba.content)
    }
  }

  const topic: TopicDetail = {
    id: data.id,
    title: data.title,
    view: data.view,
    status: data.status,
    isNSFW: data.is_nsfw,
    category: data.category,
    section: data.section.map((s) => s.topic_section.name),
    tag: data.tag,
    user: data.user,

    likeCount: data._count.like,
    isLiked: data.like.length > 0,
    dislikeCount: data._count.dislike,
    isDisliked: data.dislike.length > 0,
    favoriteCount: data._count.favorite,
    isFavorited: data.favorite.length > 0,
    upvoteCount: data._count.upvote,
    isUpvoted: data.upvote.length > 0,

    replyCount: data._count.reply,
    bestAnswer,
    isPollTopic: !!data._count.poll,

    contentHtml: await markdownToHtml(data.content),
    contentMarkdown: data.content,
    statusUpdateTime: data.status_update_time,
    upvoteTime: data.upvote_time,
    edited: data.edited,
    created: data.created
  }

  return topic
})
