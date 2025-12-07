import { prisma } from '~~/prisma/prisma'
import { deleteGalgameSeriesSchema } from '~/validations/galgame-series'

export default defineEventHandler(async (event) => {
  const input = kunParseDeleteQuery(event, deleteGalgameSeriesSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }

  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限删除 Galgame 系列')
  }

  await prisma.galgame_series.delete({
    where: { id: input.seriesId }
  })

  return 'Moemoe delete galgame series successfully!'
})
