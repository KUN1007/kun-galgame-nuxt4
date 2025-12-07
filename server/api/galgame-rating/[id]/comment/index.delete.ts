import { prisma } from '~~/prisma/prisma'
import { deleteGalgameRatingCommentSchema } from '~/validations/galgame-rating'

export default defineEventHandler(async (event) => {
  const input = kunParseDeleteQuery(event, deleteGalgameRatingCommentSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
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
      }
    }
  })
  if (!comment) {
    return kunError(event, '评论不存在', 404)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const user = await prisma.user.findUnique({ where: { id: userInfo.uid } })
  if (!user) {
    return kunError(event, '未找到该用户')
  }

  if (
    comment.user_id !== user.id &&
    comment.galgame_rating.galgame.user_id !== user.id &&
    user.role < 2
  ) {
    return kunError(event, '您没有删除该评论的权限', 403)
  }

  await prisma.galgame_rating_comment.delete({
    where: { id: input.galgameRatingCommentId }
  })

  return 'MOEMOE delete galgame rating comment successfully!'
})
