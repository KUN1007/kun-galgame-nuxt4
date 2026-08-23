<script setup lang="ts">
import { useRouteQuery } from '@vueuse/router'
import type { KunTabItem } from '@kungal/ui-vue'

const opts = { mode: 'replace' as const }
const view = useRouteQuery<string>('view', 'month', opts)
const month = useRouteQuery<string>('month', '', opts)
const year = useRouteQuery<string>('year', '', opts)

const tabs: KunTabItem[] = [
  { value: 'month', textValue: '月历' },
  { value: 'upcoming', textValue: '未发售' },
  { value: 'pending', textValue: '年内待定' },
  { value: 'tba', textValue: '发售日期未定' }
]

const isMonthView = computed(
  () =>
    view.value !== 'upcoming' &&
    view.value !== 'pending' &&
    view.value !== 'tba'
)

const monthQuery = computed(() => (month.value ? { month: month.value } : {}))
const yearQuery = computed(() => (year.value ? { year: year.value } : {}))

const {
  data: monthData,
  status: monthStatus,
  refresh: refreshMonth
} = await useKunFetch<GalgameCalendarMonth>('/galgame/calendar', {
  method: 'GET',
  query: monthQuery,
  watch: [month],
  immediate: isMonthView.value,
  server: isMonthView.value
})

const {
  data: upcomingData,
  status: upcomingStatus,
  refresh: refreshUpcoming
} = await useKunFetch<GalgameCalendarUpcoming>('/galgame/calendar/upcoming', {
  method: 'GET',
  immediate: view.value === 'upcoming',
  server: view.value === 'upcoming'
})

const {
  data: pendingData,
  status: pendingStatus,
  refresh: refreshPending
} = await useKunFetch<GalgameCalendarPending>('/galgame/calendar/pending', {
  method: 'GET',
  query: yearQuery,
  watch: [year],
  immediate: view.value === 'pending',
  server: view.value === 'pending'
})

const {
  data: tbaData,
  status: tbaStatus,
  refresh: refreshTba
} = await useKunFetch<GalgameCalendarTBA>('/galgame/calendar/tba', {
  method: 'GET',
  immediate: view.value === 'tba',
  server: view.value === 'tba'
})

watch(view, () => {
  if (isMonthView.value && !monthData.value) {
    refreshMonth()
  } else if (view.value === 'upcoming' && !upcomingData.value) {
    refreshUpcoming()
  } else if (view.value === 'pending' && !pendingData.value) {
    refreshPending()
  } else if (view.value === 'tba' && !tbaData.value) {
    refreshTba()
  }
})

const ymLabel = (m: string) => {
  const [y, mo] = m.split('-')
  return `${y} 年 ${Number(mo)} 月`
}

const addMonth = (ym: string, delta: number) => {
  const [y, m] = ym.split('-').map(Number)
  const total = (y ?? 0) * 12 + ((m ?? 1) - 1) + delta
  return `${String(Math.floor(total / 12)).padStart(4, '0')}-${String(
    (total % 12) + 1
  ).padStart(2, '0')}`
}

const canGoPrev = computed(
  () =>
    !!monthData.value && monthData.value.month > monthData.value.meta.min_month
)

const goPrevMonth = () => {
  if (canGoPrev.value && monthData.value) {
    month.value = addMonth(monthData.value.month, -1)
  }
}
const goNextMonth = () => {
  if (monthData.value) {
    month.value = addMonth(monthData.value.month, 1)
  }
}
const goToday = () => {
  month.value = ''
}

const goPrevYear = () => {
  const y = Number(pendingData.value?.year)
  if (y) {
    year.value = String(y - 1)
  }
}
const goNextYear = () => {
  const y = Number(pendingData.value?.year)
  if (y) {
    year.value = String(y + 1)
  }
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <KunHeader
      name="Galgame 发售月历"
      description="按发售月份浏览即将与已发售的 Galgame。数据来自 Galgame 资料库, 月份边界按日本时间 (JST) 计。打开任意作品即可发布资源。"
    />

    <KunTab :items="tabs" v-model="view" variant="bordered" />

    <template v-if="view === 'upcoming'">
      <div
        class="border-default-200 flex items-center justify-center gap-2 rounded-xl border px-3 py-3"
      >
        <KunIcon name="lucide:calendar-clock" class="text-default-400 size-5" />
        <span class="font-medium">未发售 · 已定档排期</span>
        <span class="text-default-400 text-sm">
          共 {{ upcomingData?.count ?? 0 }} 部
        </span>
      </div>

      <KunLoading :loading="upcomingStatus === 'pending'">
        <template v-if="upcomingData">
          <div v-if="upcomingData.months.length" class="flex flex-col gap-6">
            <section
              v-for="grp in upcomingData.months"
              :key="grp.month"
              class="flex flex-col gap-2"
            >
              <div class="flex items-center gap-2">
                <KunIcon
                  name="lucide:calendar"
                  class="text-default-500 size-4"
                />
                <h3 class="font-medium">{{ ymLabel(grp.month) }}</h3>
                <span class="text-default-400 text-sm">
                  {{ grp.items.length }} 部
                </span>
              </div>
              <GalgameCard :galgames="grp.items" />
            </section>
          </div>
          <KunNull v-else description="暂无已定档的未发售作品" />
        </template>
      </KunLoading>
    </template>

    <template v-else-if="view === 'pending'">
      <div
        class="border-default-200 flex items-center justify-between gap-2 rounded-xl border px-3 py-2"
      >
        <KunButton variant="light" :is-icon-only="true" @click="goPrevYear">
          <KunIcon name="lucide:chevron-left" class="size-5" />
        </KunButton>
        <div class="flex flex-col items-center">
          <span class="text-xl font-bold sm:text-2xl">
            {{ pendingData?.year }} 年
          </span>
          <span class="text-default-400 text-xs">
            仅知年份 · 月份待定 · 共 {{ pendingData?.count ?? 0 }} 部
          </span>
        </div>
        <KunButton variant="light" :is-icon-only="true" @click="goNextYear">
          <KunIcon name="lucide:chevron-right" class="size-5" />
        </KunButton>
      </div>

      <KunLoading :loading="pendingStatus === 'pending'">
        <template v-if="pendingData">
          <GalgameCard
            v-if="pendingData.items.length"
            :galgames="pendingData.items"
          />
          <KunNull v-else description="该年暂无仅知年份的待定作品" />
        </template>
      </KunLoading>
    </template>

    <template v-else-if="view === 'tba'">
      <div
        class="border-default-200 flex items-center justify-center gap-2 rounded-xl border px-3 py-3"
      >
        <KunIcon name="lucide:calendar-x" class="text-default-400 size-5" />
        <span class="font-medium">发售日期未定</span>
        <span class="text-default-400 text-sm">
          共 {{ tbaData?.count ?? 0 }} 部
        </span>
      </div>

      <KunLoading :loading="tbaStatus === 'pending'">
        <template v-if="tbaData">
          <GalgameCard v-if="tbaData.items.length" :galgames="tbaData.items" />
          <KunNull v-else description="暂无发售日期未定的作品" />
        </template>
      </KunLoading>
    </template>

    <template v-else>
      <KunLoading :loading="monthStatus === 'pending'">
        <GalgameCalendarMonth
          v-if="monthData"
          :data="monthData"
          @prev="goPrevMonth"
          @next="goNextMonth"
          @today="goToday"
        />
      </KunLoading>
    </template>
  </div>
</template>
