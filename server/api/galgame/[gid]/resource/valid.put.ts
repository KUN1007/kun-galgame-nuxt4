import { prisma } from '~~/prisma/prisma'
import { updateGalgameResourceValidSchema } from '~/validations/galgame'

export default defineEventHandler(async (event) => {
  const input = await kunParsePutBody(event, updateGalgameResourceValidSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const resource = await prisma.galgame_resource.findUnique({
    where: { id: input.galgameResourceId, user_id: userInfo.uid }
  })
  if (!resource) {
    return kunError(event, '未找到这个 Galgame 资源')
  }

  await prisma.galgame_resource.update({
    where: { id: input.galgameResourceId },
    data: { status: 0 }
  })

  return 'MOEMOE mark galgame resource link valid successfully!'
})
