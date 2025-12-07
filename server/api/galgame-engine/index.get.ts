import { prisma } from '~~/prisma/prisma'
import type { GalgameEngineItem } from '~/types/api/galgame-engine'

export default defineEventHandler(async (event) => {
  const data = await prisma.galgame_engine.findMany({
    include: {
      _count: {
        select: {
          galgame: true
        }
      }
    },
    orderBy: {
      galgame: {
        _count: 'desc'
      }
    }
  })

  const engines: GalgameEngineItem[] = data.map((o) => ({
    id: o.id,
    name: o.name,
    alias: o.alias,
    galgameCount: o._count.galgame
  }))

  return engines
})
