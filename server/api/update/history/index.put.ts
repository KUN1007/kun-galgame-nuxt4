import { prisma } from '~~/prisma/prisma'
import { updateUpdateLogSchema } from '~/validations/update-log'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限更新更新日志')
  }

  const input = await kunParsePutBody(event, updateUpdateLogSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { updateLogId, ...rest } = input

  await prisma.update_log.update({
    where: { id: updateLogId },
    data: rest
  })

  return 'Moemoe update update log successfully!'
})
