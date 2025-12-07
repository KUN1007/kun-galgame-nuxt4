import { prisma } from '~~/prisma/prisma'
import { updateTopicFavoriteSchema } from '~/validations/topic'
import { markdownToText } from '~~/server/utils/markdownToText'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const input = await kunParsePutBody(event, updateTopicFavoriteSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const topic = await prisma.topic.findUnique({
    where: { id: input.topicId, status: { not: 1 } },
    include: {
      favorite: {
        where: {
          user_id: uid
        }
      }
    }
  })
  if (!topic) {
    return kunError(event, '未找到该话题')
  }

  const isFavorited = topic.favorite.length > 0

  return prisma.$transaction(async (prisma) => {
    if (isFavorited) {
      await prisma.topic_favorite.delete({
        where: {
          topic_id_user_id: {
            user_id: uid,
            topic_id: input.topicId
          }
        }
      })
    } else {
      await prisma.topic_favorite.create({
        data: {
          user_id: uid,
          topic_id: input.topicId
        }
      })
    }

    if (topic.user_id !== uid) {
      await prisma.user.update({
        where: { id: topic.user_id },
        data: { moemoepoint: { increment: isFavorited ? -1 : 1 } }
      })

      await createDedupMessage(
        prisma,
        uid,
        topic.user_id,
        'favorite',
        markdownToText(topic.content).slice(0, 233),
        topic.id
      )
    }

    return 'MOEMOE favorite topic successfully!'
  })
})
