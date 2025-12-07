import { prisma } from '~~/prisma/prisma'
import { updateTopicHideStatusSchema } from '~/validations/topic'

export default defineEventHandler(async (event) => {
  const input = await kunParsePutBody(event, updateTopicHideStatusSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const topic = await prisma.topic.findUnique({
    where: { id: input.topicId }
  })
  if (!topic) {
    return kunError(event, '未找到该话题')
  }
  if (topic.user_id !== uid && userInfo.role < 2) {
    return kunError(event, '您没有权限隐藏该话题')
  }

  if (topic.status === 1) {
    await prisma.topic.update({
      where: { id: input.topicId },
      data: { status: 0 }
    })
  } else {
    await prisma.topic.update({
      where: { id: input.topicId },
      data: { status: 1 }
    })
  }

  return 'Moemoe update topic status successfully!'
})
