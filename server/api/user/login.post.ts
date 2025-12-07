import { compare } from 'bcrypt'
import { prisma } from '~~/prisma/prisma'
import { userLoginSchema } from '~/validations/user'
import type { AuthLoginResponseData } from '~/types/api/user'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, userLoginSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { name, password } = input

  const ip = getRemoteIp(event)
  const loginCD = await useStorage('redis').getItem(`login:login:cd:${ip}`)
  if (loginCD) {
    return kunError(event, '登录冷却中, 请 17 秒后重试')
  } else {
    useStorage('redis').setItem(`login:login:cd:${ip}`, ip, { ttl: 17 })
  }

  const user = await prisma.user.findFirst({
    where: {
      OR: [{ name }, { email: name }]
    }
  })
  if (!user) {
    return kunError(event, '未找到用户')
  }
  if (user.status) {
    return kunError(event, '该用户已被封禁')
  }

  const isCorrectPassword = await compare(password, user.password)
  if (!isCorrectPassword) {
    return kunError(event, '用户密码错误')
  }

  const { refreshToken } = await createTokens(user.id, user.name, user.role)
  deleteCookie(event, 'kungalgame-is-navigate-to-login')

  const isProd = process.env.NODE_ENV === 'production'
  setCookie(event, 'kungalgame-moemoe-refresh-token', refreshToken, {
    httpOnly: true,
    sameSite: 'strict',
    secure: isProd,
    path: '/',
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
