import { prisma } from '~~/prisma/prisma'
import { getGalgameOfficialSchema } from '~/validations/galgame-official'
import type { GalgameOfficialItem } from '~/types/api/galgame-official'
import type { KunGalgameOfficialCategory } from '~/constants/galgameOfficial'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameOfficialSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { page, limit } = input
  const skip = (page - 1) * limit

  const [data, totalCount] = await Promise.all([
    prisma.galgame_official.findMany({
      skip,
      take: limit,
      include: {
        _count: {
          select: {
            galgame: true
          }
        },
        alias: true
      },
      orderBy: {
        galgame: {
          _count: 'desc'
        }
      }
    }),
    prisma.galgame_official.count()
  ])

  const officials: GalgameOfficialItem[] = data.map((o) => ({
    id: o.id,
    name: o.name,
    link: o.link,
    category: o.category as KunGalgameOfficialCategory,
    lang: o.lang,
    alias: o.alias.map((a) => a.name),
    galgameCount: o._count.galgame
  }))

  return { officials, totalCount }
})
