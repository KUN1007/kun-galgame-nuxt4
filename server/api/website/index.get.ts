import { prisma } from '~~/prisma/prisma'
import type { WebsiteCard } from '~/types/api/website'

export default defineEventHandler(async (event) => {
  const nsfw = getNSFWCookie(event)
  const ageLimit = nsfw === 'nsfw' ? undefined : 'all'

  const data = await prisma.galgame_website.findMany({
    where: { age_limit: ageLimit },
    orderBy: {
      created: 'desc'
    },
    include: {
      category: true,
      tag: {
        include: {
          tag: true
        }
      }
    }
  })

  const websites: WebsiteCard[] = data.map((w) => ({
    id: w.id,
    name: w.name,
    description: w.description,
    domain: w.url,
    ageLimit: w.age_limit,
    icon: w.icon,
    level: w.tag.reduce((acc, curr) => acc + curr.tag.level, 0),
    price: w.tag.map((t) => t.tag).reduce((sum, tag) => sum + tag.level, 0),
    category: w.category.name
  }))

  return websites
})
