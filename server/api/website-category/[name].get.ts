import { prisma } from '~~/prisma/prisma'
import { getWebsiteByTagSchema } from '~/validations/website'
import type { WebsiteCategoryDetail } from '~/types/api/website'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getWebsiteByTagSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const data = await prisma.galgame_website_category.findUnique({
    where: { name: input.name },
    include: {
      website: {
        include: {
          category: true,
          tag: {
            include: {
              tag: true
            }
          }
        }
      }
    }
  })
  if (!data) {
    return kunError(event, '未找到这个分类')
  }

  const websites: WebsiteCategoryDetail = {
    id: data.id,
    name: data.name,
    label: data.label,
    description: data.description,
    websiteCount: data.website.length,
    websites: data.website.map((w) => ({
      id: w.id,
      name: w.name,
      description: w.description,
      domain: w.url,
      ageLimit: w.age_limit,
      icon: w.icon,
      level: w.tag.reduce((acc, curr) => acc + curr.tag.level, 0),
      price: w.tag.map((t) => t.tag).reduce((sum, tag) => sum + tag.level, 0),
      category: w.category.name
    })),
    created: data.created,
    updated: data.updated
  }

  return websites
})
