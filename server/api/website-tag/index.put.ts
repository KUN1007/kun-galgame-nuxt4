import { prisma } from '~~/prisma/prisma'
import { updateWebsiteTagSchema } from '~/validations/website'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限更新网站标签')
  }

  const input = await kunParsePutBody(event, updateWebsiteTagSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { tagId, ...rest } = input

  await prisma.galgame_website_tag.update({
    where: { id: tagId },
    data: rest
  })

  return 'Moemoe update website tag successfully!'
})
