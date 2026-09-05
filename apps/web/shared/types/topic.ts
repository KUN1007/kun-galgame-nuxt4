export interface TopicCard {
  id: number
  title: string
  view: number
  section: string[]
  cover_images: string[]
  cover_image_meta?: Record<string, KunImageMeta>
  user: KunUser
  status: number
  has_best_answer: boolean
  mini_apps: string[]
  is_nsfw_topic: boolean
  like_count: number
  reply_count: number
  comment_count: number
  status_update_time: Date | string
  created: Date | string
  upvote_time: Date | string | null
}

export interface TopicBestAnswerSummary {
  id: number
  floor: number
  user: KunUser
  content_markdown: string
  content_html: string
  created: Date | string
}

export interface TopicAccessGrants {
  roles: string[]
  user_ids: number[]
}

export interface TopicDetail {
  id: number
  title: string
  content_markdown: string
  content_html: string
  view: number
  status: number
  hidden_by: string
  access_scope: string
  access_grants?: TopicAccessGrants
  is_nsfw: boolean
  category: string
  section: string[]
  cover_images: string[]
  cover_image_meta?: Record<string, KunImageMeta>
  user: KunUser & { moemoepoint: number }

  like_count: number
  is_liked: boolean
  dislike_count: number
  is_disliked: boolean
  favorite_count: number
  is_favorited: boolean
  upvote_count: number
  is_upvoted: boolean
  reactions: KunReaction[]

  reply_count: number
  mini_apps: string[]

  status_update_time: Date | string
  upvote_time: Date | string | null
  edited: Date | string | null
  created: Date | string

  best_answer?: TopicBestAnswerSummary
}
