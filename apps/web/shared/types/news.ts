export interface KunNewsPublisher {
  id: number
  name: string
  avatar: string
}

export interface KunNewsSource {
  key: string
  name: string
  homepage_url: string
  column_url: string
  attribution: string
  publisher: KunNewsPublisher | null
}

export interface KunNewsItem {
  id: number
  source_key: string
  lane: 'news' | 'column'
  title: string
  preview: string
  source_url: string
  banner_url: string
  published_at: string
}

export interface KunNewsFeed {
  items: KunNewsItem[]
  sources: Record<string, KunNewsSource>
  count: number
  next_cursor: string
}

// A feed page is grouped on (date, source), never on date alone: one partner
// republishes a whole week of bulletins under a single timestamp, and a header
// that spanned two partners would print one partner's attribution over the
// other's items.
export interface KunNewsGroup {
  key: string
  date: string
  source: KunNewsSource | undefined
  items: KunNewsItem[]
}

export interface KunNewsArchiveYear {
  year: number
  count: number
}

export interface KunNewsArchiveMonth {
  month: number
  count: number
}

export interface KunNewsArchive {
  years: KunNewsArchiveYear[]
  months: KunNewsArchiveMonth[]
}

export interface KunNewsDay {
  day: number
  count: number
}

export interface KunNewsMonth {
  items: KunNewsItem[]
  sources: Record<string, KunNewsSource>
  days: KunNewsDay[]
  total: number
  count: number
  page: number
  limit: number
}
