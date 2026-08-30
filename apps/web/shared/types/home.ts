import type { GalgameCard } from './galgame.d.ts'

export interface HomeTopic {
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
  upvote_time: Date | string | null
}

export type HomeGalgame = GalgameCard
