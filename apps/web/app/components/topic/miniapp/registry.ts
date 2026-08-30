export interface TopicMiniApp {
  key: 'poll' | 'lottery'
  label: string
  icon: string
  help: string[]
}

// The 话题小程序 panel used to hardcode its one button and its help text inside
// the poll component, so the second mini-app could only be added by editing the
// first one. Everything the panel renders now comes from here.
export const TOPIC_MINI_APPS: TopicMiniApp[] = [
  {
    key: 'poll',
    label: '投票',
    icon: 'lucide:bar-chart-3',
    help: [
      '让读者在若干选项里表态, 结果实时统计。',
      '可以设置单选/多选、截止时间、匿名与否, 以及结果什么时候公开。',
      '每个话题最多 30 个投票, 每个投票最多 20 个选项。'
    ]
  },
  {
    key: 'lottery',
    label: '抽奖',
    icon: 'lucide:gift',
    help: [
      '送激活码、实物周边或萌萌点, 由系统产生中奖名单。',
      '开奖前公示随机数承诺 (哈希), 开奖后公布随机数, 任何人都能自己验算一遍。',
      '楼层抽奖不使用随机数, 中奖楼层数出来就是, 读者数楼即可核对。',
      '激活码由系统托管, 只有中奖者本人能揭示一次; 实物请与中奖者私聊沟通。',
      '本站只负责产生名单, 不担保实物履约, 请自行与对方确认。'
    ]
  }
]
