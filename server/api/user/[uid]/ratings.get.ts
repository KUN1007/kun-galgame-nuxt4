import { prisma } from '~~/prisma/prisma'
import { getUserRatingSchema } from '~/validations/user'
import type { GalgameRatingCard } from '~/types/api/galgame-rating'
import type { Prisma } from '~~/prisma/generated/prisma/client'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getUserRatingSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const nsfw = getNSFWCookie(event)
  const { userId, page, limit } = input
  const skip = (page - 1) * limit

  const where: Prisma.galgame_ratingWhereInput = {
    user_id: userId,
    galgame: {
      content_limit: nsfw === 'sfw' ? 'sfw' : undefined
    }
  }

  const [rows, totalCount] = await Promise.all([
    prisma.galgame_rating.findMany({
      take: limit,
      skip,
      where,
      orderBy: { created: 'desc' },
      include: {
        user: { select: { id: true, name: true, avatar: true } },
        galgame: {
          select: {
            id: true,
            banner: true,
            name_en_us: true,
            name_ja_jp: true,
            name_zh_cn: true,
            name_zh_tw: true,
            content_limit: true
          }
        },
        _count: {
          select: {
            like: true
          }
        }
      }
    }),
    prisma.galgame_rating.count({ where })
  ])

  const ratingData: GalgameRatingCard[] = rows.map((r) => ({
    id: r.id,
    user: r.user,
    recommend: r.recommend,
    overall: r.overall,
    view: r.view,
    galgameType: r.galgame_type,
    play_status: r.play_status,
    short_summary: r.short_summary,
    spoiler_level: r.spoiler_level,
    art: r.art,
    story: r.story,
    music: r.music,
    character: r.character,
    route: r.route,
    system: r.system,
    voice: r.voice,
    replay_value: r.replay_value,
    likeCount: r._count.like,
    created: r.created,
    updated: r.updated,
    galgame: {
      id: r.galgame.id,
      contentLimit: r.galgame.content_limit,
      name: {
        'en-us': r.galgame.name_en_us,
        'ja-jp': r.galgame.name_ja_jp,
        'zh-cn': r.galgame.name_zh_cn,
        'zh-tw': r.galgame.name_zh_tw
      }
    }
  }))

  return { ratingData, totalCount }
})
