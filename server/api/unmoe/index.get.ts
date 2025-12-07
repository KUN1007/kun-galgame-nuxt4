import { prisma } from '~~/prisma/prisma'
import { getUnmoeLogSchema } from '~/validations/unmoe'
import type { UnmoeLog } from '~/types/api/unmoe'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getUnmoeLogSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { page, limit } = input
  const skip = (page - 1) * limit

  const [total, data] = await Promise.all([
    prisma.unmoe.count(),
    prisma.unmoe.findMany({
      orderBy: { created: 'desc' },
      skip,
      take: limit,
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
  ])

  const logs: UnmoeLog[] = data.map((log) => ({
    id: log.id,
    user: log.user,
    name: log.name,
    description: {
      'en-us': log.desc_en_us,
      'ja-jp': log.desc_ja_jp,
      'zh-cn': log.desc_zh_cn,
      'zh-tw': log.desc_zh_tw
    },
    result: log.result,
    created: log.created
  }))

  return { logs, total }
})
