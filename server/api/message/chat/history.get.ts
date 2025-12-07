import { prisma } from '~~/prisma/prisma'
import { getChatMessageHistorySchema } from '~/validations/chat'
import type { ChatMessage } from '~/types/api/chat-message'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const input = kunParseGetQuery(event, getChatMessageHistorySchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const receiverUid = Number(input.receiverUid)
  if (receiverUid === userInfo.uid) {
    return kunError(event, '不能给自己发送消息')
  }
  const roomId = generateRoomId(receiverUid, uid)

  return prisma.$transaction(async (prisma) => {
    const room = await prisma.chat_room.findFirst({
      where: { name: roomId }
    })
    if (!room) {
      const newRoom = await prisma.chat_room.create({
        data: {
          name: roomId,
          type: 'private'
        }
      })

      await prisma.chat_room_participant.createMany({
        data: [
          { chat_room_id: newRoom.id, user_id: uid },
          { chat_room_id: newRoom.id, user_id: receiverUid }
        ]
      })

      return []
    }

    const skip = (input.page - 1) * input.limit
    const data = await prisma.chat_message.findMany({
      skip,
      take: input.limit,
      orderBy: { id: 'desc' },
      where: { chatroom_name: roomId },
      include: {
        sender: {
          select: {
            id: true,
            name: true,
            avatar: true
          }
        },
        read_by: {
          include: {
            user: {
              select: {
                id: true,
                name: true,
                avatar: true
              }
            }
          }
        }
      }
    })

    const messageReadByArray = data.map((m) => ({
      chat_message_id: m.id,
      user_id: uid
    }))
    if (messageReadByArray.length > 0) {
      await prisma.chat_message_read_by.createMany({
        data: messageReadByArray,
        skipDuplicates: true
      })
    }

    const messages: ChatMessage[] = data.map((message) => ({
      id: message.id,
      chatroomName: message.chatroom_name,
      sender: message.sender,
      readBy: message.read_by.map((r) => r.user),
      receiverUid: message.receiver_id,
      content: message.content,
      isRecall: message.is_recall,
      created: message.created,
      recallTime: message.recall_time,
      editTime: message.edit_time
    }))

    return messages.reverse()
  })
})
