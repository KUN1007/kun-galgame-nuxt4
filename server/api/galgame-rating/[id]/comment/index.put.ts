import { prisma } from '~~/prisma/prisma'
import { updateGalgameRatingCommentSchema } from '~/validations/galgame-rating'
import type { GalgameRatingComment } from '~/types/api/galgame-rating'

export default defineEventHandler(async (event) => {
  const input = await kunParsePutBody(event, updateGalgameRatingCommentSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '请先登录', 205)
  }

  const comment = await prisma.galgame_rating_comment.findUnique({
    where: { id: input.galgameRatingCommentId },
    include: {
      galgame_rating: {
        select: {
          galgame_id: true,
          user_id: true,
          galgame: { select: { user_id: true } }
        }
      },
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
  if (!comment) {
    return kunError(event, '评论不存在', 404)
  }

  const user = await prisma.user.findUnique({ where: { id: userInfo.uid } })
  if (!user) {
    return kunError(event, '用户不存在')
  }
  if (comment.user_id !== user.id) {
    return kunError(event, '您无权限编辑该评论', 403)
  }

  await prisma.galgame_rating_comment.update({
    where: { id: input.galgameRatingCommentId },
    data: { content: input.content }
  })

  const newComment: GalgameRatingComment = {
    id: comment.id,
    content: input.content,
    user: comment.user,
    targetUser: comment.target_user,
    created: comment.created,
    updated: comment.updated
  }

  return newComment
})
