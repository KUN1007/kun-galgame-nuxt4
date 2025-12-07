import { prisma } from '~~/prisma/prisma'
import { createGalgameLinkSchema } from '~/validations/galgame'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, createGalgameLinkSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const uid = userInfo.uid
  const { galgameId, name, link } = input

  return prisma.$transaction(async (prisma) => {
    await prisma.galgame_link.create({
      data: {
        user_id: uid,
        galgame_id: galgameId,
        name,
        link
      }
    })

    await prisma.galgame_contributor.createMany({
      data: [{ user_id: uid, galgame_id: galgameId }],
      skipDuplicates: true
    })

    await createGalgameHistory(prisma, {
      galgame_id: galgameId,
      user_id: uid,
      action: 'created',
      type: 'link',
      content: name
    })

    return 'MOEMOE create visualnovel-related link successfully!'
  })
})
