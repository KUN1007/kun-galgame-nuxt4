import { prisma } from '~~/prisma/prisma'
import { getGalgameResourceSchema } from '~/validations/galgame'
import { getDomain } from '~/utils/getDomain'
import type { GalgameResource } from '~/types/api/galgame-resource'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameResourceSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)

  const data = await prisma.galgame_resource.findMany({
    where: { galgame_id: input.galgameId },
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
      link: {
        select: {
          url: true
        }
      }
    },
    orderBy: { created: 'desc' }
  })

  const galgames: GalgameResource[] = data.map((resource) => {
    const linkDomain = [
      ...new Set(
        resource.link.map((l) => getDomain(l.url, { getRootDomain: true }))
      )
    ]
    return {
      id: resource.id,
      galgameId: resource.galgame_id,
      user: resource.user,
      type: resource.type,
      language: resource.language,
      platform: resource.platform,
      size: resource.size,
      status: resource.status,
      download: resource.download,
      likeCount: resource._count.like,
      isLiked: resource.like.length > 0,
      linkDomain: linkDomain.length ? linkDomain[0] : '',
      created: resource.created
    }
  })

  return galgames
})
