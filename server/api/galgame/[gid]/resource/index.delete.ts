import { prisma } from '~~/prisma/prisma'
import { deleteGalgameResourceSchema } from '~/validations/galgame'

export default defineEventHandler(async (event) => {
  const input = kunParseDeleteQuery(event, deleteGalgameResourceSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const resource = await prisma.galgame_resource.findUnique({
    where: { id: input.galgameResourceId },
    include: {
      _count: {
        select: {
          like: true
        }
      }
    }
  })
  if (!resource) {
    return kunError(event, '未找到该 Galgame 资源')
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  if (resource.user_id !== userInfo.uid && userInfo.role < 2) {
    return kunError(event, '您没有权限删除这个 Galgame 资源')
  }

  return prisma.$transaction(async (prisma) => {
    await prisma.user.update({
      where: { id: resource.user_id },
      data: {
        moemoepoint: { increment: -(resource._count.like + 5) }
      }
    })

    await prisma.galgame_resource.delete({
      where: { id: resource.id }
    })

    return 'MOEMOE delete visualnovel resource successfully!'
  })
})
