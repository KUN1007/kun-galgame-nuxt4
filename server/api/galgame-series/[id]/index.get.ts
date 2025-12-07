import { prisma } from '~~/prisma/prisma'
import { getSeriesDetailSchema } from '~/validations/galgame-series'
import type {
  GalgameSeriesDetail,
  GalgameSeriesSample
} from '~/types/api/galgame-series'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getSeriesDetailSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const nsfw = getNSFWCookie(event)

  const data = await prisma.galgame_series.findUnique({
    where: { id: input.seriesId },
    include: {
      _count: {
        select: {
          galgame: true
        }
      },
      galgame: {
        where: { content_limit: nsfw === 'sfw' ? 'sfw' : undefined },
        include: {
          user: {
            select: {
              id: true,
              name: true,
              avatar: true
            }
          },
          _count: {
            select: {
              like: true
            }
          },
          resource: {
            select: {
              platform: true,
              language: true
            }
          }
        },
        orderBy: {
          created: 'desc'
        }
      }
    }
  })
  if (!data) {
    return kunError(event, '未找到这个 Galgame 系列')
  }

  const galgameData = data.galgame.map((galgame) => {
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

  const isNSFW = data.galgame.some((g) => g.content_limit === 'nsfw')
  const sampleGalgame: GalgameSeriesSample[] = data.galgame.map((g) => ({
    banner: g.banner,
    name: {
      'en-us': g.name_en_us,
      'ja-jp': g.name_ja_jp,
      'zh-cn': g.name_zh_cn,
      'zh-tw': g.name_zh_tw
    }
  }))

  const series: GalgameSeriesDetail = {
    id: data.id,
    name: data.name,
    isNSFW,
    description: data.description,
    sampleGalgame,
    galgameCount: data._count.galgame,
    galgame: galgameData,
    created: data.created,
    updated: data.updated
  }

  return series
})
