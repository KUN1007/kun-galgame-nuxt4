import { prisma } from '~~/prisma/prisma'
import { createWebsiteTagSchema } from '~/validations/website'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限创建网站标签')
  }

  const input = await kunParsePostBody(event, createWebsiteTagSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  await prisma.galgame_website_tag.create({
    data: input
  })

  return 'Moemoe create website tag successfully!'
})
