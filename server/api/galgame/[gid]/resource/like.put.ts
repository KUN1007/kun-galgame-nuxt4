import { prisma } from '~~/prisma/prisma'
import { updateGalgameResourceLikeSchema } from '~/validations/galgame'

export default defineEventHandler(async (event) => {
  const input = await kunParsePutBody(event, updateGalgameResourceLikeSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const resource = await prisma.galgame_resource.findUnique({
    where: { id: input.galgameResourceId },
    include: {
      like: {
        where: {
          user_id: uid
        }
      },
      link: {
        select: {
          url: true
        }
      }
    }
  })
  if (!resource) {
    return kunError(event, '未找到该资源')
  }
  if (resource.user_id === uid) {
    return kunError(event, '您不能给自己的资源点赞')
  }

  const isLikedResource = resource.like.length > 0

  return prisma.$transaction(async (prisma) => {
    if (isLikedResource) {
      await prisma.galgame_resource_like.delete({
        where: {
          galgame_resource_id_user_id: {
            user_id: uid,
            galgame_resource_id: input.galgameResourceId
          }
        }
      })
    } else {
      await prisma.galgame_resource_like.create({
        data: {
          galgame_resource_id: input.galgameResourceId,
          user_id: uid
        }
      })
    }

    await prisma.user.update({
      where: { id: uid },
      data: { moemoepoint: { increment: isLikedResource ? -1 : 1 } }
    })

    await createDedupMessage(
      prisma,
      uid,
      resource.user_id,
      'liked',
      resource.link[0].url.slice(0, 233),
      undefined,
      resource.galgame_id
    )

    return 'MOEMOE like galgame resource operation successfully!'
  })
})
