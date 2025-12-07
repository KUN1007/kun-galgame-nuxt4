import { prisma } from '~~/prisma/prisma'
import { updateGalgameResourceExpireSchema } from '~/validations/galgame'

export default defineEventHandler(async (event) => {
  const input = await kunParsePutBody(event, updateGalgameResourceExpireSchema)
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
      link: {
        select: {
          url: true
        }
      }
    }
  })
  if (!resource) {
    return kunError(event, '未找到该 Galgame 资源')
  }
  if (resource.status === 1) {
    return kunError(event, '该资源已经被标记为失效')
  }

  await prisma.galgame_resource.update({
    where: { id: input.galgameResourceId },
    data: { status: 1 }
  })

  await createMessage(
    prisma,
    uid,
    resource.user_id,
    'expired',
    resource.link[0].url.slice(0, 233),
    undefined,
    resource.galgame_id
  )

  return 'MOEMOE expired galgame resource link successfully!'
})
