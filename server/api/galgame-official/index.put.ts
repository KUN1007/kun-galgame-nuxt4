import { prisma } from '~~/prisma/prisma'
import { updateGalgameOfficialSchema } from '~/validations/galgame-official'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限更新制作会社信息')
  }

  const input = await kunParsePutBody(event, updateGalgameOfficialSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { officialId, alias, ...dataToUpdate } = input

  await prisma.$transaction(async (tx) => {
    await tx.galgame_official.update({
      where: { id: officialId },
      data: dataToUpdate
    })

    if (alias.length) {
      await tx.galgame_official_alias.deleteMany({
        where: { galgame_official_id: officialId }
      })

      if (alias.length > 0) {
        await tx.galgame_official_alias.createMany({
          data: alias.map((name) => ({
            name: name,
            galgame_official_id: officialId
          }))
        })
      }
    }
  })

  return 'Moemoe update galgame official successfully!'
})
