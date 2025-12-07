import { prisma } from '~~/prisma/prisma'
import { getGalgamePrSchema } from '~/validations/galgame'
import type { GalgamePR } from '~/types/api/galgame-pr'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgamePrSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { galgameId, page, limit } = input

  const offset = (page - 1) * limit
  const totalCount = await prisma.galgame_pr.count({
    where: { galgame_id: galgameId }
  })

  const data = await prisma.galgame_pr.findMany({
    take: Number(limit),
    skip: offset,
    orderBy: { created: 'desc' },
    where: { galgame_id: galgameId },
    include: {
      user: {
        select: {
          id: true,
          name: true,
          avatar: true
        }
      }
    }
  })

  const prs: GalgamePR[] = data.map((pr) => ({
    id: pr.id,
    galgameId: pr.galgame_id,
    index: pr.index,
    status: pr.status,
    completedTime: pr.completed_time,
    user: pr.user,
    created: pr.created
  }))

  return { prs, totalCount }
})
