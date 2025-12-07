import { prisma } from '~~/prisma/prisma'
import { getGalgameHistorySchema } from '~/validations/galgame'
import type {
  GalgameHistoryAction,
  GalgameHistoryType,
  GalgameHistory
} from '~/types/api/galgame-history'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameHistorySchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { galgameId, page, limit } = input

  const skip = (page - 1) * limit
  const totalCount = await prisma.galgame_history.count({
    where: { galgame_id: galgameId }
  })

  const data = await prisma.galgame_history.findMany({
    take: limit,
    skip,
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

  const historyData: GalgameHistory[] = data.map((history) => ({
    id: history.id,
    action: history.action as GalgameHistoryAction,
    type: history.type as GalgameHistoryType,
    content: history.content,
    user: history.user,
    created: history.created
  }))

  return { historyData, totalCount }
})
