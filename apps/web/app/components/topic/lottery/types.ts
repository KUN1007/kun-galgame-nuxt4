export interface LotteryPrizeFormData {
  name: string
  description: string
  image_hashes: string[]
  image_urls: string[]
  nsfw_hashes: string[]
  delivery: 'code' | 'manual' | 'point'
  point_mode: TopicLotteryPointMode
  point_amount: number
  slots: number
  codes: string
}

export interface LotteryFormData {
  topic_id: number
  lottery_id: number
  title: string
  description: string
  entry_mode: 'signup' | 'reply' | 'floor'
  floor_rule: string
  draw_mode: 'deadline' | 'manual' | 'threshold'
  draw_threshold: number
  deadline?: string
  min_account_age_days: number
  min_moemoepoint: number
  show_entrants: boolean
  prizes: LotteryPrizeFormData[]
}
