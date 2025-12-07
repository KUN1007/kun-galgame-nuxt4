import { prisma } from '~~/prisma/prisma'
import { markdownToText } from '~~/server/utils/markdownToText'
import { updateTopicUpvoteSchema } from '~/validations/topic'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const input = await kunParsePutBody(event, updateTopicUpvoteSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const topic = await prisma.topic.findUnique({
    where: { id: input.topicId, status: { not: 1 } }
  })
  if (!topic) {
    return kunError(event, '未找到该话题')
  }
  if (topic.user_id === uid) {
    return kunError(event, '您不能推自己的话题')
  }

  return prisma.$transaction(async (prisma) => {
    await prisma.topic_upvote.create({
      data: {
        user_id: uid,
        topic_id: input.topicId
      }
    })

    await prisma.topic.update({
      where: { id: input.topicId },
      data: { status_update_time: new Date(), upvote_time: new Date() }
    })

    await prisma.user.update({
      where: { id: topic.user_id },
      data: { moemoepoint: { increment: 3 } }
    })

    await prisma.user.update({
      where: { id: uid },
      data: { moemoepoint: { increment: -7 } }
    })

    await createDedupMessage(
      prisma,
      uid,
      topic.user_id,
      'upvoted',
      markdownToText(topic.content).slice(0, 233),
      topic.id
    )

    return 'MOEMOE upvoted topic successfully!'
  })
})
