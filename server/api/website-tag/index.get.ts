import { prisma } from '~~/prisma/prisma'
import { getWebsiteTagSchema } from '~/validations/website'
import type { WebsiteTag } from '~/types/api/website'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getWebsiteTagSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const data = await prisma.galgame_website_tag.findMany({
    where: input.websiteId
      ? { website: { some: { galgame_website_id: input.websiteId } } }
      : {}
  })

  const tags: WebsiteTag[] = data.map((t) => ({
    id: t.id,
    name: t.name,
    level: t.level,
    label: t.label,
    description: t.description
  }))

  return tags
})
