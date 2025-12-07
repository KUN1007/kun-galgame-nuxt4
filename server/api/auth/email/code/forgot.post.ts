import { prisma } from '~~/prisma/prisma'
import { createSendForgotPasswordVerificationCodeSchema } from '~/validations/auth'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(
    event,
    createSendForgotPasswordVerificationCodeSchema
  )
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const existEmail = await prisma.user.findFirst({
    where: { email: { equals: input.email, mode: 'insensitive' } }
  })
  if (!existEmail) {
    return kunError(event, '邮箱不存在')
  }

  const result = await sendVerificationCodeEmail(event, input.email, 'forgot')
  if (typeof result === 'string') {
    return kunError(event, result)
  }

  return result.salt
})
