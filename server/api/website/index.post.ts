import { prisma } from '~~/prisma/prisma'
import { createWebsiteSchema } from '~/validations/website'
import { getDomain } from '~/utils/getDomain'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo || userInfo.role <= 2) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限添加网站')
  }

  const input = await kunParsePostBody(event, createWebsiteSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { tag_ids, url, ...rest } = input

  return prisma.$transaction(async (prisma) => {
    const website = await prisma.galgame_website.create({
      data: { ...rest, url: getDomain(url) }
    })

    const dataToCreate = tag_ids.map((sectionId) => ({
      galgame_website_id: website.id,
      galgame_website_tag_id: sectionId
    }))
    await prisma.galgame_website_tag_relation.createMany({
      data: dataToCreate,
      skipDuplicates: true
    })

    return 'Moemoe create website success'
  })
})
