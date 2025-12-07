import { prisma } from '~~/prisma/prisma'
import { getGalgameRatingDetailSchema } from '~/validations/galgame-rating'
import type { GalgameRatingDetails } from '~/types/api/galgame-rating'
import type {
  GalgameSeriesSample,
  GalgameSeries
} from '~/types/api/galgame-series'
import type { KunGalgameOfficialCategory } from '~/constants/galgameOfficial'

const userSelect = {
  id: true,
  name: true,
  avatar: true
} as const

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameRatingDetailSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)

  const data = await prisma.galgame_rating.findUnique({
    where: { id: input.galgameRatingId },
    include: {
      like: { include: { user: { select: userSelect } } },
      user: { select: userSelect },
      comment: {
        include: {
          user: { select: userSelect },
          target_user: { select: userSelect }
        }
      },
      galgame: {
        include: {
          rating: { select: { overall: true } },
          official: {
            select: {
              official: {
                include: {
                  _count: {
                    select: {
                      galgame: true
                    }
                  },
                  alias: true
                }
              }
            }
          },
          series: {
            include: {
              _count: { select: { galgame: true } },
              galgame: {
                select: {
                  content_limit: true,
                  name_en_us: true,
                  name_ja_jp: true,
                  name_zh_cn: true,
                  name_zh_tw: true,
                  banner: true
                },
                take: 5,
                orderBy: { created: 'asc' }
              }
            }
          }
        }
      }
    }
  })
  if (!data) {
    return kunError(event, '评分不存在')
  }

  // do not need to await, avoid blocking main thread
  prisma.galgame_rating
    .update({
      where: { id: input.galgameRatingId },
      data: { view: { increment: 1 } }
    })
    .catch(console.error)

  const likedUsers = data.like.map((l) => l.user)
  const isLiked = !!userInfo && likedUsers.some((u) => u.id === userInfo.uid)

  const sampleGalgame: GalgameSeriesSample[] =
    data.galgame.series?.galgame.map((g) => ({
      banner: g.banner,
      name: {
        'en-us': g.name_en_us,
        'ja-jp': g.name_ja_jp,
        'zh-cn': g.name_zh_cn,
        'zh-tw': g.name_zh_tw
      }
    })) ?? []

  const galgameSeries: GalgameSeries | null = data.galgame.series
    ? {
        id: data.galgame.series.id,
        name: data.galgame.series.name,
        description: data.galgame.series.description,
        isNSFW: false,
        sampleGalgame,
        galgameCount: data.galgame.series._count.galgame,
        created: data.galgame.series.created,
        updated: data.galgame.series.updated
      }
    : null

  const details: GalgameRatingDetails = {
    id: data.id,
    user: data.user,
    recommend: data.recommend,
    overall: data.overall,
    view: data.view + 1,
    galgameType: data.galgame_type,
    play_status: data.play_status,
    short_summary: data.short_summary,
    spoiler_level: data.spoiler_level,
    art: data.art,
    story: data.story,
    music: data.music,
    character: data.character,
    route: data.route,
    system: data.system,
    voice: data.voice,
    replay_value: data.replay_value,
    likeCount: likedUsers.length,
    isLiked,
    created: data.created,
    updated: data.updated,
    likedUsers,
    comments: data.comment.map((c) => ({
      id: c.id,
      content: c.content,
      user: c.user,
      targetUser: c.target_user,
      created: c.created,
      updated: c.updated
    })),
    galgame: {
      id: data.galgame.id,
      contentLimit: data.galgame.content_limit,
      name: {
        'en-us': data.galgame.name_en_us,
        'ja-jp': data.galgame.name_ja_jp,
        'zh-cn': data.galgame.name_zh_cn,
        'zh-tw': data.galgame.name_zh_tw
      },
      official: data.galgame.official.map((o) => ({
        id: o.official.id,
        name: o.official.name,
        link: o.official.link,
        category: o.official.category as KunGalgameOfficialCategory,
        lang: o.official.lang,
        alias: o.official.alias.map((a) => a.name),
        galgameCount: o.official._count.galgame
      })),
      ageLimit: data.galgame.age_limit,
      originalLanguage: data.galgame.original_language,
      banner: data.galgame.banner,
      rating: data.galgame.rating.reduce((sum, r) => sum + (r.overall ?? 0), 0),
      ratingCount: data.galgame.rating.length
    },
    galgameSeries
  }

  return details
})
