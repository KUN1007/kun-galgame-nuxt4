import { prisma } from '~~/prisma/prisma'
import { getGalgameByOfficialSchema } from '~/validations/galgame-official'
import type { GalgameCard } from '~/types/api/galgame'
import type { GalgameOfficialDetail } from '~/types/api/galgame-official'
import type { KunGalgameOfficialCategory } from '~/constants/galgameOfficial'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameByOfficialSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const nsfw = getNSFWCookie(event)

  const {
    officialId,
    page,
    limit,
    type,
    language,
    platform,
    sortField,
    sortOrder
  } = input
  const skip = (page - 1) * limit

  const resourceFilters = []
  if (type !== 'all') resourceFilters.push({ type })
  if (language !== 'all') resourceFilters.push({ language })
  if (platform !== 'all') resourceFilters.push({ platform })

  const data = await prisma.galgame_official.findUnique({
    where: { id: officialId },
    include: {
      alias: true,
      _count: {
        select: {
          galgame: true
        }
      },
      galgame: {
        where: {
          galgame: {
            content_limit: nsfw === 'sfw' ? 'sfw' : undefined,
            resource: { some: { AND: resourceFilters } }
          }
        },
        skip,
        take: limit,
        orderBy: {
          galgame: {
            [sortField === 'time' ? 'resource_update_time' : sortField]:
              sortOrder
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
    return kunError(event, '未找到这个制作会社')
  }

  const result: GalgameOfficialDetail = {
    id: data.id,
    name: data.name,
    link: data.link,
    category: data.category as KunGalgameOfficialCategory,
    lang: data.lang,
    description: data.description,
    alias: data.alias.map((a) => a.name),
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
