import { Feed } from 'feed'

interface KunRSSContent {
  title: string
  description: string
  generator: string
  copyright: string
}

const contentMap: Record<'topic' | 'galgame', KunRSSContent> = {
  topic: {
    title: '鲲 Galgame 论坛 - 新话题订阅',
    description: '最新更新关于 Galgame 的话题',
    generator: '萌萌 RSS 生成器',
    copyright: `版权所有 © ${new Date().getFullYear()} 鲲 Galgame 保留所有权利`
  },
  galgame: {
    title: '鲲 Galgame 论坛 - 新 Galgame 订阅',
    description: '最新更新的 Galgame',
    generator: '萌萌 RSS 生成器',
    copyright: `版权所有 © ${new Date().getFullYear()} 鲲 Galgame 保留所有权利`
  }
}

export const useKunFeed = (baseUrl: string, link: 'galgame' | 'topic') => {
  const rss = `${baseUrl}/rss/${link}.xml`
  const feedContent = contentMap[link]
  return new Feed({
    ...feedContent,
    id: baseUrl,
    link: baseUrl,
    image: `${baseUrl}/kungalgame.webp`,
    favicon: `${baseUrl}/favicon.ico`,
    feedLinks: { rss }
  })
}

interface KunApiResponse<T> {
  code: number
  message: string
  data: T
}

export const fetchKunApi = async <T>(
  path: string,
  query?: Record<string, string | number>
): Promise<T> => {
  const config = useRuntimeConfig()
  const resp = await $fetch<KunApiResponse<T>>(
    `${config.apiBaseUrl}/api${path}`,
    { method: 'GET', query, timeout: 10000 }
  )
  return resp.data
}
