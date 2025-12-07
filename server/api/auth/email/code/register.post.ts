import { prisma } from '~~/prisma/prisma'
import { KUN_ALLOW_REGISTER_EMAIL } from '~/config/email-whitelist'
import {
  ADMIN_DELETE_EMAIL_CACHE_KEY,
  ADMIN_DELETE_IP_CACHE_KEY,
  KUN_FORUM_DISABLE_REGISTER_KEY
} from '~/config/admin'
import { createSendRegisterVerificationCodeSchema } from '~/validations/auth'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(
    event,
    createSendRegisterVerificationCodeSchema
  )
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const { name, email } = input

  const isDisableRegister = await useStorage('redis').getItem(
    KUN_FORUM_DISABLE_REGISTER_KEY
  )
  if (isDisableRegister) {
    return kunError(
      event,
      '由于网站近日遭受大量攻击，当前时间段暂时不可注册，请明天下午再来，一定要来哦 (-B)'
    )
  }

  const isDeletedUserEmail = await useStorage('redis').getItem(
    `${ADMIN_DELETE_EMAIL_CACHE_KEY}:${email}`
  )
  if (isDeletedUserEmail) {
    return kunError(event, '您已经被永久封禁')
  }
  const authUserIp = getRemoteIp(event)
  const isDeletedUserIp = await useStorage('redis').getItem(
    `${ADMIN_DELETE_IP_CACHE_KEY}:${authUserIp}`
  )
  if (isDeletedUserIp) {
    return kunError(event, '您已经被永久封禁')
  }

  const emailDomain = email.split('@')[1]
  const isEmailAllowed = KUN_ALLOW_REGISTER_EMAIL.some(
    (whitelistedDomain) => whitelistedDomain === emailDomain
  )
  if (!isEmailAllowed) {
    return kunError(event, '您的邮箱地址暂时不受支持, 请换一个邮箱试试')
  }

  const existUser = await prisma.user.findFirst({
    where: { name: { equals: name.toLowerCase(), mode: 'insensitive' } }
  })
  if (existUser) {
    return kunError(event, '您的用户名已经被使用了, 请换一个试试')
  }

  const existEmail = await prisma.user.findFirst({
    where: { email }
  })
  if (existEmail) {
    return kunError(event, '您的邮箱已经被使用了, 请换一个试试')
  }

  const result = await sendVerificationCodeEmail(event, email, 'register')
  if (typeof result === 'string') {
    return kunError(event, result)
  }

  return result.salt
})
