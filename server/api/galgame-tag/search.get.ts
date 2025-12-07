import { prisma } from '~~/prisma/prisma'
import { getGalgameTagBySearchSchema } from '~/validations/galgame-tag'
import type { GalgameTagItem } from '~/types/api/galgame-tag'
import type { KunGalgameTagCategory } from '~/constants/galgameTag'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameTagBySearchSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const searchTerms = input.q
    .split(',')
    .map((term) => term.trim())
    .filter(Boolean)
  if (searchTerms.length === 0) {
    return []
  }

  const tags = await prisma.galgame_tag.findMany({
    where: {
      AND: searchTerms.map((st) => ({
        OR: [
          { name: { contains: st, mode: 'insensitive' } },
          {
            alias: {
              some: {
                name: { contains: st, mode: 'insensitive' }
              }
            }
          }
        ]
      }))
    },
    include: {
      _count: {
        select: {
          galgame: true
        }
      }
    },
    take: 50
  })

  const result: GalgameTagItem[] = tags.map((tag) => ({
    id: tag.id,
    name: tag.name,
    category: tag.category as KunGalgameTagCategory,
    galgameCount: tag._count.galgame
  }))

  return result
})
