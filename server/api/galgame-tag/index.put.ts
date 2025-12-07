import { prisma } from '~~/prisma/prisma'
import { updateGalgameTagSchema } from '~/validations/galgame-tag'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限更新标签')
  }

  const input = await kunParsePutBody(event, updateGalgameTagSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { tagId, alias, ...dataToUpdate } = input

  await prisma.$transaction(async (prisma) => {
    await prisma.galgame_tag.update({
      where: { id: tagId },
      data: dataToUpdate
    })

    if (alias.length) {
      await prisma.galgame_tag_alias.deleteMany({
        where: { galgame_tag_id: tagId }
      })

      if (alias.length > 0) {
        await prisma.galgame_tag_alias.createMany({
          data: alias.map((name) => ({
            name: name,
            galgame_tag_id: tagId
          }))
        })
      }
    }
  })

  return 'Moemoe update galgame tag successfully!'
})
