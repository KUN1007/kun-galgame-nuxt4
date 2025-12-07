import { prisma } from '~~/prisma/prisma'
import { getUpdateLogSchema } from '~/validations/update-log'
import type { UpdateType, UpdateLog } from '~/types/api/update-log'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getUpdateLogSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { page, limit } = input

  const skip = (page - 1) * limit
  const totalCount = await prisma.update_log.count()

  const updateLogs = await prisma.update_log.findMany({
    skip,
    take: limit,
    orderBy: { created: 'desc' }
  })

  const updates: UpdateLog[] = updateLogs.map((log) => ({
    id: log.id,
    type: log.type as UpdateType,
    content: {
      'en-us': log.content_en_us,
      'ja-jp': log.content_ja_jp,
      'zh-cn': log.content_zh_cn,
      'zh-tw': log.content_zh_tw
    },
    version: log.version,
    created: log.created
  }))

  return { updates, totalCount }
})
