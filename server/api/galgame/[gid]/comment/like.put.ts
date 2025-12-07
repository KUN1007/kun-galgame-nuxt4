import { prisma } from '~~/prisma/prisma'
import { updateGalgameCommentLikeSchema } from '~/validations/galgame'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, updateGalgameCommentLikeSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const comment = await prisma.galgame_comment.findUnique({
    where: { id: input.galgameCommentId },
    include: {
      like: {
        where: {
          user_id: uid
        }
      }
    }
  })
  if (!comment) {
    return kunError(event, '未找到该评论')
  }
  if (comment.user_id === uid) {
    return kunError(event, '您不能给自己点赞')
  }
  if (comment.like.length > 0) {
    return kunError(event, '您已经点赞过了')
  }

  return prisma.$transaction(async (prisma) => {
    await prisma.galgame_comment_like.create({
      data: {
        user_id: uid,
        galgame_comment_id: input.galgameCommentId
      }
    })

    await prisma.user.update({
      where: { id: comment.user_id },
      data: { moemoepoint: { increment: 1 } }
    })

    if (comment.target_user_id) {
      await createMessage(
        prisma,
        uid,
        comment.user_id,
        'liked',
        comment.content.slice(0, 233),
        0,
        comment.galgame_id
      )
    }

    return 'MOEMOE like galgame comment successfully!'
  })
})
