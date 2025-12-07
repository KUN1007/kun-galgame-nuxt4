import { prisma } from '~~/prisma/prisma'
import { getNSFWCookie } from '~~/server/utils/getNSFWCookie'
import { getTopicRankingSchema } from '~/validations/ranking'
import type { RankingTopicItem } from '~/types/api/ranking'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getTopicRankingSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { page, limit, sortField, sortOrder } = input
  const nsfw = getNSFWCookie(event)
  const isSFW = nsfw === 'sfw'
  const skip = (page - 1) * limit

  const countFields = ['upvote', 'like', 'reply', 'comment', 'favorite']
  const isCountField = countFields.includes(sortField)

  const select = {
    id: true,
    title: true,
    created: true,
    user: {
      select: {
        id: true,
        name: true,
        avatar: true
      }
    },
    ...(sortField === 'view' && { view: true }),
    ...(isCountField && {
      _count: {
        select: {
          [sortField]: true
        }
      }
    })
  }

  const topics = await prisma.topic.findMany({
    where: { status: { not: 1 }, ...(isSFW ? { is_nsfw: false } : {}) },
    skip,
    take: limit,
    orderBy: isCountField
      ? { [sortField]: { _count: sortOrder } }
      : { [sortField]: sortOrder },
    select: select
  })

  const items: RankingTopicItem[] = topics.map((topic) => {
    let value: number
    if (sortField === 'view') {
      value = topic.view
    } else {
      value = topic._count[sortField]
    }

    return {
      id: topic.id,
      title: topic.title,
      user: topic.user,
      value,
      sortField
    }
  })
  return items
})
