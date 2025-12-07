import { prisma } from '~~/prisma/prisma'
import { createUpdateLogSchema } from '~/validations/update-log'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限创建网站更新日志')
  }

  const input = await kunParsePostBody(event, createUpdateLogSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  await prisma.update_log.create({
    data: input
  })

  return 'Moemoe create update log successfully!'
})
