import { prisma } from '~~/prisma/prisma'
import { deleteWebsiteTagSchema } from '~/validations/website'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限删除这个网站标签')
  }

  const input = kunParseDeleteQuery(event, deleteWebsiteTagSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  await prisma.galgame_website_tag.delete({
    where: { id: input.tagId }
  })

  return 'Moemoe delete website tag successfully!'
})
