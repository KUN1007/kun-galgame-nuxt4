import { prisma } from '~~/prisma/prisma'
import { deletePollSchema } from '~/validations/topic-poll'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }

  const input = kunParseGetQuery(event, deletePollSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const poll = await prisma.topic_poll.findUnique({
    where: { id: input.poll_id }
  })
  if (!poll) {
    return kunError(event, '未找到该投票')
  }
  if (poll.user_id !== userInfo.uid && userInfo.role < 2) {
    return kunError(event, '您无权删除该话题投票')
  }

  await prisma.topic_poll.delete({
    where: { id: input.poll_id }
  })

  return 'Moe moe delete topic poll successfully!'
})
