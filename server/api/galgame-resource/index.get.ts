import { prisma } from '~~/prisma/prisma'
import { getGalgameResourceSchema } from '~/validations/galgame-resource'
import type { GalgameResourceCard } from '~/types/api/galgame-resource'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameResourceSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const nsfw = getNSFWCookie(event)
  const userInfo = await getCookieTokenInfo(event)

  const { page, limit } = input
  const skip = (page - 1) * limit

  const [data, totalCount] = await Promise.all([
    prisma.galgame_resource.findMany({
      where: {
        galgame: {
          content_limit: nsfw === 'sfw' ? 'sfw' : undefined
        }
      },
      skip,
      take: limit,
      include: {
        like: {
          where: {
            user_id: userInfo?.uid
          }
        },
        _count: {
          select: { like: true }
        },
        user: {
          select: {
            id: true,
            name: true,
            avatar: true
          }
        },
        galgame: {
          select: {
            name_en_us: true,
            name_ja_jp: true,
            name_zh_cn: true,
            name_zh_tw: true
          }
        }
      },
      orderBy: { created: 'desc' }
    }),
    prisma.galgame_resource.count({
      where: {
        galgame: {
          content_limit: nsfw === 'sfw' ? 'sfw' : undefined
        }
      }
    })
  ])

  const resources: GalgameResourceCard[] = data.map((resource) => ({
    id: resource.id,
    galgameId: resource.galgame_id,
    user: resource.user,
    type: resource.type,
    language: resource.language,
    platform: resource.platform,
    size: resource.size,
    status: resource.status,
    likeCount: resource._count.like,
    isLiked: resource.like.length > 0,
    created: resource.created,
    linkDomain: '',
    download: resource.download,

    galgameName: {
      'en-us': resource.galgame.name_en_us,
      'ja-jp': resource.galgame.name_ja_jp,
      'zh-cn': resource.galgame.name_zh_cn,
      'zh-tw': resource.galgame.name_zh_tw
    }
  }))

  return { resources, totalCount }
})
