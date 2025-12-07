import { prisma } from '~~/prisma/prisma'
import { getGalgameOfficialBySearchSchema } from '~/validations/galgame-official'
import type { GalgameOfficialItem } from '~/types/api/galgame-official'
import type { KunGalgameOfficialCategory } from '~/constants/galgameOfficial'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameOfficialBySearchSchema)
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

  const officials = await prisma.galgame_official.findMany({
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
      },
      alias: true
    },
    take: 50
  })

  const result: GalgameOfficialItem[] = officials.map((o) => ({
    id: o.id,
    name: o.name,
    link: o.link,
    category: o.category as KunGalgameOfficialCategory,
    lang: o.lang,
    alias: o.alias.map((a) => a.name),
    galgameCount: o._count.galgame
  }))

  return result
})
