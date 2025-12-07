import { prisma } from '~~/prisma/prisma'
import { createReportSchema } from '~/validations/report'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, createReportSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  await prisma.report.create({
    data: { reason: input.reason, type: input.type }
  })

  return 'Moe Moe submit report successfully!'
})
