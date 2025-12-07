import { prisma } from '~~/prisma/prisma'
import { deleteCommentSchema } from '~/validations/topic'

export default defineEventHandler(async (event) => {
  const input = kunParseDeleteQuery(event, deleteCommentSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const comment = await prisma.topic_comment.findUnique({
    where: { id: input.commentId },
    include: {
      _count: {
        select: {
          like: true
        }
      }
    }
  })
  if (!comment) {
    return kunError(event, '未找到这个评论')
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  if (comment.user_id !== userInfo.uid && userInfo.role < 2) {
    return kunError(event, '您无权删除这个评论')
  }

  const moemoepointToDecreaseIfUserDelete = 3 * (comment._count.like + 1)
  const moemoepointToDecrease =
    userInfo.role < 2 ? moemoepointToDecreaseIfUserDelete : 3

  const user = await prisma.user.findUnique({
    where: { id: userInfo.uid },
    select: { moemoepoint: true }
  })
  if (!user) {
    return kunError(event, '未找到用户')
  }
  if (
    user.moemoepoint < moemoepointToDecreaseIfUserDelete &&
    userInfo.role < 2
  ) {
    return kunError(
      event,
      `您的萌萌点不足, 删除这个评论需要 ${moemoepointToDecreaseIfUserDelete} 萌萌点`
    )
  }

  return prisma.$transaction(async (prisma) => {
    await prisma.topic_comment.delete({
      where: { id: input.commentId }
    })

    await prisma.user.update({
      where: { id: comment.user_id },
      data: { moemoepoint: { increment: -moemoepointToDecrease } }
    })

    return 'MOEMOE delete comment successfully!'
  })
})
