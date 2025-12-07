import { prisma } from '~~/prisma/prisma'
import { getNSFWCookie } from '~~/server/utils/getNSFWCookie'
import { getSectionSchema } from '~/validations/section'
import type { z } from 'zod'
import type { SectionTopic } from '~/types/api/section'

const getSectionTopic = async (
  input: z.infer<typeof getSectionSchema>,
  isSFW: boolean
) => {
  const { section, page, limit, sortOrder } = input
  const skip = (page - 1) * limit

  const totalCount = await prisma.topic.count({
    where: {
      status: {
        not: 1
      },
      ...(isSFW ? { is_nsfw: false } : {}),
      section: {
        some: {
          topic_section: {
            name: section
          }
        }
      }
    }
  })

  const data = await prisma.topic.findMany({
    skip,
    take: limit,
    orderBy: { created: sortOrder },
    where: {
      status: { not: 1 },
      ...(isSFW ? { is_nsfw: false } : {}),
      section: {
        some: {
          topic_section: {
            name: section
          }
        }
      }
    },
    include: {
      best_answer: true,
      user: {
        select: {
          id: true,
          name: true,
          avatar: true
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
      },
      _count: {
        select: {
          like: true,
          reply: true,
          comment: true,
          poll: true
        }
      }
    }
  })

  const topics: SectionTopic[] = data.map((topic) => ({
    id: topic.id,
    title: topic.title,
    content: topic.content.slice(0, 233),
    section: topic.section.map((s) => s.topic_section.name),
    tag: topic.tag,
    view: topic.view,
    like: topic._count.like,
    reply: topic._count.reply,
    user: topic.user,
    hasBestAnswer: !!topic.best_answer,
    isPollTopic: !!topic._count.poll,
    isNSFWTopic: topic.is_nsfw,
    created: topic.created
  }))

  return { topics, totalCount }
}

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getSectionSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const nsfw = getNSFWCookie(event)
  const isSFW = nsfw === 'sfw'

  const res = await getSectionTopic(input, isSFW)

  return res
})
