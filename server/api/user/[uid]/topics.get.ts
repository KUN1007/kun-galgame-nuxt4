import { prisma } from '~~/prisma/prisma'
import { getNSFWCookie } from '~~/server/utils/getNSFWCookie'
import { getUserTopicSchema } from '~/validations/user'
import type { UserTopic } from '~/types/api/user'
import type { Prisma } from '~~/prisma/generated/prisma/client'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getUserTopicSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { userId, page, limit, type } = input

  if (type === 'topic_hide') {
    const userInfo = await getCookieTokenInfo(event)
    if (userId !== userInfo?.uid) {
      return { topics: [], totalCount: 0 }
    }

    const [topics, totalCount]: [UserTopic[], number] =
      await prisma.$transaction([
        prisma.topic.findMany({
          where: { status: 1, user_id: userInfo.uid },
          skip: (page - 1) * limit,
          take: limit,
          orderBy: {
            created: 'desc'
          },
          select: {
            id: true,
            title: true,
            created: true
          }
        }),

        prisma.topic.count({ where: { status: 1, user_id: userInfo.uid } })
      ])

    return {
      topics,
      totalCount
    }
  }

  const nsfw = getNSFWCookie(event)
  const isSFW = nsfw === 'sfw'
  const where: Prisma.topicWhereInput = {
    status: { not: 1 },
    ...(isSFW ? { is_nsfw: false } : {})
  }

  switch (type) {
    case 'topic':
      where.user_id = userId
      break

    case 'topic_like':
      where.like = { some: { user_id: userId } }
      break

    case 'topic_upvote':
      where.upvote = { some: { user_id: userId } }
      break

    case 'topic_favorite':
      where.favorite = { some: { user_id: userId } }
      break
  }

  const [topics, totalCount]: [UserTopic[], number] = await prisma.$transaction(
    [
      prisma.topic.findMany({
        where,
        skip: (page - 1) * limit,
        take: limit,
        orderBy: {
          created: 'desc'
        },
        select: {
          id: true,
          title: true,
          created: true
        }
      }),

      prisma.topic.count({ where })
    ]
  )

  return {
    topics,
    totalCount
  }
})
