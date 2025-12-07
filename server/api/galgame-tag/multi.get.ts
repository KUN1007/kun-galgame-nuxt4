import { prisma } from '~~/prisma/prisma'
import { getGalgameByTagsSchema } from '~/validations/galgame-tag'
import type { GalgameCard } from '~/types/api/galgame'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameByTagsSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const nsfw = getNSFWCookie(event)
  const { page, limit, tagIds } = input
  const skip = (page - 1) * limit

  const andTags = tagIds.map((id: number) => ({
    tag: {
      some: { tag_id: id }
    }
  }))

  const where = {
    status: { not: 1 },
    content_limit: nsfw === 'sfw' ? 'sfw' : undefined,
    AND: andTags
  } as const

  const [totalCount, data] = await Promise.all([
    prisma.galgame.count({ where }),
    prisma.galgame.findMany({
      where,
      orderBy: { resource_update_time: 'desc' },
      skip,
      take: limit,
      include: {
        user: {
          select: { id: true, name: true, avatar: true }
        },
        _count: { select: { like: true } },
        resource: { select: { platform: true, language: true } }
      }
    })
  ])

  const galgames: GalgameCard[] = data.map((g) => {
    const platforms = [...new Set(g.resource.map((r) => r.platform))]
    const languages = [...new Set(g.resource.map((r) => r.language))]
    return {
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
  })

  return { galgames, totalCount }
})
