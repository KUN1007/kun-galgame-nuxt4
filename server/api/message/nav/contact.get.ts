import { prisma } from '~~/prisma/prisma'
import type { ChatMessageAsideItem } from '~/types/api/chat-message'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const userId = userInfo.uid

  const chatRooms = await prisma.chat_room.findMany({
    where: {
      participant: {
        some: { user_id: userId }
      },
      last_message_sender_id: {
        not: 0
      },
      last_message_time: {
        not: null
      }
    },
    orderBy: {
      last_message_time: 'desc'
    },
    include: {
      participant: {
        include: {
          user: {
            select: {
              id: true,
              name: true,
              avatar: true
            }
          }
        }
      },
      message: {
        where: {
          sender_id: {
            not: userId
          },
          read_by: {
            none: {
              user_id: userId
            }
          }
        },
        select: {
          id: true
        }
      },
      _count: {
        select: {
          message: true
        }
      }
    }
  })

  if (!chatRooms.length) {
    return []
  }

  const asideItems: ChatMessageAsideItem[] = chatRooms.map((room) => {
    let title = room.name
    let avatar = room.avatar
    let route = room.name

    if (room.type === 'private') {
      const otherParticipant = room.participant.find(
        (p) => p.user_id !== userId
      )

      if (otherParticipant && otherParticipant.user) {
        title = otherParticipant.user.name
        avatar = otherParticipant.user.avatar
        route = otherParticipant.user_id.toString()
      }
    }

    return {
      chatroomName: room.name,
      content: room.last_message_content,
      lastMessageTime: room.last_message_time || '',
      count: room._count.message,
      unreadCount: room.message.length,
      route: route,
      title: title,
      avatar: avatar
    }
  })

  return asideItems
})
