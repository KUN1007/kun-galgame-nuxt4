import { prisma } from '~~/prisma/prisma'
import type { ChatMessage } from '~/types/api/chat-message'

export const sendingMessage = async (
  uid: number,
  receiverUid: number,
  content: string
): Promise<ChatMessage> => {
  const senderUser = await prisma.user.findUnique({
    where: { id: uid },
    select: {
      id: true,
      name: true,
      avatar: true
    }
  })
  if (!senderUser) {
    return {
      content: '用户不存在'
    } as ChatMessage
  }

  const roomName = generateRoomId(receiverUid, uid)

  return prisma.$transaction(async (prisma) => {
    const newMessage = await prisma.chat_message.create({
      data: {
        content,
        chatroom_name: roomName,

        chat_room: { connect: { name: roomName } },
        sender: { connect: { id: uid } },
        receiver: { connect: { id: receiverUid } },

        read_by: {
          create: {
            user_id: uid
          }
        }
      }
    })

    await prisma.chat_room.update({
      where: { name: roomName },
      data: {
        last_message_content: content,
        last_message_time: new Date(),
        last_message_sender_id: uid,
        last_message_sender_name: senderUser.name
      }
    })

    const responseData: ChatMessage = {
      id: newMessage.id,
      chatroomName: newMessage.chatroom_name,
      sender: senderUser,

      readBy: [senderUser],
      receiverUid: newMessage.receiver_id,
      content: newMessage.content,
      isRecall: newMessage.is_recall,
      created: newMessage.created,
      recallTime: newMessage.recall_time,
      editTime: newMessage.edit_time
    }

    return responseData
  })
}
