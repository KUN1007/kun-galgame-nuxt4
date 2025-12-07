import { prisma } from '~~/prisma/prisma'
import { getGalgameDuplicateSchema } from '~/validations/galgame'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameDuplicateSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const galgame = await prisma.galgame.count({
    where: { vndb_id: input.vndbId }
  })
  if (galgame) {
    return kunError(event, '这个 Galgame 已经有人发布过了')
  }

  return 'The galgame is unique!'
})
