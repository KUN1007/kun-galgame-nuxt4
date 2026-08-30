export type ActivityEventType =
  | 'GALGAME_CREATION'
  | 'GALGAME_COMMENT_CREATION'
  | 'GALGAME_RATING_CREATION'
  | 'GALGAME_RATING_COMMENT_CREATION'
  | 'GALGAME_PR_CREATION'
  | 'GALGAME_EDIT'
  | 'GALGAME_WEBSITE_CREATION'
  | 'GALGAME_WEBSITE_COMMENT_CREATION'
  | 'GALGAME_RESOURCE_CREATION'
  | 'GALGAME_QUIZ_CREATION'
  | 'TOOLSET_CREATION'
  | 'TOOLSET_RESOURCE_CREATION'
  | 'TOOLSET_COMMENT_CREATION'
  | 'TOPIC_CREATION'
  | 'TOPIC_REPLY_CREATION'
  | 'TOPIC_COMMENT_CREATION'
  | 'TOPIC_UPVOTE'
  | 'TODO_CREATION'
  | 'UPDATE_LOG_CREATION'
  | 'MESSAGE_UPVOTE'
  | 'MESSAGE_SOLUTION'

export interface ActivityTopReply {
  reply_id: number
  floor: number
  user: KunUser
  content: string
  like_count: number
}

export interface TopicActivityData {
  topic_id: number
  title?: string
  author_id?: number
  excerpt: string
  sections: string[]
  cover_images: string[]
  cover_image_meta?: Record<string, KunImageMeta>
  view: number
  like_count: number
  favorite_count: number
  reply_count: number
  comment_count: number
  upvote_time: Date | string | null
  edited?: Date | string | null
  has_best_answer: boolean
  mini_apps: string[]
  is_nsfw: boolean
  top_reply?: ActivityTopReply
  best_answer?: ActivityTopReply
  upvotes?: {
    id: number
    user: KunUser
    description: string
    created: Date | string
  }[]
  latest_activity?: {
    kind: 'reply' | 'comment'
    reply_id: number
    floor: number
    comment_id: number
    user: KunUser
    content: string
    created: Date | string
  }
  reactions: { reaction: string; count: number; reactors?: KunUser[] }[]
}

export interface GalgameActivityData {
  name: string
  cover_hash: string
  language: string
  age_limit: string
  release_date: string | null
  galgame_id?: number
  resource_count?: number
  like_count?: number
  favorite_count?: number
  revision_id?: number
  revision_number?: number
  developer?: string
  intro?: string
  rating?: ActivityRatingInfo
  parent_comment?: { content: string }
  resource?: {
    type: string
    language: string
    platform: string
    size: string
    note: string
    like_count: number
  }
}

export interface ActivityRatingInfo {
  rating_id: number
  overall: number
  play_status: string
  recommend: string
  short_summary: string
  spoiler_level: string
  like_count: number
  author_id: number
}

export interface ActivityQuotedReply {
  floor: number
  content: string
}

export interface ReplyActivityData {
  topic_title: string
  floor: number
  quoted_reply?: ActivityQuotedReply
}

export interface TopicCommentActivityData {
  topic_title: string
  comment_id: number
  quoted_reply?: ActivityQuotedReply
}

export interface NoteActivityData {
  version?: string
  status?: number
}

export interface EntityRefActivityData {
  parent_name: string
}

export interface SolutionActivityData {
  topic_title: string
  floor: number
}

export interface QuizActivityData {
  category: string
  type: string
  difficulty: number
  answer_count: number
  correct_count: number
  favorite_count: number
  description: string
}

export type ActivityData =
  | TopicActivityData
  | GalgameActivityData
  | ReplyActivityData
  | TopicCommentActivityData
  | NoteActivityData
  | EntityRefActivityData
  | SolutionActivityData
  | QuizActivityData

export interface ActivityItem {
  unique_id: string
  type: ActivityEventType
  timestamp: Date | string
  actor: KunUser
  link: string
  content: string
  data?: ActivityData
}
