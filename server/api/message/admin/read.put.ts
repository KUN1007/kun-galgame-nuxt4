import { prisma } from '~~/prisma/prisma'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  await prisma.system_message.updateMany({
    data: { status: 'read' }
  })

  return 'MOEMOE read all messages successfully!'
})
