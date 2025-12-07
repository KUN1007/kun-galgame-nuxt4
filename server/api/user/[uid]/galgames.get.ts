import { prisma } from '~~/prisma/prisma'
import { getUserGalgameSchema } from '~/validations/user'
import type { GalgameCard } from '~/types/api/galgame'
import type { Prisma } from '~~/prisma/generated/prisma/client'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getUserGalgameSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { userId, page, limit, type } = input

  const where: Prisma.galgameWhereInput = {}

  switch (type) {
    case 'galgame':
      where.user_id = userId
      break
    case 'galgame_like':
      where.like = { some: { user_id: userId } }
      break
    case 'galgame_favorite':
      where.favorite = { some: { user_id: userId } }
      break
    case 'galgame_contribute':
      where.contributor = { some: { user_id: userId } }
      break

    case 'galgame_pr':
      where.pr = { some: { user_id: userId } }
      break
    case 'galgame_history':
      where.history = { some: { user_id: userId } }
      break
    case 'galgame_link':
      where.link = { some: { user_id: userId } }
      break
    case 'galgame_comment':
      where.comment = { some: { user_id: userId } }
      break
    case 'galgame_comment_target':
      where.comment = { some: { target_user_id: userId } }
      break
    case 'galgame_comment_like':
      where.comment = { some: { like: { some: { user_id: userId } } } }
      break
  }

  const [data, totalCount] = await prisma.$transaction([
    prisma.galgame.findMany({
      where,
      skip: (page - 1) * limit,
      take: limit,
      orderBy: {
        resource_update_time: 'desc'
      },
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
            like: true,
            favorite: true
          }
        },
        resource: {
          select: {
            platform: true,
            language: true
          }
        }
      }
    }),
    prisma.galgame.count({ where: where })
  ])

  const galgames: GalgameCard[] = data.map((galgame) => {
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

      user: galgame.user as unknown as KunUser,
      contentLimit: galgame.content_limit,
      view: galgame.view,
      likeCount: galgame._count.like,
      favorites: galgame._count.favorite,
      resourceUpdateTime: galgame.resource_update_time,
      platform: platforms,
      language: languages
    }
  })

  return {
    galgames,
    totalCount
  }
})
