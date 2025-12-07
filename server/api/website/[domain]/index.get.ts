import { prisma } from '~~/prisma/prisma'
import { getWebsiteDetailSchema } from '~/validations/website'
import type { WebsiteDetail } from '~/types/api/website'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getWebsiteDetailSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  const uid = userInfo?.uid

  const data = await prisma.galgame_website.findUnique({
    where: { url: input.domain },
    include: {
      like: { where: { user_id: uid } },
      favorite: { where: { user_id: uid } },
      comment: {
        include: {
          user: {
            select: { id: true, name: true, avatar: true }
          }
        }
      },
      category: true,
      tag: {
        include: {
          tag: true
        }
      },
      _count: {
        select: { like: true, favorite: true, comment: true }
      }
    }
  })

  if (!data) {
    return kunError(event, '未找到该网站')
  }

  await prisma.galgame_website.update({
    where: { url: input.domain },
    data: { view: { increment: 1 } }
  })

  const website: WebsiteDetail = {
    id: data.id,
    name: data.name,
    url: data.url,
    description: data.description,
    icon: data.icon,
    view: data.view,
    language: data.language,
    ageLimit: data.age_limit as 'all' | 'r18',
    category: data.category,
    tags: data.tag.map((t) => t.tag),
    likeCount: data._count.like,
    isLiked: data.like.length > 0,
    domain: data.domain,
    favoriteCount: data._count.favorite,
    isFavorited: data.favorite.length > 0,
    createTime: data.create_time,
    comment: data.comment,
    created: data.created,
    updated: data.updated
  }

  return website
})
