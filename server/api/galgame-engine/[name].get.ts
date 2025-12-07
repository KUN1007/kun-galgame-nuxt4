import { prisma } from '~~/prisma/prisma'
import { getGalgameByEngineSchema } from '~/validations/galgame-engine'
import type { GalgameCard } from '~/types/api/galgame'
import type { GalgameEngineDetail } from '~/types/api/galgame-engine'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameByEngineSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const nsfw = getNSFWCookie(event)

  const { engineId, page, limit } = input
  const skip = (page - 1) * limit

  const data = await prisma.galgame_engine.findUnique({
    where: { id: engineId },
    include: {
      _count: {
        select: {
          galgame: true
        }
      },
      galgame: {
        where: {
          galgame: {
            content_limit: nsfw === 'sfw' ? 'sfw' : undefined
          }
        },
        skip,
        take: limit,
        orderBy: {
          galgame: {
            resource_update_time: 'desc'
          }
        },
        include: {
          galgame: {
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
            }
          }
        }
      }
    }
  })

  if (!data) {
    return kunError(event, '未找到这个引擎')
  }

  const result: GalgameEngineDetail = {
    id: data.id,
    name: data.name,
    description: data.description,
    alias: data.alias,
    galgameCount: data._count.galgame,
    galgame: data.galgame.map((relation) => {
      const g = relation.galgame
      const platforms = [...new Set(g.resource.map((r) => r.platform))]
      const languages = [...new Set(g.resource.map((r) => r.language))]

      const galgameCard: GalgameCard = {
        id: g.id,
        name: {
          'en-us': g.name_en_us,
          'ja-jp': g.name_ja_jp,
          'zh-cn': g.name_zh_cn,
          'zh-tw': g.name_zh_tw
        },
        banner: g.banner,
        user: g.user,
        contentLimit: g.content_limit,
        view: g.view,
        likeCount: g._count.like,
        platform: platforms,
        language: languages,
        resourceUpdateTime: g.resource_update_time
      }
      return galgameCard
    })
  }

  return result
})
