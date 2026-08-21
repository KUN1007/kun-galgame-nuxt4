export type WebsiteStatus = 'normal' | 'unreachable' | 'closed'

export interface WebsiteCategory {
  id: number
  name: string
  label: string
  description: string
}

export interface WebsiteCategoryListItem extends WebsiteCategory {
  sort_order: number
  website_count: number
}

export interface WebsiteTagGroup {
  id: number
  name: string
  label: string
  description: string
  sort_order: number
  multi_select: boolean
  tag_count: number
}

export interface WebsiteTag {
  id: number
  name: string
  description: string
  label: string
  level: number
  group_id: number | null
}

export interface WebsiteTagDetail {
  id: number
  name: string
  label: string
  level: number
  description: string
  group_id: number | null
  website_count: number
  websites: WebsiteCard[]
  created: Date | string
  updated: Date | string
}

export interface WebsiteCategoryDetail {
  id: number
  name: string
  label: string
  description: string
  sort_order: number
  website_count: number
  websites: WebsiteCard[]
  created: Date | string
  updated: Date | string
}

export interface WebsiteCard {
  id: number
  name: string
  description: string
  domain: string
  age_limit: string
  status: WebsiteStatus
  level: number
  icon: string
  icon_image_hash: string
  icon_url: string
  price: number
  category: string
}

export interface WebsiteDetail {
  id: number
  name: string
  url: string
  description: string
  icon: string
  icon_image_hash: string
  icon_url: string
  view: number
  language: string
  age_limit: 'all' | 'r18'
  status: WebsiteStatus
  category: WebsiteCategory
  tags: WebsiteTag[]
  like_count: number
  is_liked: boolean
  favorite_count: number
  is_favorited: boolean
  domain: string[]
  create_time: string
  comment: {
    id: number
    content: string
    user: KunUser
    created: Date | string
    updated: Date | string
  }[]

  created: Date | string
  updated: Date | string
}
