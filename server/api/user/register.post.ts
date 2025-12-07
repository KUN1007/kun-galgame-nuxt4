import { hash } from 'bcrypt'
import { prisma } from '~~/prisma/prisma'
import { userRegisterSchema } from '~/validations/user'
import type { AuthLoginResponseData } from '~/types/api/user'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, userRegisterSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { codeSalt, name, email, password, code } = input

  const ip = getRemoteIp(event)
  const registerCD = await useStorage('redis').getItem(
    `login:register:cd:${name}`
  )
  if (registerCD) {
    return kunError(event, '注册冷却中, 请等待 60 秒后重试')
  } else {
    useStorage('redis').setItem(`login:register:cd:${ip}`, ip, { ttl: 60 })
  }

  const codeKey = `${codeSalt}:${email}`
  const isCodeValid = await verifyVerificationCode(codeKey, code)
  if (!isCodeValid) {
    return kunError(event, '非法的邮箱验证码')
  }
  await useStorage('redis').removeItem(codeKey)

  const usernameCount = await prisma.user.count({
    where: { name: { equals: name, mode: 'insensitive' } }
  })
  if (usernameCount) {
    return kunError(event, '您的用户名已经被使用, 请更换')
  }

  const emailCount = await prisma.user.count({
    where: { email }
  })
  if (emailCount) {
    return kunError(event, '您的邮箱已经被使用, 请更换')
  }

  const hashedPassword = await hash(password, 7)
  const user = await prisma.user.create({
    data: {
      name,
      email,
      password: hashedPassword,
      ip: ip.toString()
    }
  })

  const { refreshToken } = await createTokens(user.id, user.name, user.role)
  deleteCookie(event, 'kungalgame-is-navigate-to-login')
  setCookie(event, 'kungalgame-moemoe-refresh-token', refreshToken, {
    httpOnly: true,
    maxAge: 30 * 24 * 60 * 60 * 1000
  })

  const userInfo: AuthLoginResponseData = {
    id: user.id,
    name: user.name,
    avatar: user.avatar,
    avatarMin: user.avatar,
    moemoepoint: user.moemoepoint,
    role: user.role,
    isCheckIn: !!user.daily_check_in,
    dailyToolsetUploadCount: user.daily_toolset_upload_count
  }

  return userInfo
})
