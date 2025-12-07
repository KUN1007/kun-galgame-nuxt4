import { prisma } from '~~/prisma/prisma'
import { getNSFWCookie } from '~~/server/utils/getNSFWCookie'
import { getTopicCategoryStats } from '~/validations/category'
import type {
  CategoryLatestTopicInfo,
  CategorySectionStats
} from '~/types/api/category'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getTopicCategoryStats)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const nsfw = getNSFWCookie(event)
  const isSFW = nsfw === 'sfw'

  const topics = await prisma.topic.findMany({
    where: {
      status: { not: 1 },
      ...(isSFW ? { is_nsfw: false } : {}),
      category: input.category
    },
    select: {
      id: true,
      title: true,
      view: true,
      created: true,

      section: {
        select: {
          topic_section: {
            select: {
              id: true,
              name: true
            }
          }
        }
      }
    },
    orderBy: {
      created: 'desc'
    }
  })

  const statsMap = new Map<number, CategorySectionStats>()

  for (const topic of topics) {
    for (const relation of topic.section) {
      const section = relation.topic_section
      if (!section) continue

      if (!statsMap.has(section.id)) {
        statsMap.set(section.id, {
          id: section.id,
          name: section.name,
          topicCount: 0,
          viewCount: 0,
          latestTopic: null
        })
      }

      const stats = statsMap.get(section.id)!

      stats.topicCount++
      stats.viewCount += topic.view

      if (stats.latestTopic === null) {
        stats.latestTopic = {
          id: topic.id,
          title: topic.title,
          created: topic.created
        } satisfies CategoryLatestTopicInfo
      }
    }
  }

  const result = Array.from(statsMap.values())

  return result
})
