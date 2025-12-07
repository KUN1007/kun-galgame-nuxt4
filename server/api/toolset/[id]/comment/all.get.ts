import { prisma } from '~~/prisma/prisma'
import { getToolsetCommentSchema } from '~/validations/toolset'
import type { ToolsetComment } from '~/types/api/toolset-comment'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getToolsetCommentSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { toolsetId, page, limit, sortOrder } = input
  const skip = (page - 1) * limit

  const [data, totalCount] = await Promise.all([
    prisma.galgame_toolset_comment.findMany({
      take: limit,
      skip,
      where: { toolset_id: toolsetId },
      orderBy: { created: sortOrder },
      include: {
        user: { select: { id: true, name: true, avatar: true } },
        parent: {
          include: { user: { select: { id: true, name: true, avatar: true } } }
        }
      }
    }),
    prisma.galgame_toolset_comment.count({ where: { toolset_id: toolsetId } })
  ])

  const commentData: ToolsetComment[] = data.map((c) => ({
    id: c.id,
    toolsetId: c.toolset_id,
    content: c.content,
    created: c.created,
    edited: c.edited,
    parentId: c.parent_id,
    userId: c.user_id,
    reply: [],
    user: c.user,
    targetUser: c.parent?.user || null
  }))

  return { commentData, totalCount }
})
