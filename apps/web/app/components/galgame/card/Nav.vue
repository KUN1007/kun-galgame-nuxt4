<script setup lang="ts">
import { storeToRefs } from 'pinia'
import {
  KUN_GALGAME_RESOURCE_TYPE_MAP,
  KUN_GALGAME_RESOURCE_LANGUAGE_MAP,
  KUN_GALGAME_RESOURCE_PLATFORM_MAP,
  KUN_GALGAME_RESOURCE_SORT_FIELD_MAP
} from '~/constants/galgame'
import {
  KUN_GALGAME_PROVIDER_LABEL_MAP,
  PROVIDER_KEY_OPTIONS,
  type ProviderKey
} from '~/constants/galgameResource'
import { KUN_GALGAME_RATING_GAME_TYPE_MAP } from '~/constants/galgame-rating'
import type {
  KunGalgameResourceTypeOptions,
  KunGalgameResourceLanguageOptions,
  KunGalgameResourcePlatformOptions
} from '~/constants/galgame'

withDefaults(
  defineProps<{
    isShowAdvanced?: boolean
  }>(),
  { isShowAdvanced: false }
)

const {
  page,
  type,
  language,
  platform,
  gameType,
  sortField,
  sortOrder,
  releasedFrom,
  releasedTo,
  releasedMonths,
  includeProviders,
  excludeOnlyProviders,
  minRatingCount,
  minRating
} = useGalgameFilters()

const showFilters = ref(false)
const showDisplay = ref(false)

const {
  showPlatform,
  showRating,
  showViewLike,
  showLanguage,
  showNsfwBadge,
  showPublisher,
  showJapaneseName,
  isOpenInNewTab
} = storeToRefs(usePersistGalgameCardStore())

const { showKUNGalgameNoResource } = storeToRefs(usePersistSettingsStore())

watch(
  () => [
    type.value,
    language.value,
    platform.value,
    gameType.value,
    sortField.value,
    sortOrder.value,
    releasedFrom.value,
    releasedMonths.value,
    includeProviders.value,
    excludeOnlyProviders.value,
    minRatingCount.value,
    minRating.value,
    showKUNGalgameNoResource.value
  ],
  () => {
    page.value = 1
  }
)

const minCountOptions = [
  { value: 0, label: '不限' },
  { value: 5, label: '≥5' },
  { value: 10, label: '≥10' },
  { value: 20, label: '≥20' },
  { value: 50, label: '≥50' }
]
const minRatingOptions = [
  { value: 0, label: '不限' },
  { value: 7, label: '7 分+' },
  { value: 8, label: '8 分+' },
  { value: 9, label: '9 分+' }
]

const csvToSet = (csv: string) => new Set(csv.split(',').filter(Boolean))
const toggleCsv = (csv: string, key: string) => {
  const set = csvToSet(csv)
  if (set.has(key)) {
    set.delete(key)
  } else {
    set.add(key)
  }
  return [...set].sort().join(',')
}

const KUN_RELEASE_EARLIEST_YEAR = 1980
const yearOptions = [
  { value: '', label: '不限' },
  ...Array.from(
    { length: new Date().getFullYear() - KUN_RELEASE_EARLIEST_YEAR + 1 },
    (_, i) => {
      const y = String(new Date().getFullYear() - i)
      return { value: y, label: `${y}` }
    }
  )
]
const monthOptions = Array.from({ length: 12 }, (_, i) => ({
  value: i + 1,
  label: `${i + 1} 月`
}))

const applyFromYear = (year: string) => {
  releasedFrom.value = year
  if (year && releasedTo.value && Number(releasedTo.value) < Number(year)) {
    releasedTo.value = year
  }
}
const applyToYear = (year: string) => {
  releasedTo.value = year
  if (year && releasedFrom.value && Number(releasedFrom.value) > Number(year)) {
    releasedFrom.value = year
  }
}

const selectedMonths = computed(() => csvToSet(releasedMonths.value))
const isMonthSelected = (m: number) => selectedMonths.value.has(String(m))
const toggleMonth = (m: number) => {
  const set = csvToSet(releasedMonths.value)
  const key = String(m)
  if (set.has(key)) {
    set.delete(key)
  } else {
    set.add(key)
  }
  releasedMonths.value = [...set]
    .map(Number)
    .sort((a, b) => a - b)
    .join(',')
}

const includeSet = computed(() => csvToSet(includeProviders.value))
const excludeSet = computed(() => csvToSet(excludeOnlyProviders.value))
const toggleInclude = (key: string) => {
  includeProviders.value = toggleCsv(includeProviders.value, key)
}
const toggleExclude = (key: string) => {
  excludeOnlyProviders.value = toggleCsv(excludeOnlyProviders.value, key)
}

const typeOptions = Object.entries(KUN_GALGAME_RESOURCE_TYPE_MAP)
  .filter(([k]) => k !== 'name')
  .map(([value, label]) => ({ value, label }))

const gameTypeOptions = [
  { value: 'all', label: '全部作品' },
  ...Object.entries(KUN_GALGAME_RATING_GAME_TYPE_MAP).map(([value, label]) => ({
    value,
    label
  })),
  { value: 'uncategorized', label: '未分类' }
]

const langOptions = Object.entries(KUN_GALGAME_RESOURCE_LANGUAGE_MAP).map(
  ([value, label]) => ({ value, label })
)

const platformOptions = Object.entries(KUN_GALGAME_RESOURCE_PLATFORM_MAP)
  .filter(([k]) => k !== 'name')
  .map(([value, label]) => ({ value, label }))

const sortOptions = Object.entries(KUN_GALGAME_RESOURCE_SORT_FIELD_MAP).map(
  ([value, label]) => ({
    value: value === 'views' ? 'view' : value,
    label
  })
)

const hasActiveFilter = computed(
  () =>
    type.value !== 'all' ||
    language.value !== 'all' ||
    platform.value !== 'all' ||
    gameType.value !== 'all' ||
    sortField.value !== 'popularity' ||
    sortOrder.value !== 'desc' ||
    !!releasedFrom.value ||
    !!releasedTo.value ||
    !!releasedMonths.value ||
    !!includeProviders.value ||
    !!excludeOnlyProviders.value ||
    minRatingCount.value > 0 ||
    minRating.value > 0
)

const hasAdvancedFilter = computed(
  () =>
    !!includeProviders.value ||
    !!excludeOnlyProviders.value ||
    !!releasedFrom.value ||
    !!releasedTo.value ||
    !!releasedMonths.value ||
    minRatingCount.value > 0 ||
    minRating.value > 0
)

const resetFilters = () => {
  type.value = 'all'
  language.value = 'all'
  platform.value = 'all'
  gameType.value = 'all'
  sortField.value = 'popularity'
  sortOrder.value = 'desc'
  releasedFrom.value = ''
  releasedTo.value = ''
  releasedMonths.value = ''
  includeProviders.value = ''
  excludeOnlyProviders.value = ''
  minRatingCount.value = 0
  minRating.value = 0
}
</script>

<template>
  <div class="space-y-1">
    <KunScrollShadow>
      <button
        v-for="opt in typeOptions"
        :key="opt.value"
        class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
        :class="
          type === opt.value
            ? 'bg-primary/15 text-primary font-medium'
            : 'text-default-600 hover:bg-default-100'
        "
        @click="type = opt.value as KunGalgameResourceTypeOptions"
      >
        {{ opt.label }}
      </button>
    </KunScrollShadow>

    <KunScrollShadow>
      <button
        v-for="opt in langOptions"
        :key="opt.value"
        class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
        :class="
          language === opt.value
            ? 'bg-primary/15 text-primary font-medium'
            : 'text-default-600 hover:bg-default-100'
        "
        @click="language = opt.value as KunGalgameResourceLanguageOptions"
      >
        {{ opt.label }}
      </button>
    </KunScrollShadow>

    <KunScrollShadow>
      <button
        v-for="opt in platformOptions"
        :key="opt.value"
        class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
        :class="
          platform === opt.value
            ? 'bg-primary/15 text-primary font-medium'
            : 'text-default-600 hover:bg-default-100'
        "
        @click="platform = opt.value as KunGalgameResourcePlatformOptions"
      >
        {{ opt.label }}
      </button>
    </KunScrollShadow>

    <KunScrollShadow>
      <button
        v-for="opt in gameTypeOptions"
        :key="opt.value"
        class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
        :class="
          gameType === opt.value
            ? 'bg-primary/15 text-primary font-medium'
            : 'text-default-600 hover:bg-default-100'
        "
        @click="gameType = opt.value"
      >
        {{ opt.label }}
      </button>
    </KunScrollShadow>

    <KunScrollShadow>
      <button
        v-for="opt in sortOptions"
        :key="opt.value"
        class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
        :class="
          sortField === opt.value
            ? 'bg-primary/15 text-primary font-medium'
            : 'text-default-600 hover:bg-default-100'
        "
        @click="
          sortField = opt.value as
            | 'popularity'
            | 'time'
            | 'view'
            | 'created'
            | 'view_1d'
            | 'view_7d'
            | 'view_30d'
            | 'release_date'
            | 'rating'
        "
      >
        {{ opt.label }}
      </button>
    </KunScrollShadow>

    <div class="flex flex-wrap items-center gap-1.5">
      <button
        class="shrink-0 cursor-pointer rounded-md p-1 transition-colors"
        :class="
          sortOrder === 'desc'
            ? 'bg-primary/15 text-primary'
            : 'text-default-500 hover:bg-default-100'
        "
        @click="sortOrder = 'desc'"
      >
        <KunIcon name="lucide:arrow-down" />
      </button>
      <button
        class="shrink-0 cursor-pointer rounded-md p-1 transition-colors"
        :class="
          sortOrder === 'asc'
            ? 'bg-primary/15 text-primary'
            : 'text-default-500 hover:bg-default-100'
        "
        @click="sortOrder = 'asc'"
      >
        <KunIcon name="lucide:arrow-up" />
      </button>

      <button
        v-if="isShowAdvanced"
        class="text-default-500 hover:text-primary flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-sm transition-colors"
        :class="hasAdvancedFilter && 'text-warning'"
        @click="showFilters = !showFilters"
      >
        <KunIcon name="lucide:sliders-horizontal" class="text-inherit" />
        <span>高级筛选</span>
      </button>

      <button
        v-if="isShowAdvanced"
        class="text-default-500 hover:text-primary flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-sm transition-colors"
        :class="showDisplay && 'text-primary'"
        @click="showDisplay = !showDisplay"
      >
        <KunIcon name="lucide:layout-grid" class="text-inherit" />
        <span>显示设置</span>
      </button>

      <button
        v-if="hasActiveFilter"
        class="text-default-500 hover:text-danger flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-sm transition-colors"
        @click="resetFilters"
      >
        <KunIcon name="lucide:rotate-ccw" class="text-inherit" />
        <span>重置筛选</span>
      </button>
    </div>

    <div
      v-if="showFilters"
      class="bg-default-50 space-y-4 rounded-lg border p-3"
    >
      <div class="text-primary border-b pb-1 text-sm font-semibold">评分</div>

      <div>
        <div class="text-default-700 mb-1.5 text-xs font-medium">
          最低评分人数
          <span class="text-default-400 font-normal">(过滤小样本)</span>
        </div>
        <KunScrollShadow>
          <button
            v-for="opt in minCountOptions"
            :key="`mc-${opt.value}`"
            class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
            :class="
              minRatingCount === opt.value
                ? 'bg-primary/15 text-primary font-medium'
                : 'text-default-600 hover:bg-default-100'
            "
            @click="minRatingCount = opt.value"
          >
            {{ opt.label }}
          </button>
        </KunScrollShadow>
      </div>

      <div>
        <div class="text-default-700 mb-1.5 text-xs font-medium">
          最低评分
          <span class="text-default-400 font-normal">(贝叶斯平滑后)</span>
        </div>
        <KunScrollShadow>
          <button
            v-for="opt in minRatingOptions"
            :key="`mr-${opt.value}`"
            class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
            :class="
              minRating === opt.value
                ? 'bg-primary/15 text-primary font-medium'
                : 'text-default-600 hover:bg-default-100'
            "
            @click="minRating = opt.value"
          >
            {{ opt.label }}
          </button>
        </KunScrollShadow>
      </div>

      <div class="text-primary border-b pb-1 text-sm font-semibold">
        发售日期
      </div>

      <div>
        <div class="text-default-700 mb-1.5 text-xs font-medium">起始年份</div>
        <KunScrollShadow>
          <button
            v-for="opt in yearOptions"
            :key="opt.value || 'from-all'"
            class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
            :class="
              releasedFrom === opt.value
                ? 'bg-primary/15 text-primary font-medium'
                : 'text-default-600 hover:bg-default-100'
            "
            @click="applyFromYear(opt.value)"
          >
            {{ opt.label }}
          </button>
        </KunScrollShadow>
      </div>

      <div>
        <div class="text-default-700 mb-1.5 text-xs font-medium">结束年份</div>
        <KunScrollShadow>
          <button
            v-for="opt in yearOptions"
            :key="opt.value || 'to-all'"
            class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
            :class="
              releasedTo === opt.value
                ? 'bg-primary/15 text-primary font-medium'
                : 'text-default-600 hover:bg-default-100'
            "
            @click="applyToYear(opt.value)"
          >
            {{ opt.label }}
          </button>
        </KunScrollShadow>
      </div>

      <div>
        <div class="text-default-700 mb-1.5 text-xs font-medium">
          发售月份
          <span class="text-default-400 font-normal">(可多选, 含历年)</span>
        </div>
        <KunScrollShadow>
          <button
            v-for="opt in monthOptions"
            :key="opt.value"
            class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
            :class="
              isMonthSelected(opt.value)
                ? 'bg-primary/15 text-primary font-medium'
                : 'text-default-600 hover:bg-default-100'
            "
            @click="toggleMonth(opt.value)"
          >
            {{ opt.label }}
          </button>
        </KunScrollShadow>
      </div>

      <div class="text-primary border-b pb-1 text-sm font-semibold">网盘</div>

      <div>
        <div class="text-default-700 mb-1.5 text-xs font-medium">
          必须含有以下网盘
        </div>
        <KunScrollShadow>
          <button
            v-for="key in PROVIDER_KEY_OPTIONS"
            :key="key"
            class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
            :class="
              includeSet.has(key)
                ? 'bg-primary/15 text-primary font-medium'
                : 'text-default-600 hover:bg-default-100'
            "
            @click="toggleInclude(key)"
          >
            {{ KUN_GALGAME_PROVIDER_LABEL_MAP[key as ProviderKey] }}
          </button>
        </KunScrollShadow>
      </div>

      <div>
        <div class="text-default-700 mb-1.5 text-xs font-medium">
          排除仅含以下网盘
        </div>
        <KunScrollShadow>
          <button
            v-for="key in PROVIDER_KEY_OPTIONS"
            :key="key + '-ex'"
            class="cursor-pointer rounded-md px-2.5 py-1 text-sm whitespace-nowrap transition-colors"
            :class="
              excludeSet.has(key)
                ? 'bg-danger/15 text-danger font-medium'
                : 'text-default-600 hover:bg-default-100'
            "
            @click="toggleExclude(key)"
          >
            {{ KUN_GALGAME_PROVIDER_LABEL_MAP[key as ProviderKey] }}
          </button>
        </KunScrollShadow>
      </div>
    </div>

    <div
      v-if="showDisplay"
      class="bg-default-50 space-y-4 rounded-lg border p-3"
    >
      <div class="text-primary border-b pb-1 text-sm font-semibold">
        卡片四角
      </div>
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <KunSwitch v-model="showPlatform" label="左上角 · 游戏平台" />
        <KunSwitch v-model="showRating" label="右上角 · 总评分" />
        <KunSwitch v-model="showViewLike" label="左下角 · 浏览 / 点赞" />
        <KunSwitch v-model="showLanguage" label="右下角 · 可用语言" />
      </div>

      <div class="text-primary border-b pb-1 text-sm font-semibold">其它</div>
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <KunSwitch v-model="showNsfwBadge" label="显示 NSFW 角标" />
        <KunSwitch v-model="showPublisher" label="底部发布者与时间" />
        <KunSwitch v-model="showJapaneseName" label="标题下显示另一个名称" />
        <KunSwitch v-model="isOpenInNewTab" label="在新页面打开卡片" />
        <KunSwitch
          v-model="showKUNGalgameNoResource"
          label="显示没有下载资源的 Galgame"
        />
      </div>
    </div>
  </div>
</template>
