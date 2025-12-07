import { prisma } from '~~/prisma/prisma'
import { getToolsetSchema } from '~/validations/toolset'
import type { ToolsetCard } from '~/types/api/toolset'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getToolsetSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const {
    page,
    limit,
    type,
    language,
    platform,
    version,
    sortField,
    sortOrder
  } = input
  const skip = (page - 1) * limit

  const where = {
    status: { not: 1 },
    ...(type !== 'all' ? { type } : {}),
    ...(language !== 'all' ? { language } : {}),
    ...(platform !== 'all' ? { platform } : {}),
    ...(version !== 'all' ? { version } : {})
  }

  const [totalCount, data] = await Promise.all([
    prisma.galgame_toolset.count({ where }),
    prisma.galgame_toolset.findMany({
      where,
      orderBy: { [sortField]: sortOrder },
      skip,
      take: limit,
      include: {
        user: { select: { id: true, name: true, avatar: true } },
        comment: { select: { id: true } },
        practicality: { select: { rate: true } },
        resource: { select: { download: true } }
      }
    })
  ])

  const items: ToolsetCard[] = data.map((t) => {
    const practicalityAvg = t.practicality.length
      ? t.practicality.reduce((a, b) => a + b.rate, 0) / t.practicality.length
      : null
    const totalDownload = t.resource.reduce(
      (sum, r) => sum + (r.download ?? 0),
      0
    )

    return {
      id: t.id,
      name: t.name,
      user: t.user,
      type: t.type,
      platform: t.platform,
      language: t.language,
      version: t.version,
      download: totalDownload,
      view: t.view,
      commentCount: t.comment.length,
      resource_update_time: t.resource_update_time,
      practicalityAvg
    }
  })

  return { items, totalCount }
})
