import { prisma } from '~~/prisma/prisma'
import { updateToolsetCommentSchema } from '~/validations/toolset'

export default defineEventHandler(async (event) => {
  const input = await kunParsePutBody(event, updateToolsetCommentSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const comment = await prisma.galgame_toolset_comment.findUnique({
    where: { id: input.commentId },
    select: { user_id: true }
  })
  if (!comment) {
    return kunError(event, '评论不存在')
  }

  if (comment.user_id !== userInfo.uid) {
    return kunError(event, '您没有权限编辑该评论')
  }

  await prisma.galgame_toolset_comment.update({
    where: { id: input.commentId },
    data: { content: input.content, edited: new Date() }
  })

  return 'MOEMOE update toolset comment successfully!'
})
