import { prisma } from '~~/prisma/prisma'
import { getAdminUserSearchSchema } from '~/validations/admin'
import type { AdminUser } from '~/types/api/admin'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getAdminUserSearchSchema)
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

  const data = await prisma.user.findMany({
    where: {
      name: { contains: input.q, mode: 'insensitive' }
    },
    take: 50
  })

  const users: AdminUser[] = data.map((user) => ({
    id: user.id,
    name: user.name,
    avatar: user.avatar,
    created: user.created,
    status: user.status
  }))

  return users
})
