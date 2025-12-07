import { prisma } from '~~/prisma/prisma'
import { createGalgameRatingCommentSchema } from '~/validations/galgame-rating'
import type { GalgameRatingComment } from '~/types/api/galgame-rating'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, createGalgameRatingCommentSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const rating = await prisma.galgame_rating.findUnique({
    where: { id: input.galgameRatingId },
    select: { galgame_id: true }
  })
  if (!rating) {
    return kunError(event, '未找到这个评分')
  }

  const { galgameRatingId, targetUserId, content } = input

  return prisma.$transaction(async (prisma) => {
    const res = await prisma.galgame_rating_comment.create({
      data: {
        galgame_rating_id: galgameRatingId,
        target_user_id: targetUserId,
        content,
        user_id: userInfo.uid
      },
      include: {
        user: {
          select: {
            id: true,
            name: true,
            avatar: true
          }
        },
        target_user: {
          select: {
            id: true,
            name: true,
            avatar: true
          }
        }
      }
    })

    if (userInfo.uid !== targetUserId) {
      await createMessage(
        prisma,
        userInfo.uid,
        targetUserId,
        'commented',
        content.slice(0, 233),
        undefined,
        rating.galgame_id
      )
    }

    const newComment: GalgameRatingComment = {
      id: res.id,
      content: res.content,
      user: res.user,
      targetUser: res.target_user,
      created: res.created,
      updated: res.updated
    }

    return newComment
  })
})
