import { prisma } from '~~/prisma/prisma'
import { createToolsetCommentSchema } from '~/validations/toolset'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, createToolsetCommentSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  return prisma.$transaction(async (prisma) => {
    const created = await prisma.galgame_toolset_comment.create({
      data: {
        content: input.content,
        user_id: uid,
        toolset_id: input.toolsetId,
        parent_id: input.parentId || null
      },
      include: {
        user: { select: { id: true, name: true, avatar: true } }
      }
    })

    if (input.parentId) {
      const parentComment = await prisma.galgame_toolset_comment.findUnique({
        where: { id: input.parentId },
        select: { user_id: true }
      })

      if (parentComment && parentComment.user_id !== uid) {
        await createMessage(
          prisma,
          uid,
          parentComment.user_id,
          'commented',
          input.content.slice(0, 233),
          undefined,
          undefined,
          undefined,
          input.toolsetId
        )
      }
    }

    return {
      id: created.id,
      content: created.content,
      parentId: created.parent_id,
      userId: created.user_id,
      toolsetId: created.toolset_id,
      created: created.created,
      edited: created.edited,
      reply: [],
      user: created.user
    }
  })
})
