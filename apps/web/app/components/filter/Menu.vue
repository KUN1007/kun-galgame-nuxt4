<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    icon: string
    label: string
    options: FilterOption[]
    modelValue: string | string[]
    multiple?: boolean
    /** The single-select value that means 不限, where the API spells it 'all'. */
    emptyValue?: string
    /** Lay the options out as a grid instead of a list; 发售月份 uses three. */
    columns?: number
  }>(),
  { multiple: false, emptyValue: '', columns: 0 }
)

defineEmits<{ 'update:modelValue': [value: string | string[]] }>()

// A single-select dimension always holds a value, so what makes it "filtered"
// is holding something other than its default — 排序 is never off, it is only
// ever back on 资源更新时间.
const active = computed(() =>
  props.multiple
    ? (props.modelValue as string[]).length > 0
    : (props.modelValue as string) !== props.emptyValue
)

// Tailwind scans source text, so `grid-cols-${columns}` is a class that never
// gets generated — the panel silently stays one column.
const GRID_CLASS: Record<number, string> = {
  2: 'grid grid-cols-2',
  3: 'grid grid-cols-3',
  4: 'grid grid-cols-4'
}

const classNames = computed(() => ({
  trigger: filterPillClass(active.value),
  list: GRID_CLASS[props.columns] ?? ''
}))
</script>

<template>
  <KunSelect
    :model-value="modelValue"
    :options="options"
    :multiple="multiple"
    :icon="icon"
    :placeholder="label"
    :aria-label="label"
    :full-width="false"
    :max-visible-tags="0"
    :class-names="classNames"
    popup-width="auto"
    rounded="full"
    size="sm"
    @update:model-value="
      $emit('update:modelValue', $event as string | string[])
    "
  >
    <template #option="{ option }">
      <span class="min-w-0 flex-1 truncate">{{ option.label }}</span>
      <span v-if="option.hint" class="text-default-400 shrink-0 text-xs">
        {{ option.hint }}
      </span>
    </template>
  </KunSelect>
</template>
