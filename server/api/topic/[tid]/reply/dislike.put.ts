import { prisma } from '~~/prisma/prisma'
import { updateReplyDislikeSchema } from '~/validations/topic'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const input = await kunParsePutBody(event, updateReplyDislikeSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const reply = await prisma.topic_reply.findUnique({
    where: { id: input.replyId },
    include: {
      dislike: {
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
    return kunError(event, '您不能给自己点踩')
  }

  const isDislikedReply = reply.dislike.length > 0

  if (isDislikedReply) {
    await prisma.topic_reply_dislike.delete({
      where: {
        user_id_topic_reply_id: {
          user_id: uid,
          topic_reply_id: input.replyId
        }
      }
    })
  } else {
    await prisma.topic_reply_dislike.create({
      data: {
        user_id: uid,
        topic_reply_id: input.replyId
      }
    })
  }

  return 'MOEMOE dislike reply successfully!'
})
