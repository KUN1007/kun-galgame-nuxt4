export interface GalgameResource {
  id: number
  view: number
  galgame_id: number
  user: KunUser
  type: string
  language: string
  platform: string
  size: string
  status: number
  download: number
  like_count: number
  is_liked: boolean
  comment_count: number
  link_domain: string
  provider_names: string[]
  note: string
  note_html: string
  created: Date | string
  edited: Date | string | null
  dlsite_purchase_url?: string
  dlsite_coupon_url?: string
  dlsite_campaign_name?: string
}

export interface GalgameResourceDetailLink extends GalgameResource {
  link: string[]
  code: string
  note: string
  password: string
}

export interface GalgameResourceCard extends GalgameResource {
  galgame_name: string
}

export interface GalgameResourceSummary {
  id: number
  name: string
  effective_banner_hash?: string
  effective_banner_url?: string
  effective_banner_width?: number
  effective_banner_height?: number
  effective_banner_thumbhash?: string
  content_limit: string
  resource_update_time: Date | string
  view: number
  original_language: string
  age_limit: KunAgeLimit
  platform: string[]
  language: string[]
  type: string[]
}

export interface GalgameResourcePageData {
  galgame: GalgameResourceSummary
  resource: GalgameResource
  recommendations: GalgameResourceCard[]
}
