import { prisma } from '~~/prisma/prisma'
import { subMonths } from 'date-fns'
import { getNSFWCookie } from '~~/server/utils/getNSFWCookie'
import { getActivityTimelineData } from '~~/server/utils/activityTimeline'
import type { HomeGalgame, HomeTopic } from '~/types/api/home'

const GALGAME_LIMIT = 12
const TOPIC_LIMIT = 10
const ACTIVITY_LIMIT = 30

export default defineEventHandler(async (event) => {
  const nsfw = getNSFWCookie(event)
  const isSFW = nsfw === 'sfw'
  const contentLimit = isSFW ? 'sfw' : undefined

  const galgamesPromise = prisma.galgame
    .findMany({
      where: { status: { not: 1 }, content_limit: contentLimit },
      orderBy: { resource_update_time: 'desc' },
      skip: 0,
      take: GALGAME_LIMIT,
      include: {
        user: { select: { id: true, name: true, avatar: true } },
        _count: { select: { like: true } },
        resource: { select: { platform: true, language: true } }
      }
    })
    .then((data) => {
      const items: HomeGalgame[] = data.map((galgame) => {
        const platforms = [...new Set(galgame.resource.map((r) => r.platform))]
        const languages = [...new Set(galgame.resource.map((r) => r.language))]

        return {
          id: galgame.id,
          name: {
            'en-us': galgame.name_en_us,
            'ja-jp': galgame.name_ja_jp,
            'zh-cn': galgame.name_zh_cn,
            'zh-tw': galgame.name_zh_tw
          },
          banner: galgame.banner,
          user: galgame.user,
          contentLimit: galgame.content_limit,
          view: galgame.view,
          likeCount: galgame._count.like,
          resourceUpdateTime: galgame.resource_update_time,
          platform: platforms,
          language: languages
        }
      })
      return items
    })

  const topicsPromise = prisma.topic
    .findMany({
      skip: 0,
      take: TOPIC_LIMIT,
      where: {
        status: { not: 1 },
        ...(isSFW ? { is_nsfw: false } : {}),
        section: {
          none: {
            topic_section: {
              name: { in: ['g-seeking', 'g-other', 't-help'] }
            }
          }
        },
        OR: [
          { edited: { gte: subMonths(new Date(), 3) } },
          { edited: null, created: { gte: subMonths(new Date(), 3) } }
        ]
      },
      orderBy: { status_update_time: 'desc' },
      include: {
        best_answer: true,
        user: { select: { id: true, name: true, avatar: true } },
        section: { include: { topic_section: true } },
        _count: {
          select: { like: true, reply: true, comment: true, poll: true }
        }
      }
    })
    .then((data) => {
      const items: HomeTopic[] = data.map((topic) => ({
        id: topic.id,
        title: topic.title,
        view: topic.view,
        likeCount: topic._count.like,
        replyCount: topic._count.reply,
        commentCount: topic._count.comment,
        hasBestAnswer: !!topic.best_answer,
        isPollTopic: !!topic._count.poll,
        isNSFWTopic: topic.is_nsfw,
        tag: topic.tag,
        section: topic.section.map((s) => s.topic_section.name),
        user: topic.user,
        status: topic.status,
        upvoteTime: topic.upvote_time,
        statusUpdateTime: topic.status_update_time
      }))
      return items
    })

  const activitiesPromise = getActivityTimelineData({
    page: 1,
    limit: ACTIVITY_LIMIT,
    nsfw
  }).then((r) => r.items)

  const [galgames, topics, activities] = await Promise.all([
    galgamesPromise,
    topicsPromise,
    activitiesPromise
  ])

  return { galgames, topics, activities }
})
