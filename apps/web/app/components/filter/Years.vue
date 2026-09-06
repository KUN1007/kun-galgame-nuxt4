<script setup lang="ts">
const props = withDefaults(
  defineProps<{ from: string; to: string; min?: number }>(),
  { min: 1980 }
)

const emit = defineEmits<{ update: [range: { from: string; to: string }] }>()

const popover = useTemplateRef('popover')

const currentYear = new Date().getFullYear()

const decades = computed(() => {
  const rows: { start: number; years: number[] }[] = []
  for (
    let start = Math.floor(currentYear / 10) * 10;
    start >= props.min;
    start -= 10
  ) {
    const years: number[] = []
    for (let year = start; year <= Math.min(start + 9, currentYear); year++) {
      if (year >= props.min) {
        years.push(year)
      }
    }
    rows.push({ start, years })
  }
  return rows
})

const from = computed(() => Number(props.from) || 0)
const to = computed(() => Number(props.to) || 0)
const active = computed(() => !!from.value || !!to.value)

const summary = computed(() => {
  if (from.value && to.value) {
    return from.value === to.value
      ? `${from.value} 年`
      : `${from.value} - ${to.value}`
  }
  if (from.value) {
    return `${from.value} 年至今`
  }
  if (to.value) {
    return `${to.value} 年以前`
  }
  return ''
})

// The first click picks that one year, the second turns it into a range — so a
// half-finished pick has to be forgotten whenever the panel is toggled, or the
// next visit silently anchors a range on a year the reader no longer sees.
const anchor = ref<number | null>(null)
const hovered = ref<number | null>(null)

const apply = (range: { from: string; to: string }) => {
  anchor.value = null
  emit('update', range)
}

const pick = (year: number) => {
  if (anchor.value === null) {
    anchor.value = year
    emit('update', { from: String(year), to: String(year) })
    return
  }
  const [lo, hi] = [anchor.value, year].sort((a, b) => a - b) as [
    number,
    number
  ]
  apply({ from: String(lo), to: String(hi) })
  popover.value?.close()
}

const recent = (years: number) =>
  apply({ from: String(currentYear - years + 1), to: String(currentYear) })

const inRange = (year: number) => {
  if (anchor.value !== null && hovered.value !== null) {
    const lo = Math.min(anchor.value, hovered.value)
    const hi = Math.max(anchor.value, hovered.value)
    return year >= lo && year <= hi
  }
  if (from.value && to.value) {
    return year >= from.value && year <= to.value
  }
  return year === from.value || year === to.value
}

const isEdge = (year: number) =>
  year === anchor.value || year === from.value || year === to.value

const presets = [
  { label: '今年', years: 1 },
  { label: '近 3 年', years: 3 },
  { label: '近 5 年', years: 5 },
  { label: '近 10 年', years: 10 }
]
</script>

<template>
  <KunPopover
    ref="popover"
    position="bottom-start"
    aria-label="发售年份"
    inner-class="p-2 w-80 max-w-[calc(100vw-2rem)]"
  >
    <template #trigger>
      <FilterTrigger
        icon="lucide:calendar-range"
        label="发售年份"
        :value="summary"
        :active="active"
        @click="anchor = null"
      />
    </template>

    <div class="flex flex-wrap gap-1">
      <button
        v-for="preset in presets"
        :key="preset.label"
        type="button"
        class="text-default-600 hover:bg-default-100 cursor-pointer rounded-md px-2 py-1 text-xs transition-colors"
        @click="recent(preset.years)"
      >
        {{ preset.label }}
      </button>
      <button
        v-if="active"
        type="button"
        class="text-danger hover:bg-danger/10 ml-auto cursor-pointer rounded-md px-2 py-1 text-xs transition-colors"
        @click="apply({ from: '', to: '' })"
      >
        不限
      </button>
    </div>

    <div class="border-default-200 mt-2 space-y-1 border-t pt-2">
      <div
        v-for="row in decades"
        :key="row.start"
        class="grid grid-cols-[2.75rem_repeat(10,minmax(0,1fr))] items-center gap-0.5"
      >
        <button
          type="button"
          class="text-default-400 hover:text-primary cursor-pointer pr-1 text-left text-xs tabular-nums transition-colors"
          @click="
            apply({
              from: String(row.start),
              to: String(Math.min(row.start + 9, currentYear))
            })
          "
        >
          {{ row.start }}s
        </button>
        <button
          v-for="year in row.years"
          :key="year"
          type="button"
          :style="{ gridColumnStart: (year % 10) + 2 }"
          :class="
            cn(
              'cursor-pointer rounded py-1 text-xs tabular-nums transition-colors',
              isEdge(year)
                ? 'bg-primary text-primary-foreground font-medium'
                : inRange(year)
                  ? 'bg-primary/15 text-primary'
                  : 'text-default-600 hover:bg-default-100'
            )
          "
          @click="pick(year)"
          @mouseenter="hovered = year"
          @mouseleave="hovered = null"
        >
          {{ String(year % 100).padStart(2, '0') }}
        </button>
      </div>
    </div>

    <p class="text-default-400 mt-2 text-xs">
      {{
        anchor === null
          ? '点一个年份筛选当年, 再点一个可选择区间'
          : `已选 ${anchor} 年, 再点一个年份选择区间`
      }}
    </p>
  </KunPopover>
</template>
