import { prisma } from '~~/prisma/prisma'
import { updateGalgameEngineSchema } from '~/validations/galgame-engine'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限更新引擎')
  }

  const input = await kunParsePutBody(event, updateGalgameEngineSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { engineId, ...dataToUpdate } = input

  await prisma.galgame_engine.update({
    where: { id: engineId },
    data: dataToUpdate
  })

  return 'Moemoe update galgame engine successfully!'
})
