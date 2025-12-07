import { prisma } from '~~/prisma/prisma'
import { markdownToText } from '~~/server/utils/markdownToText'
import { updateTopicBestAnswerSchema } from '~/validations/topic'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid
  const user = await prisma.user.findUnique({
    where: { id: uid },
    select: { name: true }
  })
  if (!user) {
    return kunError(event, '未找到该用户')
  }

  const input = await kunParsePutBody(event, updateTopicBestAnswerSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const topic = await prisma.topic.findUnique({
    where: { id: input.topicId, status: { not: 1 } },
    include: { user: { select: { name: true } } }
  })
  if (!topic) {
    return kunError(event, '未找到该话题')
  }
  if (topic.user_id !== uid && userInfo.role < 2) {
    return kunError(event, '您无权将该回复设置为最佳答案')
  }

  const reply = await prisma.topic_reply.findUnique({
    where: { id: input.replyId },
    include: {
      target: { select: { content: true } },
      best_answer: true
    }
  })
  if (!reply) {
    return kunError(event, '未找到该回复')
  }
  const targets = reply.target.map((t) => markdownToText(t.content))
  const replyContent = `${markdownToText(reply.content)}${targets.toString()}`

  return prisma.$transaction(async (prisma) => {
    if (reply.best_answer) {
      await prisma.topic.update({
        where: { id: reply.topic_id },
        data: {
          best_answer_id: null
        }
      })
    } else {
      await prisma.topic.update({
        where: {
          id: input.topicId
        },
        data: {
          best_answer_id: input.replyId,
          status_update_time: new Date()
        }
      })

      if (uid !== reply.user_id) {
        await createDedupMessage(
          prisma,
          uid,
          reply.user_id,
          'solution',
          replyContent.slice(0, 233),
          topic.id
        )
      }
    }

    await prisma.user.update({
      where: { id: reply.user_id },
      data: { moemoepoint: { increment: reply.best_answer ? -7 : 7 } }
    })

    return 'MOEMOE set best answer successfully!'
  })
})
