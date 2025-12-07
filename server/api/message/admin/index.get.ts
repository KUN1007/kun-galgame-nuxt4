import { prisma } from '~~/prisma/prisma'
import type { MessageSystemMessage } from '~/types/api/message'

export default defineEventHandler(async (_) => {
  const data = await prisma.system_message.findMany({
    orderBy: { created: 'desc' },
    include: {
      user: {
        select: {
          id: true,
          name: true,
          avatar: true
        }
      }
    }
  })

  const messages: MessageSystemMessage[] = data.map((message) => ({
    id: message.id,
    status: message.status as 'read' | 'unread',
    content: {
      'en-us': message.content_en_us,
      'ja-jp': message.content_ja_jp,
      'zh-cn': message.content_zh_cn,
      'zh-tw': message.content_zh_tw
    },
    admin: message.user,
    created: message.created
  }))

  return messages
})
