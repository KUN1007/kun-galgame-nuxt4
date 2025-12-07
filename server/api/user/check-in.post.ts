import { randomNum } from '~/utils/random'
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
    return kunError(event, '未找到用户')
  }
  if (user.daily_check_in) {
    return kunError(event, '您今天已经签到过了')
  }

  const randomMoemoepoints = randomNum(0, 7)

  await prisma.user.update({
    where: { id: userInfo.uid },
    data: { moemoepoint: { increment: randomMoemoepoints }, daily_check_in: 1 }
  })

  return randomMoemoepoints
})
