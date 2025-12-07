import { prisma } from '~~/prisma/prisma'
import type { ChatMessageAsideItem } from '~/types/api/chat-message'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const [
    latestMessage,
    messageCount,
    messageUnreadCount,
    latestSystemMessage,
    systemMessageCount,
    systemMessageUnreadCount
  ] = await prisma.$transaction([
    prisma.message.findFirst({
      where: { receiver_id: uid },
      orderBy: { created: 'desc' }
    }),
    prisma.message.count({
      where: { receiver_id: uid }
    }),
    prisma.message.count({
      where: { receiver_id: uid, status: 'unread' }
    }),
    prisma.system_message.findFirst({
      orderBy: { created: 'desc' }
    }),
    prisma.system_message.count(),
    prisma.system_message.count({
      where: { status: 'unread' }
    })
  ])

  const responseData: ChatMessageAsideItem[] = [
    {
      chatroomName: '',
      content: latestMessage ? latestMessage.content.slice(0, 100) : '',
      lastMessageTime: latestMessage?.created || '',
      count: messageCount,
      unreadCount: messageUnreadCount,
      route: 'notice',
      title: 'zako~',
      avatar: ''
    },
    {
      chatroomName: '',
      content: '',
      lastMessageTime: latestSystemMessage?.created || '',
      count: systemMessageCount,
      unreadCount: systemMessageUnreadCount,
      route: 'system',
      title: 'zako~',
      avatar: ''
    }
  ]

  return responseData
})
