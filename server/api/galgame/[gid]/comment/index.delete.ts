import { prisma } from '~~/prisma/prisma'
import { deleteGalgameCommentSchema } from '~/validations/galgame'

export default defineEventHandler(async (event) => {
  const input = kunParseDeleteQuery(event, deleteGalgameCommentSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const comment = await prisma.galgame_comment.findUnique({
    where: { id: input.galgameCommentId },
    include: {
      _count: { select: { like: true } },
      galgame: {
        select: {
          user_id: true
        }
      }
    }
  })
  if (!comment) {
    return
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const user = await prisma.user.findUnique({
    where: { id: userInfo.uid }
  })
  if (!user) {
    return kunError(event, '未找到用户')
  }

  if (
    comment.user_id !== user.id &&
    comment.galgame.user_id !== user.id &&
    user.role < 2
  ) {
    return kunError(event, '您没有权限删除该评论')
  }

  return prisma.$transaction(async (prisma) => {
    if (comment.target_user_id && comment.user_id !== comment.target_user_id) {
      await prisma.user.update({
        where: { id: comment.target_user_id },
        data: { moemoepoint: { increment: -1 } }
      })
    }

    await prisma.user.update({
      where: { id: comment.user_id },
      data: { moemoepoint: { increment: -comment._count.like } }
    })

    await prisma.galgame_comment.delete({
      where: { id: input.galgameCommentId }
    })

    return 'MOEMOE delete visualnovel comment successfully!'
  })
})
