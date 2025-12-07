import { prisma } from '~~/prisma/prisma'
import { deleteGalgameLinkSchema } from '~/validations/galgame'

export default defineEventHandler(async (event) => {
  const input = kunParseDeleteQuery(event, deleteGalgameLinkSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const link = await prisma.galgame_link.findUnique({
    where: { id: input.galgameLinkId }
  })
  if (!link) {
    return kunError(event, '未找到相关链接')
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  if (link.user_id !== userInfo.uid) {
    return kunError(event, '您无权删除这个相关链接')
  }

  return prisma.$transaction(async (prisma) => {
    await prisma.galgame_link.delete({
      where: { id: input.galgameLinkId }
    })

    await createGalgameHistory(prisma, {
      galgame_id: link.galgame_id,
      user_id: link.user_id,
      action: 'deleted',
      type: 'link',
      content: link.name
    })

    return 'MOEMOE delete visualnovel-related link successfully!'
  })
})
