import { prisma } from '~~/prisma/prisma'
import { getReplyDetailSchema } from '~/validations/topic'
import { markdownToText } from '~~/server/utils/markdownToText'
import type { TopicReplyDetail } from '~/types/api/topic-reply'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getReplyDetailSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  const uid = userInfo?.uid

  const reply = await prisma.topic_reply.findUnique({
    where: {
      id: input.replyId
    },
    include: {
      pinned: true,
      best_answer: true,
      user: {
        select: { id: true, name: true, avatar: true, moemoepoint: true }
      },
      target: {
        orderBy: {
          target_reply: { floor: 'asc' }
        },
        select: {
          content: true,
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

      _count: {
        select: { like: true, dislike: true, comment: true, target_by: true }
      },
      like: { where: { user_id: uid } },
      dislike: { where: { user_id: uid } }
    }
  })
  if (!reply) {
    return kunError(event, '未找到该回复')
  }

  const targets = await Promise.all(
    reply.target.map(async (targetRelation) => {
      const targetReply = targetRelation.target_reply
      const contentText = markdownToText(targetReply.content)
      return {
        id: targetReply.id,
        floor: targetReply.floor,
        user: targetReply.user,
        contentPreview:
          contentText.slice(0, 150) + (contentText.length > 150 ? '...' : ''),
        replyContentMarkdown: targetRelation.content,
        replyContentHtml: await markdownToHtml(targetRelation.content)
      }
    })
  )

  const result: TopicReplyDetail = {
    id: reply.id,
    topicId: reply.topic_id,
    floor: reply.floor,
    user: reply.user,
    edited: reply.edited,
    contentMarkdown: reply.content,
    contentHtml: await markdownToHtml(reply.content),
    likeCount: reply._count.like,
    isLiked: reply.like.length > 0,
    dislikeCount: reply._count.dislike,
    isDisliked: reply.dislike.length > 0,
    targetByCount: reply._count.target_by,
    created: reply.created,
    targets: targets,
    isPinned: !!reply.pinned,
    isBestAnswer: !!reply.best_answer
  }

  return result
})
