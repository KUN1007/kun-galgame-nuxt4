import { prisma } from '~~/prisma/prisma'
import { getSearchResultSchema } from '~/validations/galgame-series'
import { getPreferredLanguageText } from '~/utils/getPreferredLanguageText'
import type { GalgameSeriesSearchItem } from '~/types/api/galgame-series'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getSearchResultSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const keywordsArray: string[] = input.keywords
    .trim()
    .slice(0, 107)
    .split(' ')
    .filter((keyword) => keyword.trim() !== '')

  const galgames = await prisma.galgame.findMany({
    where: {
      status: { not: 1 },
      AND: keywordsArray.map((kw) => ({
        OR: [
          { vndb_id: { in: keywordsArray } },
          {
            tag: {
              some: {
                tag: {
                  name: { contains: kw, mode: 'insensitive' }
                }
              }
            }
          },
          { name_en_us: { contains: kw, mode: 'insensitive' } },
          { name_ja_jp: { contains: kw, mode: 'insensitive' } },
          { name_zh_cn: { contains: kw, mode: 'insensitive' } },
          { name_zh_tw: { contains: kw, mode: 'insensitive' } },
          {
            alias: {
              some: {
                name: { contains: kw, mode: 'insensitive' }
              }
            }
          }
        ]
      }))
    },
    select: {
      id: true,
      name_en_us: true,
      name_ja_jp: true,
      name_zh_cn: true,
      name_zh_tw: true
    },
    take: 20
  })

  const formattedResult: GalgameSeriesSearchItem[] = galgames.map((g) => ({
    id: g.id,
    name: getPreferredLanguageText({
      'en-us': g.name_en_us,
      'ja-jp': g.name_ja_jp,
      'zh-cn': g.name_zh_cn,
      'zh-tw': g.name_zh_tw
    })
  }))

  return formattedResult
})
