import { prisma } from '~~/prisma/prisma'
import { updateReplySchema } from '~/validations/topic'
import { markdownToText } from '~~/server/utils/markdownToText'
import type { TopicReply } from '~/types/api/topic-reply'

export default defineEventHandler(async (event) => {
  const input = await kunParsePutBody(event, updateReplySchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const { replyId, targets, content } = input
  const validTargets = targets?.filter((item) => item.content?.trim())
  const uid = userInfo.uid

  const existingReply = await prisma.topic_reply.findUnique({
    where: { id: replyId },
    include: { target: true }
  })
  if (!existingReply) {
    return kunError(event, '回复不存在')
  }
  if (existingReply.user_id !== uid) {
    return kunError(event, '您没有权限修改此回复')
  }

  const originalTargetUserIds = new Set(
    existingReply.target.map((t) => t.target_reply_id)
  )

  return prisma.$transaction(async (prisma) => {
    if (targets) {
      await prisma.topic_reply_target.deleteMany({
        where: { reply_id: replyId }
      })
    }

    const result = await prisma.topic_reply.update({
      where: { id: replyId },
      data: {
        content,
        edited: new Date(),
        target: {
          create: validTargets?.map((target) => ({
            target_reply_id: target.targetReplyId,
            content: target.content
          }))
        }
      },
      include: {
        pinned: true,
        best_answer: true,
        user: {
          select: { id: true, name: true, avatar: true, moemoepoint: true }
        },
        target: {
          include: {
            target_reply: {
              select: {
                id: true,
                floor: true,
                content: true,
                user: {
                  select: { id: true, name: true, avatar: true }
                }
              }
            }
          }
        },
        like: { where: { user_id: uid } },
        dislike: { where: { user_id: uid } },
        _count: {
          select: {
            like: true,
            dislike: true,
            target_by: true
          }
        }
      }
    })

    const targetUsersMap = new Map<number, { user: KunUser; content: string }>()
    for (const newTarget of result.target) {
      if (
        newTarget.target_reply.user.id !== userInfo.uid &&
        !originalTargetUserIds.has(newTarget.target_reply_id)
      ) {
        targetUsersMap.set(newTarget.target_reply.user.id, {
          user: newTarget.target_reply.user,
          content: newTarget.content
        })
      }
    }

    for (const [, { user, content }] of targetUsersMap) {
      await prisma.user.update({
        where: { id: user.id },
        data: { moemoepoint: { increment: 1 } }
      })

      await createDedupMessage(
        prisma,
        userInfo.uid,
        user.id,
        'replied',
        markdownToText(content).slice(0, 233),
        result.topic_id
      )
    }

    const formattedTargets = await Promise.all(
      result.target.map(async (targetRelation) => {
        const targetReply = targetRelation.target_reply
        const originalContentText = markdownToText(targetReply.content)

        return {
          id: targetReply.id,
          floor: targetReply.floor,
          user: targetReply.user,
          contentPreview:
            originalContentText.slice(0, 150) +
            (originalContentText.length > 150 ? '...' : ''),
          replyContentMarkdown: targetRelation.content,
          replyContentHtml: await markdownToHtml(targetRelation.content)
        }
      })
    )

    const formattedReply: TopicReply = {
      id: result.id,
      topicId: result.topic_id,
      floor: result.floor,
      user: result.user as KunUser & { moemoepoint: number },
      edited: result.edited,
      contentMarkdown: result.content,
      contentHtml: await markdownToHtml(result.content),
      likeCount: result._count.like,
      isLiked: result.like.length > 0,
      dislikeCount: result._count.dislike,
      isDisliked: result.dislike.length > 0,
      targetByCount: result._count.target_by,
      comment: [],
      created: result.created,
      targets: formattedTargets,
      isPinned: !!result.pinned,
      isBestAnswer: !!result.best_answer
    }

    return formattedReply
  })
})
