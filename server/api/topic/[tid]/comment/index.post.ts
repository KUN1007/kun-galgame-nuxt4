import { prisma } from '~~/prisma/prisma'
import { createCommentSchema } from '~/validations/topic'
import type { TopicComment } from '~/types/api/topic-comment'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, createCommentSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const { topicId, replyId, targetUserId, content } = input
  const uid = userInfo.uid

  return prisma.$transaction(async (prisma) => {
    const newComment = await prisma.topic_comment.create({
      data: {
        topic_id: topicId,
        topic_reply_id: replyId,
        user_id: uid,
        target_user_id: targetUserId,
        content
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

    if (uid !== targetUserId) {
      await createMessage(
        prisma,
        uid,
        targetUserId,
        'commented',
        newComment.content,
        topicId,
        undefined
      )
      await prisma.user.update({
        where: { id: targetUserId },
        data: { moemoepoint: { increment: 1 } }
      })
    }

    const responseData: TopicComment = {
      id: newComment.id,
      replyId: newComment.topic_reply_id,
      topicId: newComment.topic_id,
      user: newComment.user,
      targetUser: newComment.target_user,
      content: newComment.content,
      isLiked: false,
      likeCount: 0,
      created: newComment.created
    }

    return responseData
  })
})
