import { subDays } from 'date-fns'
import { prisma } from '~~/prisma/prisma'
import { createTopicSchema } from '~/validations/topic'
import {
  TOPIC_SECTION_CONSUME_MOEMOEPOINTS,
  MOEMOEPOINT_COST_FOR_CONSUME_SECTION
} from '~/config/moemoepoint'

export default defineEventHandler(async (event) => {
  const input = await kunParsePutBody(event, createTopicSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const user = await prisma.user.findUnique({
    where: { id: userInfo.uid },
    include: {
      topic: {
        where: {
          created: {
            gte: subDays(new Date(), 1)
          }
        }
      }
    }
  })
  if (!user) {
    return kunError(event, '未找到该用户')
  }
  if (user.moemoepoint / 10 + 1 < user.topic.length) {
    return kunError(
      event,
      '您今日发布的话题已经达到上限, 您每天可以发布您的 (萌萌点 / 10) + 1 个话题'
    )
  }
  const hasConsumeSection = TOPIC_SECTION_CONSUME_MOEMOEPOINTS.some((item) =>
    input.section.includes(item as 'g-seeking')
  )
  if (hasConsumeSection) {
    if (user.moemoepoint < MOEMOEPOINT_COST_FOR_CONSUME_SECTION) {
      return kunError(
        event,
        '您的萌萌点不足, 无法发布求助类或者寻求资源类的话题, 您可以通过发布 Galgame, 签到, 接受别人的赞赏, 等等来获取萌萌点'
      )
    }
  }

  const { section, ...topicData } = input

  return prisma.$transaction(async (prisma) => {
    const newTopic = await prisma.topic.create({
      data: {
        ...topicData,
        user_id: userInfo.uid
      }
    })

    const sections = await prisma.topic_section.findMany({
      where: { name: { in: section } },
      select: {
        id: true
      }
    })
    const newSectionIds = sections.map((s) => s.id)
    const dataToCreate = newSectionIds.map((sectionId) => ({
      topic_id: newTopic.id,
      topic_section_id: sectionId
    }))
    await prisma.topic_section_relation.createMany({
      data: dataToCreate,
      skipDuplicates: true
    })

    await prisma.user.update({
      where: { id: userInfo.uid },
      data: { moemoepoint: { increment: hasConsumeSection ? -10 : 3 } }
    })

    return newTopic.id
  })
})
