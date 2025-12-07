import type { Prisma } from '~~/prisma/generated/prisma/client'

export const KUN_ADMIN_OVERVIEW_STATS_MODEL_ITEM = [
  'user',
  'topic',
  'topic_reply',
  'topic_comment',
  'galgame',
  'galgame_resource',
  'galgame_comment',
  'galgame_website',
  'galgame_website_comment',
  'chat_message'
  // 'message',
  // 'update_log',
  // 'topic_upvote',
  // 'galgame_history',
  // 'galgame_contributor'
] satisfies Prisma.ModelName[]

export type StatsModelType =
  (typeof KUN_ADMIN_OVERVIEW_STATS_MODEL_ITEM)[number]

export interface ChartItem {
  label: string
  color: string
}

export const KUN_ADMIN_OVERVIEW_STATS_MODEL_MAP: Record<
  StatsModelType,
  ChartItem
> = {
  user: { label: '用户', color: '#006FEE' },
  topic: { label: '话题', color: '#7828C8' },
  topic_reply: { label: '话题回复', color: '#17C964' },
  topic_comment: { label: '话题评论', color: '#F31260' },
  galgame: { label: 'Galgame', color: '#FF4ECD' },
  galgame_resource: { label: 'Galgame 资源', color: '#F5A524' },
  galgame_comment: { label: 'Galgame 评论', color: '#7EE7FC' },
  galgame_website: { label: 'Galgame 网站', color: '#7ccf00' },
  galgame_website_comment: { label: 'Galgame 网站评论', color: '#ff637e' },
  chat_message: { label: '聊天消息', color: '#ff8904' }
  // message: { label: '论坛动态消息', color: '#90a1b9' },
  // update_log: { label: '论坛动态消息', color: '#90a1b9' },
  // topic_upvote: { label: '论坛动态消息', color: '#90a1b9' },
  // galgame_history: { label: '论坛动态消息', color: '#90a1b9' },
  // galgame_contributor: { label: '论坛动态消息', color: '#90a1b9' }
} as const

export const KUN_ADMIN_PAGE_ROUTE = ['overview', 'user', 'setting']

export type KUN_ADMIN_PAGE_ROUTE_TYPE = (typeof KUN_ADMIN_PAGE_ROUTE)[number]

export interface KunAdminPageAsideItem {
  name: KUN_ADMIN_PAGE_ROUTE_TYPE
  label: string
  icon?: string
  router?: KUN_ADMIN_PAGE_ROUTE_TYPE
}

export const KUN_ADMIN_PAGE_ASIDE_NAV_ITEM: KunAdminPageAsideItem[] = [
  {
    name: 'overview',
    label: '数据总览',
    icon: 'lucide:chart-area',
    router: 'overview'
  },
  {
    name: 'user',
    label: '用户管理',
    icon: 'lucide:user',
    router: 'user'
  },
  {
    name: 'setting',
    label: '网站设置',
    icon: 'lucide:settings',
    router: 'setting'
  }
]
