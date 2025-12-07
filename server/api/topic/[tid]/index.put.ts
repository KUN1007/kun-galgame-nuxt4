import { prisma } from '~~/prisma/prisma'
import { updateTopicSchema } from '~/validations/topic'
import {
  TOPIC_SECTION_CONSUME_MOEMOEPOINTS,
  MOEMOEPOINT_COST_FOR_CONSUME_SECTION
} from '~/config/moemoepoint'

export default defineEventHandler(async (event) => {
  const input = await kunParsePutBody(event, updateTopicSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const oldTopic = await prisma.topic.findUnique({
    where: { id: input.topicId },
    include: {
      user: {
        select: {
          moemoepoint: true
        }
      },
      section: {
        select: {
          topic_section: {
            select: {
              name: true
            }
          }
        }
      }
    }
  })
  if (!oldTopic) {
    return kunError(event, '未找到此话题')
  }
  if (userInfo.uid !== oldTopic.user_id && userInfo.role < 2) {
    return kunError(event, '您没有权限更改此话题')
  }
  const oldSections = oldTopic.section.map((s) => s.topic_section.name)
  const oldTopicHasConsumeSection = TOPIC_SECTION_CONSUME_MOEMOEPOINTS.some(
    (item) => oldSections.includes(item as 'g-seeking')
  )
  const newTopicHasConsumeSection = TOPIC_SECTION_CONSUME_MOEMOEPOINTS.some(
    (item) => input.section.includes(item as 'g-seeking')
  )
  if (!oldTopicHasConsumeSection && newTopicHasConsumeSection) {
    if (
      oldTopic.user.moemoepoint < MOEMOEPOINT_COST_FOR_CONSUME_SECTION &&
      userInfo.role < 2
    ) {
      return kunError(
        event,
        '您的萌萌点不足, 无法更改话题的分类为求助或者寻求资源, 您可以通过发布 Galgame, 签到, 接受别人的赞赏, 等等来获取萌萌点'
      )
    }
  }

  const { topicId, section, ...topicData } = input

  return prisma.$transaction(async (prisma) => {
    await prisma.topic.update({
      where: { id: topicId },
      data: {
        ...topicData,
        edited: new Date(),
        status_update_time: new Date()
      }
    })

    const sections = await prisma.topic_section.findMany({
      where: { name: { in: section } },
      select: {
        id: true
      }
    })
    const newSectionIds = sections.map((s) => s.id)

    await prisma.topic_section_relation.deleteMany({
      where: {
        topic_id: topicId,
        topic_section_id: {
          notIn: newSectionIds
        }
      }
    })

    const dataToCreate = newSectionIds.map((sectionId) => ({
      topic_id: topicId,
      topic_section_id: sectionId
    }))

    await prisma.topic_section_relation.createMany({
      data: dataToCreate,
      skipDuplicates: true
    })

    if (!oldTopicHasConsumeSection && newTopicHasConsumeSection) {
      await prisma.user.update({
        where: { id: oldTopic.user_id },
        data: { moemoepoint: { increment: -10 } }
      })
    }

    return 'MOEMOE update topic successfully!'
  })
})
