import { prisma } from '~~/prisma/prisma'
import { getAdminUserSchema } from '~/validations/admin'
import type { AdminUser } from '~/types/api/admin'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getAdminUserSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 1) {
    return kunError(event, '您没有权限查看此页面')
  }

  const { page, limit } = input
  const skip = (page - 1) * limit

  const [data, totalCount] = await Promise.all([
    prisma.user.findMany({
      skip,
      take: limit,
      orderBy: { created: 'desc' }
    }),
    prisma.user.count()
  ])

  const users: AdminUser[] = data.map((user) => ({
    id: user.id,
    name: user.name,
    avatar: user.avatar,
    created: user.created,
    status: user.status
  }))

  return { users, totalCount }
})
