export type MessageType =
  | 'upvoted'
  | 'liked'
  | 'favorite'
  | 'replied'
  | 'solution'
  | 'pin-reply'
  | 'commented'
  | 'expired'
  | 'requested'
  | 'merged'
  | 'declined'
  | 'mentioned'
  | 'admin'
  | 'quiz-answered'
  | 'lottery-won'
  | 'lottery-closed'
  | 'lottery-expired'
  | 'poll-closed'

type MessageStatus = 'read' | 'unread'

export interface NotificationPreference {
  muted_types: string[]
}

export interface Message {
  id: number
  sender: KunUser
  receiver_id: number
  link: string
  content: string
  status: MessageStatus
  type: MessageType
  created: Date | string
}

export interface MessageList {
  messages: Message[]
  total: number
}

export interface MessageSystemMessage {
  id: number
  is_read: boolean
  content: string
  admin: KunUser
  created: Date | string
}
