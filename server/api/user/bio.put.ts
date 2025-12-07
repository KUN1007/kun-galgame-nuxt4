import { prisma } from '~~/prisma/prisma'
import { updateUserBioSchema } from '~/validations/user'

export default defineEventHandler(async (event) => {
  const input = await kunParsePutBody(event, updateUserBioSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  await prisma.user.update({
    where: { id: userInfo.uid },
    data: { bio: input.bio }
  })

  return 'MoeMoe update bio successfully!'
})
