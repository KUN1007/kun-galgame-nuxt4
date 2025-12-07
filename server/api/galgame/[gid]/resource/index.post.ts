import { prisma } from '~~/prisma/prisma'
import { createGalgameResourceSchema } from '~/validations/galgame'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, createGalgameResourceSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const { link, galgameId, ...rest } = input

  const providers = detectProvidersFromUrls(link)

  return prisma.$transaction(async (prisma) => {
    const resource = await prisma.galgame_resource.create({
      data: {
        ...rest,
        galgame_id: galgameId,
        user_id: userInfo.uid,
        provider: providers
      }
    })

    const linksData = link.map((l) => ({
      galgame_resource_id: resource.id,
      url: l
    }))

    if (linksData.length > 0) {
      await prisma.galgame_resource_link.createMany({
        data: linksData,
        skipDuplicates: true
      })
    }

    await prisma.user.update({
      where: { id: userInfo.uid },
      data: { moemoepoint: { increment: 3 } }
    })

    await prisma.galgame_contributor.createMany({
      data: [
        {
          galgame_id: galgameId,
          user_id: resource.user_id
        }
      ],
      skipDuplicates: true
    })

    await prisma.galgame.update({
      where: { id: galgameId },
      data: { resource_update_time: new Date() }
    })

    return 'MOEMOE create galgame resource successfully!'
  })
})
