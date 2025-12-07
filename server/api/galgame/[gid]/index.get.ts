import { prisma } from '~~/prisma/prisma'
import { getGalgameDetailSchema } from '~/validations/galgame'
import type { GalgameDetail } from '~/types/api/galgame'
import type {
  GalgameSeries,
  GalgameSeriesSample
} from '~/types/api/galgame-series'
import type { KunGalgameTagCategory } from '~/constants/galgameTag'
import type { KunGalgameOfficialCategory } from '~/constants/galgameOfficial'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameDetailSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const galgameId = input.galgameId

  const galgameStatus = await prisma.galgame.findUnique({
    where: { id: galgameId },
    select: { status: true }
  })
  if (!galgameStatus) {
    return kunError(event, '未找到这个 Galgame')
  }
  if (galgameStatus.status === 1) {
    return 'banned'
  }

  const userInfo = await getCookieTokenInfo(event)
  const userId = userInfo?.uid

  const [galgame] = await Promise.all([
    prisma.galgame.findUnique({
      where: { id: galgameId },
      include: {
        like: { where: { user_id: userId } },
        favorite: { where: { user_id: userId } },
        _count: {
          select: { like: true, favorite: true }
        },
        user: {
          select: { id: true, name: true, avatar: true }
        },
        contributor: {
          select: {
            user: {
              select: { id: true, name: true, avatar: true }
            }
          }
        },
        alias: { select: { name: true } },
        series: {
          include: {
            _count: {
              select: {
                galgame: true
              }
            },
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
              orderBy: {
                created: 'asc'
              }
            }
          }
        },
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
        engine: {
          select: {
            engine: {
              include: {
                _count: {
                  select: {
                    galgame: true
                  }
                }
              }
            }
          }
        },
        tag: {
          select: {
            spoiler_level: true,
            tag: {
              include: {
                _count: {
                  select: {
                    galgame: true
                  }
                }
              }
            }
          }
        },
        resource: {
          select: { type: true, language: true, platform: true }
        },
        rating: {
          include: {
            user: { select: { id: true, name: true, avatar: true } },
            like: { where: { user_id: userId } },
            _count: { select: { like: true } }
          }
        }
      }
    }),

    prisma.galgame.update({
      where: { id: galgameId },
      data: { view: { increment: 1 } }
    })
  ])
  if (!galgame) {
    return kunError(event, '未找到 Galgame')
  }

  const [introHtmlEn, introHtmlJa, introHtmlZhCn, introHtmlZhTw] =
    await Promise.all([
      markdownToHtml(galgame.intro_en_us),
      markdownToHtml(galgame.intro_ja_jp),
      markdownToHtml(galgame.intro_zh_cn),
      markdownToHtml(galgame.intro_zh_tw)
    ])

  const sampleGalgame: GalgameSeriesSample[] = galgame.series
    ? galgame.series.galgame.map((g) => ({
        banner: g.banner,
        name: {
          'en-us': g.name_en_us,
          'ja-jp': g.name_ja_jp,
          'zh-cn': g.name_zh_cn,
          'zh-tw': g.name_zh_tw
        }
      }))
    : []
  const isSeriesNSFW = galgame.series
    ? galgame.series.galgame.some((g) => g.content_limit === 'nsfw')
    : false
  const galgameSeries: GalgameSeries | null = galgame.series
    ? {
        id: galgame.series.id,
        name: galgame.series.name,
        description: galgame.series.description,
        isNSFW: isSeriesNSFW,
        sampleGalgame,
        galgameCount: galgame.series._count.galgame,
        created: galgame.series.created,
        updated: galgame.series.updated
      }
    : null

  const data: GalgameDetail = {
    id: galgame.id,
    vndbId: galgame.vndb_id,
    user: galgame.user,
    contentLimit: galgame.content_limit,
    name: {
      'en-us': galgame.name_en_us,
      'ja-jp': galgame.name_ja_jp,
      'zh-cn': galgame.name_zh_cn,
      'zh-tw': galgame.name_zh_tw
    },
    banner: galgame.banner,
    introduction: {
      'en-us': introHtmlEn,
      'ja-jp': introHtmlJa,
      'zh-cn': introHtmlZhCn,
      'zh-tw': introHtmlZhTw
    },
    markdown: {
      'en-us': galgame.intro_en_us,
      'ja-jp': galgame.intro_ja_jp,
      'zh-cn': galgame.intro_zh_cn,
      'zh-tw': galgame.intro_zh_tw
    },
    resourceUpdateTime: galgame.resource_update_time,
    view: galgame.view,
    originalLanguage: galgame.original_language,
    ageLimit: galgame.age_limit as 'all' | 'r18',
    type: [...new Set(galgame.resource.map((r) => r.type))],
    platform: [...new Set(galgame.resource.map((r) => r.platform))],
    language: [...new Set(galgame.resource.map((r) => r.language))],
    contributor: galgame.contributor.map((c) => c.user),
    likeCount: galgame._count.like,
    isLiked: galgame.like.length > 0,
    favoriteCount: galgame._count.favorite,
    isFavorited: galgame.favorite.length > 0,
    alias: galgame.alias.map((a) => a.name),
    series: galgameSeries,
    engine: galgame.engine.map((o) => ({
      id: o.engine.id,
      name: o.engine.name,
      alias: o.engine.alias,
      galgameCount: o.engine._count.galgame
    })),
    official: galgame.official.map((o) => ({
      id: o.official.id,
      name: o.official.name,
      link: o.official.link,
      category: o.official.category as KunGalgameOfficialCategory,
      lang: o.official.lang,
      alias: o.official.alias.map((a) => a.name),
      galgameCount: o.official._count.galgame
    })),
    tag: galgame.tag.map((tag) => ({
      id: tag.tag.id,
      name: tag.tag.name,
      category: tag.tag.category as KunGalgameTagCategory,
      galgameCount: tag.tag._count.galgame,
      spoilerLevel: tag.spoiler_level
    })),
    ratings: galgame.rating.map((r) => ({
      ...r,
      galgameId: r.galgame_id,
      galgameType: r.galgame_type,
      likeCount: r._count.like,
      isLiked: r.like.length > 0,
      galgame: {
        id: galgame.id,
        contentLimit: galgame.content_limit,
        name: {
          'en-us': galgame.name_en_us,
          'ja-jp': galgame.name_ja_jp,
          'zh-cn': galgame.name_zh_cn,
          'zh-tw': galgame.name_zh_tw
        }
      }
    })),
    created: galgame.created,
    updated: galgame.updated
  }

  return data
})
