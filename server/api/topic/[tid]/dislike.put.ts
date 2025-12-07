import { prisma } from '~~/prisma/prisma'
import { updateTopicDislikeSchema } from '~/validations/topic'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const input = await kunParsePutBody(event, updateTopicDislikeSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const topic = await prisma.topic.findUnique({
    where: { id: input.topicId, status: { not: 1 } },
    include: {
      dislike: {
        where: {
          user_id: uid
        }
      }
    }
  })
  if (!topic) {
    return kunError(event, '未找到该话题')
  }
  if (topic.user_id === uid) {
    return kunError(event, '您不能给自己点踩')
  }

  const isDisliked = topic.dislike.length > 0

  if (isDisliked) {
    await prisma.topic_dislike.delete({
      where: {
        topic_id_user_id: {
          user_id: uid,
          topic_id: input.topicId
        }
      }
    })
  } else {
    await prisma.topic_dislike.create({
      data: {
        user_id: uid,
        topic_id: input.topicId
      }
    })
  }

  return 'MOEMOE dislike topic successfully!'
})
