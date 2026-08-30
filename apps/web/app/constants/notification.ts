export interface NotificationCategory {
  key: string
  label: string
}

export interface NotificationCategoryGroup {
  value: string
  textValue: string
  icon: string
  stream: 'local' | 'chat'
  items: NotificationCategory[]
}

export const notificationCategoryGroups: NotificationCategoryGroup[] = [
  {
    value: 'interaction',
    textValue: '互动',
    icon: 'lucide:heart',
    stream: 'local',
    items: [
      { key: 'upvoted', label: '被推荐' },
      { key: 'liked', label: '被点赞' },
      { key: 'favorite', label: '被收藏' },
      { key: 'mentioned', label: '被 @ 提及' }
    ]
  },
  {
    value: 'reply',
    textValue: '回复评论',
    icon: 'lucide:message-circle',
    stream: 'local',
    items: [
      { key: 'replied', label: '收到回复' },
      { key: 'commented', label: '收到评论' },
      { key: 'solution', label: '回复被采纳为最佳答案' },
      { key: 'pin-reply', label: '回复被置顶' },
      { key: 'quiz-answered', label: '题目被回答' }
    ]
  },
  {
    value: 'review',
    textValue: '内容审核',
    icon: 'lucide:git-pull-request',
    stream: 'local',
    items: [
      { key: 'requested', label: '收到更新请求' },
      { key: 'merged', label: '更新请求被合并' },
      { key: 'declined', label: '更新请求被拒绝' },
      { key: 'expired', label: '资源链接被报告过期' }
    ]
  },
  {
    value: 'miniapp',
    textValue: '话题小程序',
    icon: 'lucide:gift',
    stream: 'local',
    items: [
      { key: 'lottery-won', label: '抽奖中奖' },
      { key: 'lottery-closed', label: '参与的抽奖开奖' },
      { key: 'poll-closed', label: '参与的投票截止' }
    ]
  },
  {
    value: 'chat',
    textValue: '私信',
    icon: 'lucide:mail',
    stream: 'chat',
    items: [{ key: 'chat', label: '私信消息' }]
  }
]

export const localNotificationCategories: NotificationCategory[] =
  notificationCategoryGroups
    .filter((g) => g.stream === 'local')
    .flatMap((g) => g.items)
