import { prisma } from '~~/prisma/prisma'
import { deleteMessageSchema } from '~/validations/chat'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const input = kunParseDeleteQuery(event, deleteMessageSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  await prisma.chat_message.delete({
    where: { id: input.messageId, sender_id: uid }
  })

  return 'MOEMOE delete message successfully!'
})
