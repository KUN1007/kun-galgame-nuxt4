import { prisma } from '~~/prisma/prisma'
import { createCommentSchema } from '~/validations/website'
import type { WebsiteComment } from '~/types/api/website'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const input = await kunParsePostBody(event, createCommentSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  return prisma.$transaction(async (prisma) => {
    const newComment = await prisma.galgame_website_comment.create({
      data: {
        content: input.content,
        user_id: uid,
        website_id: input.websiteId,
        parent_id: input.parentId
      },
      include: {
        user: { select: { id: true, name: true, avatar: true } }
      }
    })

    if (input.parentId) {
      const parentComment = await prisma.galgame_website_comment.findUnique({
        where: { id: input.parentId },
        include: { website: { select: { url: true } } }
      })

      if (parentComment!.user_id !== uid) {
        await createMessage(
          prisma,
          uid,
          parentComment!.user_id,
          'commented',
          input.content.slice(0, 233),
          undefined,
          undefined,
          parentComment!.website.url
        )
      }
    }

    const comment: WebsiteComment = {
      id: newComment.id,
      content: newComment.content,
      parentId: newComment.parent_id,
      userId: newComment.user_id,
      websiteId: newComment.website_id,
      created: newComment.created,
      edited: newComment.edited,
      reply: [],
      user: newComment.user
    }

    return comment
  })
})
