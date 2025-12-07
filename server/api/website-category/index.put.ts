import { prisma } from '~~/prisma/prisma'
import { updateWebsiteCategorySchema } from '~/validations/website'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限更新这个分类')
  }

  const input = await kunParsePutBody(event, updateWebsiteCategorySchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { categoryId, ...dataToUpdate } = input

  const updatedCategory = await prisma.galgame_website_category.update({
    where: { id: categoryId },
    data: dataToUpdate
  })
  return updatedCategory
})
