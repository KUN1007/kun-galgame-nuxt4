import { prisma } from '~~/prisma/prisma'
import { getReplySchema } from '~/validations/topic'
import { markdownToText } from '~~/server/utils/markdownToText'
import { Prisma } from '~~/prisma/generated/prisma/client'
import type { z } from 'zod'
import type { TopicReply } from '~/types/api/topic-reply'
import type { TopicComment } from '~/types/api/topic-comment'

const generateReplyInclude = (uid: number | undefined) => {
  return {
    user: {
      select: { id: true, name: true, avatar: true, moemoepoint: true }
    },
    target: {
      orderBy: {
        target_reply: { floor: 'asc' as 'asc' | 'desc' }
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
    dislike: { where: { user_id: uid } },
    comment: {
      orderBy: { created: 'asc' as 'asc' | 'desc' },
      include: {
        user: {
          select: { id: true, name: true, avatar: true }
        },
        target_user: {
          select: { id: true, name: true, avatar: true }
        },
        _count: {
          select: { like: true }
        },
        like: {
          where: { user_id: uid }
        }
      }
    }
  }
}

const prismaReplyWithDetails = Prisma.validator<Prisma.topic$replyArgs>()({
  // 1 is a user_id
  include: generateReplyInclude(1)
})
type PrismaReplyWithDetails = Prisma.topic_replyGetPayload<
  typeof prismaReplyWithDetails
>

const mapPrismaReplyToTopicReply = async (
  reply: PrismaReplyWithDetails
): Promise<TopicReply> => {
  const comments: TopicComment[] = reply.comment.map((comment) => ({
    id: comment.id,
    replyId: comment.topic_reply_id,
    topicId: comment.topic_id,
    user: comment.user,
    targetUser: comment.target_user,
    content: comment.content,
    isLiked: comment.like.length > 0,
    likeCount: comment._count.like,
    created: comment.created
  }))

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

  return {
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
    comment: comments,
    created: reply.created,
    targets: targets,
    isPinned: false, // default value
    isBestAnswer: false // default value
  }
}

const getTopicReplies = async (
  input: z.infer<typeof getReplySchema>,
  uid?: number
) => {
  const { topicId, page, limit, sortOrder } = input
  const skip = (page - 1) * limit

  const topic = await prisma.topic.findUnique({
    where: { id: topicId, status: { not: 1 } },
    select: { pinned_reply_id: true, best_answer_id: true }
  })

  if (!topic) {
    return []
  }

  // if pinned reply is the best answer
  const specialReplyIds = new Set<number>()
  if (topic.pinned_reply_id) {
    specialReplyIds.add(topic.pinned_reply_id)
  }
  if (topic.best_answer_id) {
    specialReplyIds.add(topic.best_answer_id)
  }
  const specialIdsArray = Array.from(specialReplyIds)

  let specialRepliesData: PrismaReplyWithDetails[] = []

  // only show best answer / pinned reply on first page
  if (page === 1 && specialIdsArray.length > 0) {
    specialRepliesData = await prisma.topic_reply.findMany({
      where: {
        id: {
          in: specialIdsArray
        }
      },
      include: generateReplyInclude(uid)
    })
  }

  const regularRepliesData = await prisma.topic_reply.findMany({
    skip,
    take: limit,
    where: {
      topic_id: topicId,
      // remove special replies in the common replies array
      id: {
        notIn: specialIdsArray
      }
    },
    orderBy: {
      floor: sortOrder
    },
    include: generateReplyInclude(uid)
  })

  const [specialReplies, regularReplies] = await Promise.all([
    Promise.all(specialRepliesData.map(mapPrismaReplyToTopicReply)),
    Promise.all(regularRepliesData.map(mapPrismaReplyToTopicReply))
  ])

  specialReplies.forEach((reply) => {
    if (reply.id === topic.pinned_reply_id) {
      reply.isPinned = true
    }
    if (reply.id === topic.best_answer_id) {
      reply.isBestAnswer = true
    }
  })

  specialReplies.sort((a, b) => {
    if (a.isPinned && !b.isPinned) return -1
    if (!a.isPinned && b.isPinned) return 1
    return 0
  })

  return [...specialReplies, ...regularReplies]
}

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getReplySchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  const result = await getTopicReplies(input, userInfo?.uid)

  return result
})
