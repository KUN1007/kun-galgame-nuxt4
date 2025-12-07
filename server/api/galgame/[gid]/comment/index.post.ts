import { prisma } from '~~/prisma/prisma'
import { createGalgameCommentSchema } from '~/validations/galgame'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, createGalgameCommentSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const { galgameId, targetUserId, content } = input

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  return prisma.$transaction(async (prisma) => {
    await prisma.galgame_comment.create({
      data: {
        galgame_id: galgameId,
        target_user_id: targetUserId,
        content: content,
        user_id: userInfo.uid
      }
    })

    if (userInfo.uid !== targetUserId) {
      await prisma.user.update({
        where: { id: targetUserId },
        data: { moemoepoint: { increment: 1 } }
      })

      await createMessage(
        prisma,
        userInfo.uid,
        targetUserId,
        'commented',
        content.slice(0, 233),
        undefined,
        galgameId
      )
    }

    return 'MOEMOE publish galgame comment successfully!'
  })
})
