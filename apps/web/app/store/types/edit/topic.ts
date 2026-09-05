import type { TopicAccessRole, TopicAccessScope } from '~/constants/topic'

export interface EditStorePersist {
  mode: 'preview' | 'code'

  title: string
  content: string
  category: string
  section: string[]
  isNSFW: boolean
  coverImages: string[]
  accessScope: TopicAccessScope
  accessRoles: TopicAccessRole[]
  accessUserIds: number[]
}

export interface EditStoreTemp {
  id: number
  title: string
  content: string
  category: string
  section: string[]
  isNSFW: boolean
  coverImages: string[]
  accessScope: TopicAccessScope
  accessRoles: TopicAccessRole[]
  accessUserIds: number[]

  isTopicRewriting: boolean
}
