export interface SectionTopic {
  id: number
  title: string
  content: string
  view: number
  like_count: number
  reply_count: number
  has_best_answer: boolean
  is_nsfw_topic: boolean
  mini_apps: string[]
  user: KunUser
  created: Date | string
}

export interface SectionTopicList {
  topics: SectionTopic[]
  total: number
}
