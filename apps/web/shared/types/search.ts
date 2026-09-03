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
  id: number
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

export type SearchEntityFamily = 'character' | 'company' | 'staff' | 'tag'

export interface SearchEntityItem {
  id: number
  family: SearchEntityFamily
  name: string
  alias?: string
}

export interface SearchEntityGroup {
  family: SearchEntityFamily
  total: number
  items: SearchEntityItem[]
}

export interface SearchEntityResult {
  groups: SearchEntityGroup[]
  total: number
}

export type SearchType =
  | 'all'
  | 'topic'
  | 'galgame'
  | 'entity'
  | 'user'
  | 'reply'
  | 'comment'
  | 'toolset'

export type SearchPagedType = Exclude<SearchType, 'all' | 'entity'>

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

export interface SearchOverviewTotals {
  topic: number
  galgame: number
  entity: number
  user: number
  reply: number
  comment: number
  toolset: number
}

export interface SearchOverviewResult {
  topics: SearchResultTopic[]
  galgames: SearchResultGalgame[]
  entities: SearchEntityGroup[]
  users: SearchResultUser[]
  replies: SearchResultReply[]
  comments: SearchResultComment[]
  toolsets: SearchResultToolset[]
  totals: SearchOverviewTotals
}
