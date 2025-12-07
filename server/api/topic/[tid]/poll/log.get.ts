import { prisma } from '~~/prisma/prisma'
import { getPollLogSchema } from '~/validations/topic-poll'
import { canUserViewPollResults } from './_canUserViewPollResults'
import type { TopicPollVoteLog } from '~/types/api/topic-poll'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)

  const input = kunParseGetQuery(event, getPollLogSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const poll = await prisma.topic_poll.findUnique({
    where: { id: input.poll_id },
    include: {
      option: {
        include: {
          vote: {
            where: {
              user_id: userInfo?.uid
            }
          }
        }
      }
    }
  })
  if (!poll) {
    return kunError(event, '未找到该投票')
  }

  const isUserVoted = poll.option.some((opt) => opt.vote.length > 0)

  const showResults = canUserViewPollResults(poll, userInfo, isUserVoted)

  if (!showResults || poll.is_anonymous) {
    return { logs: [], totalCount: 0 }
  }

  const { page, limit } = input
  const skip = (page - 1) * limit

  const [data, totalCount] = await Promise.all([
    prisma.topic_poll_vote.findMany({
      where: { poll_id: input.poll_id },
      skip,
      take: limit,
      include: {
        user: {
          select: {
            id: true,
            name: true,
            avatar: true
          }
        },
        option: {
          select: {
            text: true
          }
        }
      },
      orderBy: { created: 'desc' }
    }),
    prisma.topic_poll_vote.count({ where: { poll_id: input.poll_id } })
  ])

  const logs: TopicPollVoteLog[] = data.map((l) => ({
    id: l.id,
    created: l.created,
    user: l.user,
    option: l.option.text
  }))

  return { logs, totalCount }
})
