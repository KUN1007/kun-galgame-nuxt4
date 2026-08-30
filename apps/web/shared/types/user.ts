import type { GalgameCard } from './galgame'

export interface UserInfo {
  id: number
  name: string
  avatar: string
  roles: string[]
  status: number
  moemoepoint: number
  bio: string
  created: Date | string

  upvote: number
  like: number
  dislike: number

  reply_created: number
  comment_created: number
  topic: number
  topic_poll: number
  topic_lottery: number

  galgame: number
  contribute_galgame: number
  galgame_comment: number
  galgame_rating: number

  galgame_resource: number
  galgame_toolset: number
  galgame_toolset_resource: number

  daily_topic_count: number
  daily_galgame_count: number
}

export interface UserTopic {
  id: number
  title: string
  created: Date | string
}

export type UserGalgame = GalgameCard

export interface UserGalgameResource {
  id: number
  galgame_id: number
  galgame_name: string
  type: string
  language: string
  platform: string
  size: string
  link: string[]
  code: string
  password: string
  note: string
  status: number
  created: Date | string
}

export interface UserReply {
  topic_id: number
  floor: number
  content: string
  created: Date | string
}

export interface UserComment {
  id: number
  topic_id: number
  content: string
  created: Date | string
}
