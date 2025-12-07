import {
  KUN_GALGAME_RATING_SPOILER_CONST,
  KUN_GALGAME_RATING_SPOILER_MAP,
  KUN_GALGAME_RATING_PLAY_STATUS_CONST,
  KUN_GALGAME_RATING_PLAY_STATUS_MAP,
  KUN_GALGAME_RATING_GAME_TYPE_CONST,
  KUN_GALGAME_RATING_GAME_TYPE_MAP
} from '~/constants/galgame-rating'

export const spoilerOptions = [
  { value: 'all', label: '剧透程度' },
  ...KUN_GALGAME_RATING_SPOILER_CONST.map((s) => ({
    value: s,
    label: KUN_GALGAME_RATING_SPOILER_MAP[s] || ''
  }))
]
export const playStatusOptions = [
  { value: 'all', label: '进度程度' },
  ...KUN_GALGAME_RATING_PLAY_STATUS_CONST.map((s) => ({
    value: s,
    label: KUN_GALGAME_RATING_PLAY_STATUS_MAP[s] || ''
  }))
]
export const typeOptions = [
  { value: 'all', label: '游戏类型' },
  ...KUN_GALGAME_RATING_GAME_TYPE_CONST.map((t) => ({
    value: t,
    label: KUN_GALGAME_RATING_GAME_TYPE_MAP[t] || ''
  }))
]
export const sortFieldOptions = [
  { value: 'time', label: '最新' },
  { value: 'view', label: '热度' },
  { value: 'overall', label: '评分' }
]
