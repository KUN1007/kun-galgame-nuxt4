import type { KunSelectOption } from '@kungal/ui-vue'

export const KUN_TOPIC_CATEGORY: Record<string, string> = {
  galgame: 'Galgame',
  technique: '技术交流',
  others: '其它话题'
}

export const KUN_TOPIC_CATEGORY_CONST = [
  'galgame',
  'technique',
  'others',
  'all'
] as const

export const KUN_TOPIC_SECTION: Record<string, string> = {
  'g-walkthrough': '攻略',
  'g-chatting': '闲聊',
  'g-article': '文章',
  'g-seeking': '寻求资源',
  'g-news': '资讯',
  'g-releases': '新作消息',
  'g-other': '其它',
  't-crack': '逆向工程',
  't-web': 'Web',
  't-languages': '编程语言',
  't-help': '请求帮助',
  't-linux': 'Linux',
  't-practical': '实用技术',
  't-ai': 'AI',
  't-android': 'Android',
  't-adobe': 'Adobe',
  't-algorithm': '算法',
  't-other': '其它',
  'o-anime': '动漫',
  'o-comics': '漫画',
  'o-music': '音乐',
  'o-novel': '轻小说',
  'o-daily': '日常',
  'o-essay': '个人随笔',
  'o-forum': '论坛相关',
  'o-patch': '补丁网站',
  'o-other': '其它'
}

export const KUN_TOPIC_SECTION_CONST = [
  'g-walkthrough',
  'g-chatting',
  'g-article',
  'g-seeking',
  'g-news',
  'g-releases',
  'g-other',
  't-crack',
  't-web',
  't-languages',
  't-help',
  't-linux',
  't-practical',
  't-ai',
  't-android',
  't-adobe',
  't-algorithm',
  't-other',
  'o-anime',
  'o-comics',
  'o-music',
  'o-novel',
  'o-daily',
  'o-essay',
  'o-forum',
  'o-patch',
  'o-other'
] as const

export type TopicListSortField =
  | 'status_update_time'
  | 'created'
  | 'view_1d'
  | 'view_7d'
  | 'view_30d'
  | 'view'

export const topicListSortFieldOptions: KunSelectOption<TopicListSortField>[] =
  [
    { value: 'status_update_time', label: '更新时间' },
    { value: 'created', label: '创建时间' },
    { value: 'view_1d', label: '日浏览数' },
    { value: 'view_7d', label: '周浏览数' },
    { value: 'view_30d', label: '月浏览数' },
    { value: 'view', label: '总浏览数' }
  ]

export const TOPIC_SORT_FIELD_CONST = [
  'created',
  'view',
  'view_1d',
  'view_7d',
  'view_30d',
  'status_update_time',
  'like',
  'favorite',
  'upvote'
] as const

export const TOPIC_CATEGORIES = {
  galgame: {
    key: 'galgame',
    label: 'Galgame',
    icon: 'lucide:gamepad-2'
  },
  technique: {
    key: 'technique',
    label: '技术交流',
    icon: 'lucide:drafting-compass'
  },
  others: {
    key: 'others',
    label: '其它话题',
    icon: 'lucide:circle-ellipsis'
  }
} as const

export type TopicCategoryKey = keyof typeof TOPIC_CATEGORIES

export const TOPIC_SECTIONS: Record<
  TopicCategoryKey,
  Record<string, string>
> = {
  galgame: {
    'g-walkthrough': '攻略',
    'g-chatting': '闲聊',
    'g-article': '文章',
    'g-seeking': '寻求资源',
    'g-news': '资讯',
    'g-releases': '新作消息',
    'g-other': '其它'
  },
  technique: {
    't-crack': '逆向工程',
    't-web': 'Web',
    't-languages': '编程语言',
    't-help': '请求帮助',
    't-linux': 'Linux',
    't-practical': '实用技术',
    't-ai': 'AI',
    't-android': 'Android',
    't-adobe': 'Adobe',
    't-algorithm': '算法',
    't-other': '其它'
  },
  others: {
    'o-anime': '动漫',
    'o-comics': '漫画',
    'o-music': '音乐',
    'o-novel': '轻小说',
    'o-daily': '日常',
    'o-essay': '个人随笔',
    'o-forum': '论坛相关',
    'o-patch': '补丁网站',
    'o-other': '其它'
  }
}

export const TOPIC_POLL_VISIBILITY_OPTIONS = [
  { value: 'always', label: '任何人可见结果' },
  { value: 'after_vote', label: '投票后可见结果' },
  { value: 'after_deadline', label: '结束后可见结果' }
] as const

export const TOPIC_POLL_VISIBILITY_MAP: Record<string, string> = {
  always: '任何人可见结果',
  after_vote: '投票后可见结果',
  after_deadline: '结束后可见结果'
}

type TopicHiddenByColor = 'default' | 'warning' | 'danger'

export interface TopicHiddenByMeta {
  label: string
  notice: string
  color: TopicHiddenByColor
}

export const KUN_TOPIC_HIDDEN_BY_FALLBACK: TopicHiddenByMeta = {
  label: '已隐藏',
  notice: '该话题已被隐藏',
  color: 'default'
}

export const KUN_TOPIC_HIDDEN_BY: Record<string, TopicHiddenByMeta> = {
  author: {
    label: '作者隐藏',
    notice: '该话题已被作者隐藏',
    color: 'warning'
  },
  moderator: {
    label: '管理员隐藏',
    notice: '该话题已被管理员隐藏',
    color: 'danger'
  },
  trust: {
    label: '风纪隐藏',
    notice: '该话题已因违规处理被隐藏',
    color: 'danger'
  }
}

export const topicHiddenByMeta = (hiddenBy: string): TopicHiddenByMeta =>
  KUN_TOPIC_HIDDEN_BY[hiddenBy] ?? KUN_TOPIC_HIDDEN_BY_FALLBACK

export const KUN_LOTTERY_ENTRY_MODE: Record<string, string> = {
  signup: '报名参与',
  reply: '回帖后参与',
  floor: '楼层抽奖'
}

export const KUN_LOTTERY_ENTRY_MODE_OPTIONS = [
  { value: 'signup', label: '报名参与 — 点一下就参加' },
  { value: 'reply', label: '回帖后参与 — 需要先在本话题回复' },
  { value: 'floor', label: '楼层抽奖 — 由楼层号直接决定, 无需报名' }
] as const

export const KUN_LOTTERY_DRAW_MODE: Record<string, string> = {
  deadline: '到点自动开奖',
  manual: '楼主手动开奖',
  threshold: '满员自动开奖'
}

export const KUN_LOTTERY_DRAW_MODE_OPTIONS = [
  { value: 'deadline', label: '到点自动开奖' },
  { value: 'manual', label: '楼主手动开奖' },
  { value: 'threshold', label: '满员自动开奖' }
] as const

export const KUN_LOTTERY_DELIVERY: Record<string, string> = {
  code: '系统托管兑换码',
  manual: '楼主私聊发放',
  point: '自动发放萌萌点'
}

export const KUN_LOTTERY_DELIVERY_OPTIONS = [
  { value: 'manual', label: '楼主私聊发放 (实物周边等)' },
  { value: 'code', label: '系统托管兑换码 (激活码等)' },
  { value: 'point', label: '自动发放萌萌点' }
] as const

export const KUN_LOTTERY_POINT_MODE: Record<string, string> = {
  fixed: '每人固定',
  split: '奖池均分',
  random: '奖池拼手气'
}

export const KUN_LOTTERY_POINT_MODE_OPTIONS = [
  { value: 'fixed', label: '每人固定 — 每位中奖者都拿同样多' },
  { value: 'split', label: '奖池均分 — 总额平均分给中奖者' },
  { value: 'random', label: '奖池拼手气 — 总额随机分, 每人至少 1 点' }
] as const

export const KUN_LOTTERY_STATUS: Record<string, string> = {
  open: '进行中',
  drawing: '开奖中',
  drawn: '已开奖',
  cancelled: '已取消'
}

export const KUN_LOTTERY_FULFILLMENT: Record<string, string> = {
  pending: '待发放',
  shipped: '已发出',
  received: '已确认收到',
  forfeited: '已放弃'
}
