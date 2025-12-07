import { prisma } from '~~/prisma/prisma'
import { updateToolsetPracticalitySchema } from '~/validations/toolset'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const input = await kunParsePutBody(event, updateToolsetPracticalitySchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const exist = await prisma.galgame_toolset_practicality.findFirst({
    where: { toolset_id: input.toolsetId, user_id: userInfo.uid }
  })

  if (exist) {
    await prisma.galgame_toolset_practicality.update({
      where: { id: exist.id },
      data: { rate: input.rate }
    })
  } else {
    await prisma.galgame_toolset_practicality.create({
      data: {
        toolset_id: input.toolsetId,
        user_id: userInfo.uid,
        rate: input.rate
      }
    })
  }

  return 'Moemoe rating toolset successfully!'
})
