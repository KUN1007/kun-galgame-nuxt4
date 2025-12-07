import { prisma } from '~~/prisma/prisma'
import { getGalgameRankingSchema } from '~/validations/ranking'
import type { RankingGalgameItem } from '~/types/api/ranking'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameRankingSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const nsfw = getNSFWCookie(event)

  const { page, limit, sortField, sortOrder } = input
  const skip = (page - 1) * limit

  const countFields = ['like', 'favorite', 'resource']
  const isCountField = countFields.includes(sortField)

  const select = {
    id: true,
    name_en_us: true,
    name_ja_jp: true,
    name_zh_cn: true,
    name_zh_tw: true,
    banner: true,
    created: true,
    user: {
      select: {
        id: true,
        name: true,
        avatar: true
      }
    },
    ...(sortField === 'view' && { view: true }),
    ...(isCountField && {
      _count: {
        select: {
          [sortField]: true
        }
      }
    })
  }

  const galgames = await prisma.galgame.findMany({
    where: {
      content_limit: nsfw === 'sfw' ? 'sfw' : undefined
    },
    skip,
    take: limit,
    orderBy: isCountField
      ? { [sortField]: { _count: sortOrder } }
      : { [sortField]: sortOrder },
    select: select
  })

  const items: RankingGalgameItem[] = galgames.map((game) => {
    let value: number
    if (sortField === 'view') {
      value = game.view
    } else {
      value = game._count[sortField]
    }

    return {
      id: game.id,
      name: {
        'en-us': game.name_en_us,
        'ja-jp': game.name_ja_jp,
        'zh-cn': game.name_zh_cn,
        'zh-tw': game.name_zh_tw
      },
      user: game.user,
      banner: game.banner,
      created: game.created,
      value,
      sortField
    }
  })

  return items
})
