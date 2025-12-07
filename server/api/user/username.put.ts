import { prisma } from '~~/prisma/prisma'
import { updateUsernameSchema } from '~/validations/user'

export default defineEventHandler(async (event) => {
  const input = await kunParsePutBody(event, updateUsernameSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
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
  if (user.moemoepoint < 17) {
    return kunError(event, '更改用户名需要 17 萌萌点, 您的萌萌点不足')
  }

  const duplicatedNumber = await prisma.user.count({
    where: { name: { equals: input.username, mode: 'insensitive' } }
  })
  if (duplicatedNumber) {
    return kunError(event, '您的用户名已经被使用, 请换一个')
  }

  await prisma.user.update({
    where: { id: userInfo.uid },
    data: { name: input.username, moemoepoint: { increment: -17 } }
  })

  return 'Moe Moe'
})
