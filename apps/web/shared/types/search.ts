import type { HomeTopic, HomeGalgame } from './home'
import type { ToolsetCard } from './toolset'

export type SearchResultTopic = HomeTopic
export type SearchResultGalgame = HomeGalgame
export type SearchResultToolset = ToolsetCard

export interface SearchResultUser extends KunUser {
  bio: string
  moemoepoint?: number
  created?: Date | string
}

export interface SearchResultReply {
  topic_id: number
  topic_title: string
  floor: number
  content: string
  user: KunUser
  created: Date | string
}

export type SearchResultComment = {
  id: number
  topic_id: number
  topic_title: string
  content: string
  user: KunUser
  target_user?: KunUser
  created: Date | string
}

export type SearchType =
  | 'topic'
  | 'galgame'
  | 'toolset'
  | 'user'
  | 'reply'
  | 'comment'
export type SearchResult =
  | SearchResultTopic
  | SearchResultGalgame
  | SearchResultToolset
  | SearchResultUser
  | SearchResultReply
  | SearchResultComment

export interface SearchQuickTotals {
  topic: number
  galgame: number
  user: number
}

export interface SearchQuickResult {
  topics: SearchResultTopic[]
  galgames: SearchResultGalgame[]
  users: SearchResultUser[]
  totals: SearchQuickTotals
}
