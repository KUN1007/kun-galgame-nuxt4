import { prisma } from '~~/prisma/prisma'
import { getUserReplySchema } from '~/validations/user'
import type { Prisma } from '~~/prisma/generated/prisma/client'
import type { UserReply } from '~/types/api/user'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getUserReplySchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { userId, page, limit, type } = input
  const skip = (page - 1) * limit

  if (type === 'reply_target') {
    const where: Prisma.topic_reply_targetWhereInput = {
      target_reply: { user_id: userId },
      reply: { user_id: { not: userId } }
    }

    const [targets, totalCount] = await prisma.$transaction([
      prisma.topic_reply_target.findMany({
        where,
        select: {
          content: true,
          reply: {
            select: {
              topic_id: true,
              created: true
            }
          }
        },
        orderBy: { created: 'desc' },
        skip: skip,
        take: limit
      }),
      prisma.topic_reply_target.count({ where })
    ])

    const formattedReplies: UserReply[] = targets.map((target) => {
      return {
        topicId: target.reply.topic_id,
        content: target.content,
        created: target.reply.created
      }
    })

    return {
      replies: formattedReplies,
      totalCount: totalCount
    }
  }

  let where: Prisma.topic_replyWhereInput = {}
  switch (type) {
    case 'reply_created':
      where = { user_id: userId }
      break
    case 'reply_like':
      where = { like: { some: { user_id: userId } } }
      break
    default:
      return kunError(event, '无效的类型参数', 400)
  }

  const [replies, totalCount] = await prisma.$transaction([
    prisma.topic_reply.findMany({
      where,
      select: {
        topic_id: true,
        content: true,
        created: true,
        target: {
          select: {
            content: true
          }
        }
      },
      orderBy: { created: 'desc' },
      skip: skip,
      take: limit
    }),
    prisma.topic_reply.count({ where })
  ])

  const formattedReplies: UserReply[] = replies.flatMap((reply) => {
    if (reply.target && reply.target.length > 0) {
      return reply.target.map((t) => ({
        topicId: reply.topic_id,
        content: t.content,
        created: reply.created
      }))
    }

    if (reply.content) {
      return [
        {
          topicId: reply.topic_id,
          content: reply.content,
          created: reply.created
        }
      ]
    }

    return []
  })

  return {
    replies: formattedReplies,
    totalCount: totalCount
  }
})
