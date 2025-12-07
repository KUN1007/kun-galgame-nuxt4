import { prisma } from '~~/prisma/prisma'
import { updateGalgameFavoriteSchema } from '~/validations/galgame'

export default defineEventHandler(async (event) => {
  const input = await kunParsePutBody(event, updateGalgameFavoriteSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const uid = userInfo.uid
  const galgameId = input.galgameId

  const galgame = await prisma.galgame.findUnique({
    where: { id: galgameId, status: { not: 1 } },
    include: {
      favorite: {
        where: {
          user_id: uid
        }
      }
    }
  })
  if (!galgame) {
    return kunError(event, '未找到这个 Galgame')
  }

  const isFavoriteGalgame = galgame.favorite.length > 0

  return prisma.$transaction(async (prisma) => {
    if (isFavoriteGalgame) {
      await prisma.galgame_favorite.delete({
        where: {
          galgame_id_user_id: {
            user_id: uid,
            galgame_id: galgameId
          }
        }
      })
    } else {
      await prisma.galgame_favorite.create({
        data: {
          user_id: uid,
          galgame_id: galgameId
        }
      })
    }

    if (uid !== galgame.user_id) {
      await prisma.user.update({
        where: { id: uid },
        data: {
          moemoepoint: {
            increment: isFavoriteGalgame ? -1 : 1
          }
        }
      })

      if (!isFavoriteGalgame) {
        await createDedupMessage(
          prisma,
          uid,
          galgame.user_id,
          'favorite',
          galgame.name_zh_cn ||
            galgame.name_zh_tw ||
            galgame.name_ja_jp ||
            galgame.name_en_us,
          undefined,
          galgameId
        )
      }
    }

    return 'MOEMOE favorite galgame operation successfully!'
  })
})
