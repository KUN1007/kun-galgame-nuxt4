export interface SearchCategory {
  value: SearchType
  textValue: string
  icon: string
  /** Measure word + noun, as it reads after 共 N. */
  countUnit: string
}

export const SEARCH_CATEGORIES: SearchCategory[] = [
  {
    value: 'all',
    textValue: '全部',
    icon: 'lucide:layout-grid',
    countUnit: '个结果'
  },
  {
    value: 'topic',
    textValue: '话题',
    icon: 'lucide:message-square-text',
    countUnit: '个话题'
  },
  {
    value: 'galgame',
    textValue: 'Galgame',
    icon: 'lucide:gamepad-2',
    countUnit: '个 Galgame'
  },
  {
    value: 'entity',
    textValue: '资料库',
    icon: 'lucide:library',
    countUnit: '个条目'
  },
  {
    value: 'user',
    textValue: '用户',
    icon: 'lucide:user-round',
    countUnit: '位用户'
  },
  {
    value: 'reply',
    textValue: '回复',
    icon: 'carbon:reply',
    countUnit: '条回复'
  },
  {
    value: 'comment',
    textValue: '评论',
    icon: 'uil:comment-dots',
    countUnit: '条评论'
  },
  {
    value: 'toolset',
    textValue: 'Gal 工具资源',
    icon: 'lucide:wrench',
    countUnit: '个资源'
  }
]

export const SEARCH_CATEGORY_MAP = Object.fromEntries(
  SEARCH_CATEGORIES.map((category) => [category.value, category])
) as Record<SearchType, SearchCategory>

export interface SearchEntityFamilyMeta {
  value: SearchEntityFamily
  textValue: string
  icon: string
  path: string
}

export const SEARCH_ENTITY_FAMILIES: SearchEntityFamilyMeta[] = [
  {
    value: 'character',
    textValue: '角色',
    icon: 'lucide:drama',
    path: '/galgame/character'
  },
  {
    value: 'company',
    textValue: '会社',
    icon: 'lucide:building-2',
    path: '/galgame/official'
  },
  {
    value: 'staff',
    textValue: 'Staff',
    icon: 'lucide:signature',
    path: '/galgame/staff'
  },
  { value: 'tag', textValue: '标签', icon: 'lucide:tag', path: '/galgame/tag' }
]

export const SEARCH_ENTITY_FAMILY_MAP = Object.fromEntries(
  SEARCH_ENTITY_FAMILIES.map((family) => [family.value, family])
) as Record<SearchEntityFamily, SearchEntityFamilyMeta>

export const searchEntityPath = (item: SearchEntityItem) =>
  `${SEARCH_ENTITY_FAMILY_MAP[item.family].path}/${item.id}`
