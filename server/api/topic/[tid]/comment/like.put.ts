import { prisma } from '~~/prisma/prisma'
import { updateCommentLikeSchema } from '~/validations/topic'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const input = await kunParsePutBody(event, updateCommentLikeSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const comment = await prisma.topic_comment.findUnique({
    where: { id: input.commentId },
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
    return kunError(event, '您不能给自己的评论点赞')
  }

  const isLikedComment = comment.like.length > 0

  return prisma.$transaction(async (prisma) => {
    if (isLikedComment) {
      await prisma.topic_comment_like.delete({
        where: {
          topic_comment_id_user_id: {
            user_id: uid,
            topic_comment_id: input.commentId
          }
        }
      })
    } else {
      await prisma.topic_comment_like.create({
        data: {
          user_id: uid,
          topic_comment_id: input.commentId
        }
      })
    }

    await prisma.user.update({
      where: { id: comment.user_id },
      data: { moemoepoint: { increment: isLikedComment ? -1 : 1 } }
    })

    await createDedupMessage(
      prisma,
      uid,
      comment.user_id,
      'liked',
      comment.content.slice(0, 233),
      comment.topic_id,
      undefined
    )

    return 'MOEMOE like comment successfully!'
  })
})
