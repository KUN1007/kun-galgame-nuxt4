export interface TopicLotteryPrize {
  id: number
  name: string
  description: string
  image_hash: string
  image_url: string
  delivery: 'code' | 'manual' | 'point'
  point_amount: number
  slots: number
  codes_loaded: number
}

export interface TopicLotteryWinner {
  entry_id: number
  prize_id: number
  prize_name: string
  user: KunUser
  reply_floor: number
  rank_key: string
  fulfillment: string
  won_at: string | Date
}

export interface TopicLottery {
  id: number
  topic_id: number
  title: string
  description: string

  entry_mode: 'signup' | 'reply' | 'floor'
  floor_rule: string
  draw_mode: 'deadline' | 'manual' | 'threshold'
  draw_threshold: number
  deadline: string | Date | null

  min_account_age_days: number
  min_moemoepoint: number
  show_entrants: boolean

  status: 'open' | 'drawing' | 'drawn' | 'cancelled'
  seed_hash: string
  seed: string
  entry_count: number
  total_slots: number
  drawn_at: string | Date | null

  user: KunUser
  prizes: TopicLotteryPrize[]
  winners: TopicLotteryWinner[]

  has_entered: boolean
  can_enter: boolean
  enter_blocked: string

  my_entry_id: number
  my_prize_id: number
  my_prize_name: string
  my_delivery: string
  my_fulfillment: string
  my_code_ready: boolean

  created: string | Date
  updated: string | Date
}

export interface TopicLotteryEntrant {
  user: KunUser
  reply_floor: number
  created: string | Date
}
