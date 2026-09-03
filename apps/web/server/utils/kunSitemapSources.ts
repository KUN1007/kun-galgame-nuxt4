interface SitemapUrl {
  loc: string
  lastmod?: string
  changefreq?: string
  priority?: number
}

const SFW_COOKIE = `KUNGalgameSettings=${encodeURIComponent(
  JSON.stringify({ showKUNGalgameContentLimit: 'sfw' })
)}`

const PAGE_SIZE = 50
const MAX_PAGES = 130
const GLOBAL_CONCURRENCY = 12

const createLimiter = (max: number) => {
  let active = 0
  const waiters: Array<() => void> = []
  const release = () => {
    active--
    waiters.shift()?.()
  }
  return async <T>(fn: () => Promise<T>): Promise<T> => {
    if (active >= max)
      await new Promise<void>((resolve) => waiters.push(resolve))
    active++
    try {
      return await fn()
    } finally {
      release()
    }
  }
}

const range = (from: number, to: number): number[] =>
  Array.from({ length: Math.max(0, to - from + 1) }, (_, i) => from + i)

const unwrap = (json: unknown): unknown =>
  (json as { data?: unknown })?.data ?? json

const toIso = (d: unknown): string | undefined => {
  if (typeof d === 'string' || typeof d === 'number' || d instanceof Date) {
    const t = new Date(d)
    if (!Number.isNaN(t.getTime())) return t.toISOString()
  }
  return undefined
}

interface PagedSource {
  path: string
  pick: (data: unknown) => Record<string, unknown>[]
  total?: (data: unknown) => number | undefined
  loc: (row: Record<string, unknown>) => string
  lastmod?: (row: Record<string, unknown>) => string | undefined
  priority: number
}

export const buildSitemapUrls = async (
  apiBase: string
): Promise<SitemapUrl[]> => {
  const limit = createLimiter(GLOBAL_CONCURRENCY)

  const apiGet = (path: string): Promise<unknown | null> =>
    limit(async () => {
      try {
        return await $fetch(`${apiBase}/api${path}`, {
          headers: { cookie: SFW_COOKIE, accept: 'application/json' },
          timeout: 15000
        })
      } catch {
        return null
      }
    })

  const withPage = (path: string, page: number) => {
    const sep = path.includes('?') ? '&' : '?'
    return `${path}${sep}page=${page}&limit=${PAGE_SIZE}`
  }

  const fetchPage = async (src: PagedSource, page: number) => {
    const body = await apiGet(withPage(src.path, page))
    return body ? src.pick(unwrap(body)) : []
  }

  const toUrls = (
    src: PagedSource,
    rows: Record<string, unknown>[]
  ): SitemapUrl[] =>
    rows.map((row) => ({
      loc: src.loc(row),
      lastmod: src.lastmod?.(row),
      changefreq: 'daily',
      priority: src.priority
    }))

  const collect = async (src: PagedSource): Promise<SitemapUrl[]> => {
    const firstBody = await apiGet(withPage(src.path, 1))
    if (!firstBody) return []
    const firstRows = src.pick(unwrap(firstBody))
    const urls = toUrls(src, firstRows)

    if (src.total) {
      const total = src.total(unwrap(firstBody))
      const pages =
        typeof total === 'number' && total > 0
          ? Math.min(Math.ceil(total / PAGE_SIZE), MAX_PAGES)
          : MAX_PAGES
      const rest = await Promise.all(
        range(2, pages).map((p) =>
          fetchPage(src, p).then((r) => toUrls(src, r))
        )
      )
      for (const r of rest) urls.push(...r)
      return urls
    }

    for (let start = 2; start <= MAX_PAGES; start += GLOBAL_CONCURRENCY) {
      const batch = await Promise.all(
        range(start, Math.min(start + GLOBAL_CONCURRENCY - 1, MAX_PAGES)).map(
          (p) => fetchPage(src, p)
        )
      )
      let reachedEnd = false
      for (const rows of batch) {
        urls.push(...toUrls(src, rows))
        if (rows.length === 0) reachedEnd = true
      }
      if (reachedEnd) break
    }
    return urls
  }

  const collectSingle = async (
    path: string,
    pick: (data: unknown) => Record<string, unknown>[],
    loc: (row: Record<string, unknown>) => string,
    priority: number
  ): Promise<SitemapUrl[]> => {
    const body = await apiGet(path)
    if (!body) return []
    return pick(unwrap(body)).map((row) => ({
      loc: loc(row),
      changefreq: 'daily',
      priority
    }))
  }

  const num = (row: Record<string, unknown>, key: string) => row[key] as number

  const paged: PagedSource[] = [
    {
      path: '/topic',
      pick: (d) => (Array.isArray(d) ? (d as Record<string, unknown>[]) : []),
      loc: (r) => `/topic/${num(r, 'id')}`,
      lastmod: (r) =>
        toIso((r as Pick<TopicCard, 'status_update_time'>).status_update_time),
      priority: 0.8
    },
    {
      path: '/galgame?indexed=true',
      pick: (d) =>
        ((d as { galgames?: [] })?.galgames ?? []) as Record<string, unknown>[],
      total: (d) => (d as { total?: number })?.total,
      loc: (r) => `/galgame/${num(r, 'id')}`,
      lastmod: (r) =>
        toIso(
          (r as Pick<GalgameCard, 'resource_update_time'>).resource_update_time
        ),
      priority: 0.8
    },
    {
      path: '/galgame-resource',
      pick: (d) =>
        ((d as { resources?: [] })?.resources ?? []) as Record<
          string,
          unknown
        >[],
      total: (d) => (d as { total?: number })?.total,
      loc: (r) => `/galgame/resource/${num(r, 'id')}`,
      lastmod: (r) => toIso(r.edited ?? r.created),
      priority: 0.6
    },
    {
      path: '/galgame-rating/all',
      pick: (d) =>
        ((d as { rating_data?: [] })?.rating_data ?? []) as Record<
          string,
          unknown
        >[],
      total: (d) => (d as { total?: number })?.total,
      loc: (r) => `/galgame-rating/${num(r, 'id')}`,
      lastmod: (r) => toIso(r.updated ?? r.created),
      priority: 0.6
    },
    {
      path: '/galgame-official',
      pick: (d) =>
        ((d as { officials?: [] })?.officials ?? []) as Record<
          string,
          unknown
        >[],
      total: (d) => (d as { total?: number })?.total,
      loc: (r) => `/galgame/official/${num(r, 'id')}`,
      priority: 0.5
    },
    {
      path: '/galgame-tag',
      pick: (d) =>
        ((d as { tags?: [] })?.tags ?? []) as Record<string, unknown>[],
      total: (d) => (d as { total?: number })?.total,
      loc: (r) => `/galgame/tag/${num(r, 'id')}`,
      priority: 0.5
    }
  ]

  const groups = await Promise.all([
    ...paged.map((src) => collect(src)),
    collectSingle(
      '/galgame-engine',
      (d) => (Array.isArray(d) ? (d as Record<string, unknown>[]) : []),
      (r) => `/galgame/engine/${num(r, 'id')}`,
      0.5
    )
  ])

  const seen = new Set<string>()
  const out: SitemapUrl[] = []
  for (const url of groups.flat()) {
    if (seen.has(url.loc)) continue
    seen.add(url.loc)
    out.push(url)
  }
  return out
}
