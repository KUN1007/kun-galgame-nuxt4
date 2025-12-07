import { prisma } from '~~/prisma/prisma'
import { deleteCommentSchema } from '~/validations/website'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const input = kunParseDeleteQuery(event, deleteCommentSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const comment = await prisma.galgame_website_comment.findUnique({
    where: { id: input.commentId }
  })
  if (!comment) {
    return kunError(event, '评论不存在', 404)
  }

  if (comment.user_id !== userInfo.uid && userInfo.role < 2) {
    return kunError(event, '您没有权限删除该评论', 403)
  }

  await prisma.galgame_website_comment.delete({
    where: { id: input.commentId }
  })

  return '评论删除成功'
})
