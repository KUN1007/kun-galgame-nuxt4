import { prisma } from '~~/prisma/prisma'
import { createReplySchema } from '~/validations/topic'
import { markdownToText } from '~~/server/utils/markdownToText'
import { createMessage } from '~~/server/utils/message'
import type { TopicReply } from '~/types/api/topic-reply'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, createReplySchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 401)
  }

  const { topicId, content, targets } = input
  const validTargets = targets.filter((item) => item.content?.trim())

  const topic = await prisma.topic.findUnique({
    where: { id: topicId }
  })
  if (!topic) {
    return kunError(event, '未找到该话题')
  }

  const lastReply = await prisma.topic_reply.findFirst({
    where: { topic_id: topicId },
    orderBy: { floor: 'desc' },
    select: { floor: true }
  })
  const newFloor = (lastReply?.floor || 0) + 1

  return prisma.$transaction(async (prisma) => {
    await prisma.topic.update({
      where: { id: input.topicId },
      data: { status_update_time: new Date() }
    })

    const newReply = await prisma.topic_reply.create({
      data: {
        user_id: userInfo.uid,
        topic_id: topicId,
        floor: newFloor,
        content,

        target: {
          create: validTargets.map((target) => ({
            target_reply_id: target.targetReplyId,
            content: target.content
          }))
        }
      },

      include: {
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
        }
      }
    })

    const targetUsersMap = new Map<number, { user: KunUser; content: string }>()
    newReply.target.map((t) => {
      if (t.target_reply.user.id !== userInfo.uid) {
        targetUsersMap.set(t.target_reply.user.id, {
          user: t.target_reply.user,
          content: t.content
        })
      }
    })

    for (const [, { user, content }] of targetUsersMap) {
      await prisma.user.update({
        where: { id: user.id },
        data: { moemoepoint: { increment: 1 } }
      })

      await createMessage(
        prisma,
        userInfo.uid,
        user.id,
        'replied',
        markdownToText(content).slice(0, 233),
        topicId
      )
    }

    if (content.trim()) {
      await prisma.user.update({
        where: { id: topic.user_id },
        data: { moemoepoint: { increment: 1 } }
      })

      await createMessage(
        prisma,
        userInfo.uid,
        topic.user_id,
        'replied',
        markdownToText(content).slice(0, 233),
        topicId
      )
    }

    const formattedTargets = await Promise.all(
      newReply.target.map(async (targetRelation) => {
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
      id: newReply.id,
      topicId: newReply.topic_id,
      floor: newReply.floor,
      user: newReply.user,
      edited: newReply.edited,
      contentMarkdown: newReply.content,
      contentHtml: await markdownToHtml(newReply.content),
      likeCount: 0,
      isLiked: false,
      dislikeCount: 0,
      isDisliked: false,
      targetByCount: 0,
      created: newReply.created,
      comment: [],
      targets: formattedTargets,
      isBestAnswer: false,
      isPinned: false
    }

    return formattedReply
  })
})
