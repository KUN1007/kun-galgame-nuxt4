import { prisma } from '~~/prisma/prisma'
import { getToolsetDetailSchema } from '~/validations/toolset'
import type { ToolsetDetail } from '~/types/api/toolset'
import type { ToolsetComment } from '~/types/api/toolset-comment'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getToolsetDetailSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const toolset = await prisma.$transaction(async (prisma) => {
    const [res] = await Promise.all([
      prisma.galgame_toolset.findUnique({
        where: { id: input.toolsetId },
        include: {
          user: {
            select: {
              id: true,
              name: true,
              avatar: true
            }
          },
          alias: {
            select: {
              name: true
            }
          },
          practicality: {
            select: {
              rate: true
            }
          },
          resource: {
            select: {
              id: true,
              type: true,
              size: true,
              download: true,
              status: true
            }
          }
        }
      }),
      prisma.galgame_toolset.update({
        where: { id: input.toolsetId },
        data: { view: { increment: 1 } }
      })
    ])

    return res
  })
  if (!toolset) {
    return kunError(event, '未找到该工具资源')
  }

  const practicalityAvg = toolset.practicality.length
    ? toolset.practicality.reduce((a, b) => a + b.rate, 0) /
      toolset.practicality.length
    : null

  const totalDownload = toolset.resource.reduce(
    (sum, r) => sum + (r.download ?? 0),
    0
  )

  // rating distribution
  const ratingCounts: Record<number, number> = { 1: 0, 2: 0, 3: 0, 4: 0, 5: 0 }
  for (const p of toolset.practicality) {
    const k = Math.min(5, Math.max(1, p.rate)) as 1 | 2 | 3 | 4 | 5
    ratingCounts[k] = (ratingCounts[k] || 0) + 1
  }

  // JSON-LD (latest 5)
  const [commentCount, commentPreviewRaw] = await Promise.all([
    prisma.galgame_toolset_comment.count({
      where: { toolset_id: input.toolsetId }
    }),
    prisma.galgame_toolset_comment.findMany({
      where: { toolset_id: input.toolsetId },
      orderBy: { created: 'desc' },
      take: 5,
      include: {
        user: { select: { id: true, name: true, avatar: true } },
        parent: {
          include: { user: { select: { id: true, name: true, avatar: true } } }
        }
      }
    })
  ])
  const commentPreview: ToolsetComment[] = commentPreviewRaw.map((c) => ({
    id: c.id,
    toolsetId: c.toolset_id,
    created: c.created,
    content: c.content,
    edited: c.edited,
    parentId: c.parent_id,
    userId: c.user_id,
    reply: [],
    user: c.user,
    targetUser: c.parent?.user || null
  }))

  const detail: ToolsetDetail = {
    ...toolset,
    contentMarkdown: toolset.description,
    contentHtml: await markdownToHtml(toolset.description),
    practicalityAvg,
    download: totalDownload,
    aliases: toolset.alias.map((a) => a.name),
    practicalityCount: toolset.practicality.length,
    resource_update_time: toolset.resource_update_time,
    ratingCounts,
    commentCount,
    commentPreview
  }

  return detail
})
