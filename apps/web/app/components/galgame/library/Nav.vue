<script setup lang="ts">
import { KUN_GALGAME_LIBRARY_SORT_FIELD_MAP } from '~/constants/galgame'

const { page, sortField, sortOrder, releasedFrom, releasedTo } =
  useGalgameFilters('popularity')

const showFilters = ref(false)
const showDisplay = ref(false)

watch(
  () => [
    sortField.value,
    sortOrder.value,
    releasedFrom.value,
    releasedTo.value
  ],
  () => {
    page.value = 1
  }
)

const sortOptions = Object.entries(KUN_GALGAME_LIBRARY_SORT_FIELD_MAP).map(
  ([value, label]) => ({ value, label })
)

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

// Only the release-date sort reads the direction; the catalog's popularity and
// updated cursors are descending and have no ascending counterpart.
const isOrderable = computed(() => sortField.value === 'release_date')

const hasActiveFilter = computed(
  () =>
    sortField.value !== 'popularity' ||
    sortOrder.value !== 'desc' ||
    !!releasedFrom.value ||
    !!releasedTo.value
)

const resetFilters = () => {
  sortField.value = 'popularity'
  sortOrder.value = 'desc'
  releasedFrom.value = ''
  releasedTo.value = ''
}
</script>

<template>
  <div class="space-y-1">
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
        @click="sortField = opt.value as 'popularity' | 'release_date' | 'time'"
      >
        {{ opt.label }}
      </button>
    </KunScrollShadow>

    <div class="flex flex-wrap items-center gap-1.5">
      <template v-if="isOrderable">
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
      </template>

      <button
        class="text-default-500 hover:text-primary flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-sm transition-colors"
        :class="(!!releasedFrom || !!releasedTo) && 'text-warning'"
        @click="showFilters = !showFilters"
      >
        <KunIcon name="lucide:calendar-range" class="text-inherit" />
        <span>发售年份</span>
      </button>

      <button
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
    </div>

    <GalgameCardDisplaySettings v-if="showDisplay" />
  </div>
</template>
