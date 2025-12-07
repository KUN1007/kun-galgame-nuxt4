import { hash } from 'bcrypt'
import { prisma } from '~~/prisma/prisma'
import { createSendResetPasswordByEmailVerificationCodeSchema } from '~/validations/auth'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(
    event,
    createSendResetPasswordByEmailVerificationCodeSchema
  )
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { codeSalt, email, code, newPassword } = input

  const codeKey = `${codeSalt}:${email}`
  const isCodeValid = await verifyVerificationCode(codeKey, code)
  if (!isCodeValid) {
    return 10103
  }

  await useStorage('redis').removeItem(codeKey)
  const hashedPassword = await hash(newPassword, 7)

  const user = await prisma.user.update({
    where: { email },
    data: { password: hashedPassword }
  })
  if (!user) {
    return kunError(event, '为找到该用户')
  }

  return 'MOEMOE reset password by email successfully!'
})
