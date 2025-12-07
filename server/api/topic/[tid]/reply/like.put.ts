import { prisma } from '~~/prisma/prisma'
import { updateReplyLikeSchema } from '~/validations/topic'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const input = await kunParsePutBody(event, updateReplyLikeSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const reply = await prisma.topic_reply.findUnique({
    where: { id: input.replyId },
    include: {
      like: {
        where: {
          user_id: uid
        }
      }
    }
  })
  if (!reply) {
    return kunError(event, '未找到该回复')
  }
  if (reply.user_id === uid) {
    return kunError(event, '您不能给自己点赞')
  }

  const isLikedReply = reply.like.length > 0

  return prisma.$transaction(async (prisma) => {
    if (isLikedReply) {
      await prisma.topic_reply_like.delete({
        where: {
          user_id_topic_reply_id: {
            user_id: uid,
            topic_reply_id: input.replyId
          }
        }
      })
    } else {
      await prisma.topic_reply_like.create({
        data: {
          user_id: uid,
          topic_reply_id: input.replyId
        }
      })
    }

    await prisma.user.update({
      where: { id: reply.user_id },
      data: { moemoepoint: { increment: isLikedReply ? -1 : 1 } }
    })

    await createDedupMessage(
      prisma,
      uid,
      reply.user_id,
      'liked',
      reply.content.slice(0, 233),
      reply.topic_id
    )

    return 'MOEMOE like reply successfully!'
  })
})
