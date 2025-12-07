import { prisma } from '~~/prisma/prisma'
import { createToolsetSchema } from '~/validations/toolset'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const input = await kunParsePostBody(event, createToolsetSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  return prisma.$transaction(async (prisma) => {
    const created = await prisma.galgame_toolset.create({
      data: {
        name: input.name,
        description: input.description,
        language: input.language,
        platform: input.platform,
        type: input.type,
        version: input.version,
        homepage: input.homepage,
        user_id: userInfo.uid
      }
    })

    await prisma.user.update({
      where: { id: userInfo.uid },
      data: { moemoepoint: { increment: 3 } }
    })

    if (input.aliases.length) {
      await prisma.galgame_toolset_alias.createMany({
        data: input.aliases.map((name) => ({ name, toolset_id: created.id }))
      })
    }

    return created.id
  })
})
