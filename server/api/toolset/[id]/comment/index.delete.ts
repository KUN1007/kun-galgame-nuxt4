import { prisma } from '~~/prisma/prisma'
import { deleteToolsetCommentSchema } from '~/validations/toolset'

export default defineEventHandler(async (event) => {
  const input = kunParseDeleteQuery(event, deleteToolsetCommentSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const comment = await prisma.galgame_toolset_comment.findUnique({
    where: { id: input.commentId },
    include: { toolset: { select: { user_id: true } } }
  })
  if (!comment) {
    return kunError(event, '评论不存在')
  }

  const canDelete =
    comment.user_id === userInfo.uid ||
    comment.toolset.user_id === userInfo.uid ||
    userInfo.role >= 2
  if (!canDelete) {
    return kunError(event, '您没有权限删除该评论')
  }

  await prisma.galgame_toolset_comment.delete({
    where: { id: input.commentId }
  })

  return 'MOEMOE delete toolset comment successfully!'
})
