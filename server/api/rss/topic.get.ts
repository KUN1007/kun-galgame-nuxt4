import { prisma } from '~~/prisma/prisma'
import type { RSSTopic } from '~/types/api/rss'

export default defineEventHandler(async (event) => {
  const data = await prisma.topic.findMany({
    where: { status: { not: 1 }, is_nsfw: false },
    orderBy: {
      created: 'desc'
    },
    take: 10,
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

  const topics: RSSTopic[] = data.map((topic) => ({
    id: topic.id,
    name: topic.title,
    user: topic.user,
    description: topic.content.slice(0, 233),
    created: topic.created
  }))

  return topics
})
