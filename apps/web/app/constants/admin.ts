import type { ForumPermission } from '~/composables/useCan'

export const KUN_ADMIN_OVERVIEW_STATS_MODEL_ITEM = [
  'topic',
  'topic_reply',
  'topic_comment',
  'galgame',
  'galgame_resource',
  'galgame_comment',
  'galgame_website',
  'galgame_website_comment',
  'chat_message'
] as const

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
  topic: { label: '话题', color: '#7828C8' },
  topic_reply: { label: '话题回复', color: '#17C964' },
  topic_comment: { label: '话题评论', color: '#F31260' },
  galgame: { label: 'Galgame', color: '#FF4ECD' },
  galgame_resource: { label: 'Galgame 资源', color: '#F5A524' },
  galgame_comment: { label: 'Galgame 评论', color: '#7EE7FC' },
  galgame_website: { label: 'Galgame 网站', color: '#7ccf00' },
  galgame_website_comment: { label: 'Galgame 网站评论', color: '#ff637e' },
  chat_message: { label: '聊天消息', color: '#ff8904' }
} as const

export const KUN_ADMIN_PAGE_ROUTE = [
  'overview',
  'user',
  'submissions',
  'moderation',
  'friend-link',
  'website',
  'doc',
  'update',
  'permission',
  'setting'
]

export type KUN_ADMIN_PAGE_ROUTE_TYPE = (typeof KUN_ADMIN_PAGE_ROUTE)[number]

export interface KunAdminPageAsideItem {
  name: KUN_ADMIN_PAGE_ROUTE_TYPE
  label: string
  icon?: string
  router?: KUN_ADMIN_PAGE_ROUTE_TYPE
  permissions?: ForumPermission[]
  role?: 'admin'
  to?: string
}

export const KUN_ADMIN_PAGE_ASIDE_NAV_ITEM: KunAdminPageAsideItem[] = [
  {
    name: 'overview',
    label: '数据总览',
    icon: 'lucide:chart-area',
    router: 'overview',
    permissions: ['admin.dashboard']
  },
  {
    name: 'user',
    label: '用户管理',
    icon: 'lucide:user',
    router: 'user',
    permissions: ['user.purge_content']
  },
  {
    name: 'submissions',
    label: 'Galgame 审核',
    icon: 'lucide:clipboard-check',
    router: 'submissions',
    permissions: ['galgame.claim.review']
  },
  {
    name: 'moderation',
    label: '内容审核',
    icon: 'lucide:shield-alert',
    router: 'moderation',
    permissions: ['trust.review']
  },
  {
    name: 'friend-link',
    label: '友链管理',
    icon: 'lucide:link',
    router: 'friend-link',
    permissions: [
      'friend_link.create',
      'friend_link.edit',
      'friend_link.delete'
    ]
  },
  {
    name: 'website',
    label: '网站资料库管理',
    icon: 'lucide:globe',
    router: 'website',
    permissions: ['website.create', 'website.edit', 'website.delete']
  },
  {
    name: 'doc',
    label: '文档管理',
    icon: 'lucide:file-text',
    router: 'doc',
    permissions: ['doc.create', 'doc.edit', 'doc.delete']
  },
  {
    name: 'update',
    label: '更新日志与待办',
    icon: 'lucide:list-checks',
    to: '/update/todo',
    permissions: ['update_log.create', 'update_log.edit', 'update_log.delete']
  },
  {
    name: 'permission',
    label: '权限管理',
    icon: 'lucide:shield-check',
    router: 'permission',
    role: 'admin'
  },
  {
    name: 'setting',
    label: '网站设置',
    icon: 'lucide:settings',
    router: 'setting',
    role: 'admin'
  }
]
