import { prisma } from '~~/prisma/prisma'
import { getToolsetPracticalitySchema } from '~/validations/toolset'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getToolsetPracticalitySchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)

  const [rates, me] = await Promise.all([
    prisma.galgame_toolset_practicality.findMany({
      where: {
        toolset_id: input.toolsetId
      },
      select: { rate: true }
    }),
    prisma.galgame_toolset_practicality.findFirst({
      where: {
        toolset_id: input.toolsetId,
        user_id: userInfo?.uid
      },
      select: { rate: true }
    })
  ])

  const counts: Record<number, number> = {}
  for (let i = 1; i <= 5; i++) counts[i] = 0
  rates.forEach((r) => {
    counts[r.rate] = (counts[r.rate] || 0) + 1
  })

  const total = rates.length
  const avg = total ? rates.reduce((a, b) => a + b.rate, 0) / total : 0

  return { counts, avg, mine: me ? me.rate : 0 }
})
