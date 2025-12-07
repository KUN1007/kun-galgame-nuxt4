import { prisma } from '~~/prisma/prisma'
import { deleteWebsiteSchema } from '~/validations/website'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限删除这个网站')
  }

  const input = kunParseDeleteQuery(event, deleteWebsiteSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  await prisma.galgame_website.delete({
    where: { id: input.websiteId }
  })
  return 'Moemoe delete website successfully!'
})
