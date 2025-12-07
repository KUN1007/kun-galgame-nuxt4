import { prisma } from '~~/prisma/prisma'
import { createSendResetEmailVerificationCodeSchema } from '~/validations/auth'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(
    event,
    createSendResetEmailVerificationCodeSchema
  )
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const existEmail = await prisma.user.findFirst({
    where: { email: { equals: input.email, mode: 'insensitive' } }
  })
  if (existEmail) {
    return kunError(event, '未找到该邮箱对应的用户')
  }

  const result = await sendVerificationCodeEmail(event, input.email, 'reset')
  if (typeof result === 'string') {
    return kunError(event, result)
  }

  return result.salt
})
