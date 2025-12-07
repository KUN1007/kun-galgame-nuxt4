import { prisma } from '~~/prisma/prisma'
import { getGalgamePrDetailSchema } from '~/validations/galgame'
import type { GalgamePRDetails } from '~/types/api/galgame-pr'
import type { updateGalgameSchema } from '~/validations/galgame'
import type { z } from 'zod'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgamePrDetailSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const pr = await prisma.galgame_pr.findUnique({
    where: { id: input.galgamePrId },
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
  if (!pr) {
    return kunError(event, '未找到这个更新请求')
  }

  const details: GalgamePRDetails = {
    id: pr.id,
    galgameId: pr.galgame_id,
    index: pr.index,
    status: pr.status,
    completedTime: pr.completed_time,
    user: pr.user,
    created: pr.created,
    oldData: pr.old_data as z.infer<typeof updateGalgameSchema>,
    newData: pr.new_data as z.infer<typeof updateGalgameSchema>
  }

  return details
})
