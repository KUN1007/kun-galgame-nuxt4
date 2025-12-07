import { prisma } from '~~/prisma/prisma'
import { createPollSchema } from '~/validations/topic-poll'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }

  const input = await kunParsePostBody(event, createPollSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const topic = await prisma.topic.findUnique({
    where: { id: input.topic_id },
    include: {
      _count: {
        select: {
          poll: true
        }
      }
    }
  })
  if (!topic) {
    return kunError(event, '话题不存在')
  }
  if (topic.user_id !== userInfo.uid && userInfo.role < 2) {
    return kunError(event, '您没有权限在此话题下创建投票')
  }
  if (topic._count.poll >= 30) {
    return kunError(event, '单个话题下方最多有 30 个投票')
  }

  return prisma.$transaction(async (prisma) => {
    const { options, ...rest } = input
    const poll = await prisma.topic_poll.create({
      data: {
        ...rest,
        deadline: rest.deadline ? new Date(rest.deadline) : null,
        user_id: userInfo.uid
      },
      include: {
        user: {
          select: {
            id: true,
            name: true,
            avatar: true
          }
        }
      }
    })

    await prisma.topic_poll_option.createMany({
      data: options.map((opt) => ({
        text: opt.text,
        poll_id: poll.id
      }))
    })

    await prisma.topic.update({
      where: { id: input.topic_id },
      data: { status_update_time: new Date() }
    })
  })
})
