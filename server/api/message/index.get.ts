import { prisma } from '~~/prisma/prisma'
import { getNoticeMessageSchema } from '~/validations/message'
import type { z } from 'zod'
import type { MessageStatus, MessageType, Message } from '~/types/api/message'

const getMessages = async (
  input: z.infer<typeof getNoticeMessageSchema>,
  uid: number
) => {
  const { page, limit, type, sortOrder } = input
  const skip = (page - 1) * limit

  const data = await prisma.message.findMany({
    skip,
    take: limit,
    orderBy: {
      created: sortOrder
    },
    where: { receiver_id: uid, type },
    include: {
      sender: {
        select: {
          id: true,
          name: true,
          avatar: true
        }
      }
    }
  })

  const totalCount = await prisma.message.count({
    where: { receiver_id: uid, type }
  })
  const messages: Message[] = data.map((message) => ({
    id: message.id,
    sender: message.sender,
    receiverUid: message.receiver_id,
    link: message.link,
    content: message.content,
    status: message.status as MessageStatus,
    type: message.type as MessageType,
    created: message.created
  }))

  return { messages, totalCount }
}

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getNoticeMessageSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const result = await getMessages(input, userInfo.uid)

  return result
})
