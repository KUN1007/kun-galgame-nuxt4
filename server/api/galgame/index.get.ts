import { prisma } from '~~/prisma/prisma'
import { getGalgameSchema } from '~/validations/galgame'
import { PROVIDER_KEY_OPTIONS } from '~/constants/galgameResource'
import type { GalgameCard } from '~/types/api/galgame'
import type { Prisma } from '~~/prisma/generated/prisma/client'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const nsfw = getNSFWCookie(event)
  const {
    page,
    limit,
    type,
    language,
    platform,
    sortField,
    sortOrder,
    includeProviders,
    excludeOnlyProviders
  } = input

  const skip = (page - 1) * limit

  const resourceFilters = []
  if (type !== 'all') {
    resourceFilters.push({ type })
  }
  if (language !== 'all') {
    resourceFilters.push({ language })
  }
  if (platform !== 'all') {
    resourceFilters.push({ platform })
  }

  // Build provider based conditions
  const providerConditions: Prisma.galgameWhereInput[] = []
  if (includeProviders.length > 0) {
    providerConditions.push({
      resource: {
        some: {
          AND: [...resourceFilters, { provider: { hasSome: includeProviders } }]
        }
      }
    })
  }
  if (excludeOnlyProviders.length > 0) {
    const allowed = PROVIDER_KEY_OPTIONS.filter(
      (p) => !excludeOnlyProviders.includes(p)
    )
    providerConditions.push({
      resource: {
        some: {
          AND: [...resourceFilters, { provider: { hasSome: allowed } }]
        }
      }
    })
  }

  const where = {
    status: { not: 1 },
    content_limit: nsfw === 'sfw' ? 'sfw' : undefined,
    AND: [
      {
        resource: {
          some: {
            AND: resourceFilters
          }
        }
      },
      ...providerConditions
    ]
  }

  const orderBy = {
    [sortField === 'time' ? 'resource_update_time' : sortField]: sortOrder
  }

  const [totalCount, data] = await Promise.all([
    prisma.galgame.count({ where }),
    prisma.galgame.findMany({
      where,
      orderBy,
      skip,
      take: limit,
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
    })
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
      user: galgame.user,
      contentLimit: galgame.content_limit,
      view: galgame.view,
      likeCount: galgame._count.like,
      resourceUpdateTime: galgame.resource_update_time,
      platform: platforms,
      language: languages
    }
  })

  return { galgames, totalCount }
})
