import { prisma } from '~~/prisma/prisma'
import { getUserRankingSchema } from '~/validations/ranking'
import type { RankingUserItem } from '~/types/api/ranking'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getUserRankingSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { page, limit, sortField, sortOrder } = input
  const skip = (page - 1) * limit

  const countFields = [
    'follower_relation',
    'topic',
    'reply_created',
    'comment_created',
    'galgame',
    'galgame_resource',
    'topic_like',
    'topic_upvote'
  ]
  const isCountField = countFields.includes(sortField)

  const select = {
    id: true,
    name: true,
    avatar: true,
    bio: true,
    created: true,
    ...(sortField === 'moemoepoint' && { moemoepoint: true }),
    ...(isCountField && {
      _count: {
        select: {
          [sortField]: true
        }
      }
    })
  }

  const users = await prisma.user.findMany({
    skip,
    take: limit,
    orderBy: isCountField
      ? { [sortField]: { _count: sortOrder } }
      : { [sortField]: sortOrder },
    select: select
  })

  const items: RankingUserItem[] = users.map((user) => {
    let value: number
    if (sortField === 'moemoepoint') {
      value = user.moemoepoint
    } else {
      value = user._count[sortField]
    }

    return {
      id: user.id,
      name: user.name,
      avatar: user.avatar,
      bio: user.bio,
      created: user.created,
      value,
      sortField
    }
  })

  return items
})
