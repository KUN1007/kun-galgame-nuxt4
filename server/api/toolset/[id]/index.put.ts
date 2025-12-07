import { prisma } from '~~/prisma/prisma'
import { updateToolsetSchema } from '~/validations/toolset'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const input = await kunParsePutBody(event, updateToolsetSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const old = await prisma.galgame_toolset.findUnique({
    where: { id: input.toolsetId }
  })
  if (!old) {
    return kunError(event, '未找到该工具资源')
  }
  if (old.user_id !== userInfo.uid && userInfo.role < 2) {
    return kunError(event, '您没有权限更新该工具资源')
  }

  return prisma.$transaction(async (prisma) => {
    await prisma.galgame_toolset.update({
      where: { id: input.toolsetId },
      data: {
        name: input.name,
        description: input.description,
        language: input.language,
        platform: input.platform,
        type: input.type,
        version: input.version,
        homepage: input.homepage,
        edited: new Date()
      }
    })

    if (input.aliases && input.aliases.length) {
      await prisma.galgame_toolset_alias.deleteMany({
        where: { toolset_id: input.toolsetId }
      })
      await prisma.galgame_toolset_alias.createMany({
        data: input.aliases.map((name) => ({
          name,
          toolset_id: input.toolsetId
        }))
      })
    }

    return 'Moemoe update toolset successfully'
  })
})
