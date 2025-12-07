import { prisma } from '~~/prisma/prisma'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const user = await prisma.user.findUnique({
    where: { id: userInfo.uid }
  })
  if (!user) {
    return kunError(event, '未找到该用户')
  }

  const email = user.email
  const atIndex = email.indexOf('@')
  const localPart = email.slice(0, atIndex)
  const domain = email.slice(atIndex)

  const maskedLocalPart = localPart.slice(0, 2) + '~~~~~~~'

  return maskedLocalPart + domain
})
