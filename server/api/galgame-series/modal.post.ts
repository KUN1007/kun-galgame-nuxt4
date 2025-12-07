import { prisma } from '~~/prisma/prisma'
import { getSeriesGalgameNameSchema } from '~/validations/galgame-series'
import { getPreferredLanguageText } from '~/utils/getPreferredLanguageText'
import type { GalgameSeriesSearchItem } from '~/types/api/galgame-series'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, getSeriesGalgameNameSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { ids } = input

  const galgames = await prisma.galgame.findMany({
    where: {
      id: {
        in: ids
      }
    },
    select: {
      id: true,
      name_en_us: true,
      name_ja_jp: true,
      name_zh_cn: true,
      name_zh_tw: true
    }
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

  const sortedResult = ids
    .map((id) => formattedResult.find((g) => g.id === id))
    .filter(Boolean) as { id: number; name: string }[]

  return sortedResult
})
