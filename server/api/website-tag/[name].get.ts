import { prisma } from '~~/prisma/prisma'
import { getWebsiteByTagSchema } from '~/validations/website'
import type { WebsiteTagDetail } from '~/types/api/website'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getWebsiteByTagSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const data = await prisma.galgame_website_tag.findUnique({
    where: { name: input.name },
    include: {
      website: {
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
      }
    }
  })
  if (!data) {
    return kunError(event, '未找到这个标签')
  }

  const websites: WebsiteTagDetail = {
    id: data.id,
    name: data.name,
    label: data.label,
    level: data.level,
    description: data.description,
    websiteCount: data.website.length,
    websites: data.website.map((w) => ({
      id: w.website.id,
      name: w.website.name,
      description: w.website.description,
      domain: w.website.url,
      ageLimit: w.website.age_limit,
      icon: w.website.icon,
      level: w.website.tag.reduce((acc, curr) => acc + curr.tag.level, 0),
      price: w.website.tag
        .map((t) => t.tag)
        .reduce((sum, tag) => sum + tag.level, 0),
      category: w.website.category.name
    })),
    created: data.created,
    updated: data.updated
  }

  return websites
})
