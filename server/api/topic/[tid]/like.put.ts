import { prisma } from '~~/prisma/prisma'
import { updateTopicLikeSchema } from '~/validations/topic'
import { markdownToText } from '~~/server/utils/markdownToText'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const input = await kunParsePutBody(event, updateTopicLikeSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const topic = await prisma.topic.findUnique({
    where: { id: input.topicId, status: { not: 1 } },
    include: {
      like: {
        where: {
          user_id: uid
        }
      }
    }
  })
  if (!topic) {
    return kunError(event, '未找到该话题')
  }
  if (topic.user_id === uid) {
    return kunError(event, '您不能给自己点赞')
  }

  const isLiked = topic.like.length > 0

  return prisma.$transaction(async (prisma) => {
    if (isLiked) {
      await prisma.topic_like.delete({
        where: {
          topic_id_user_id: {
            user_id: uid,
            topic_id: input.topicId
          }
        }
      })
    } else {
      await prisma.topic_like.create({
        data: {
          user_id: uid,
          topic_id: input.topicId
        }
      })
    }

    await prisma.user.update({
      where: { id: topic.user_id },
      data: { moemoepoint: { increment: isLiked ? -1 : 1 } }
    })

    await createDedupMessage(
      prisma,
      uid,
      topic.user_id,
      'liked',
      markdownToText(topic.content).slice(0, 233),
      topic.id
    )

    return 'MOEMOE like topic successfully!'
  })
})
