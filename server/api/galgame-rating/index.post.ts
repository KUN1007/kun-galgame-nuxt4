import { prisma } from '~~/prisma/prisma'
import { createGalgameRatingSchema } from '~/validations/galgame-rating'
import type { GalgameRatingCardOnGalgamePage } from '~/types/api/galgame-rating'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, createGalgameRatingSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const { galgameId, galgameType, ...rest } = input

  const existed = await prisma.galgame_rating.findFirst({
    where: { galgame_id: galgameId, user_id: userInfo.uid }
  })
  if (existed) {
    return kunError(event, '您已经发布过该 Galgame 的评分了')
  }

  const contentLength = rest.short_summary.length
  const moemoepointIncrement =
    contentLength > 100 ? 10 : contentLength > 20 ? 5 : 3

  return prisma.$transaction(async (prisma) => {
    const res = await prisma.galgame_rating.create({
      data: {
        galgame_id: galgameId,
        user_id: userInfo.uid,
        galgame_type: galgameType,
        ...rest
      },
      include: {
        user: { select: { id: true, name: true, avatar: true } },
        _count: {
          select: {
            like: true
          }
        },
        galgame: {
          select: {
            name_en_us: true,
            name_ja_jp: true,
            name_zh_cn: true,
            name_zh_tw: true,
            content_limit: true
          }
        }
      }
    })

    await prisma.user.update({
      where: { id: userInfo.uid },
      data: { moemoepoint: { increment: moemoepointIncrement } }
    })

    const newRating: GalgameRatingCardOnGalgamePage = {
      ...res,
      galgame: {
        id: res.galgame_id,
        name: {
          'en-us': res.galgame.name_en_us,
          'ja-jp': res.galgame.name_ja_jp,
          'zh-cn': res.galgame.name_zh_cn,
          'zh-tw': res.galgame.name_zh_tw
        },
        contentLimit: res.galgame.content_limit
      },
      user: res.user,
      galgameType: res.galgame_type,
      likeCount: 0,
      isLiked: false
    }

    return newRating
  })
})
