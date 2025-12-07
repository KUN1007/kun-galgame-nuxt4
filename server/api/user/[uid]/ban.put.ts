import { prisma } from '~~/prisma/prisma'
import { updateBanUserSchema } from '~/validations/user'

export default defineEventHandler(async (event) => {
  const input = kunParseDeleteQuery(event, updateBanUserSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role < 2) {
    return kunError(event, '您没有权限操作用户')
  }

  const targetUser = await prisma.user.findUnique({
    where: {
      id: input.userId
    }
  })
  if (!targetUser) {
    return kunError(event, '用户不存在')
  }
  if (targetUser.role > 1) {
    return kunError(event, '不能操作一个管理员')
  }

  await prisma.user.update({
    where: { id: input.userId },
    data: { status: targetUser.status ? 0 : 1 }
  })

  if (!targetUser.status) {
    await useStorage('redis').del(`refreshToken:${input.userId}`)
  }

  return 'Moemoe update user status successfully!'
})
