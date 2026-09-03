<script setup lang="ts">
import { SEARCH_CATEGORIES } from './items'

const props = defineProps<{
  modelValue: SearchType
  totals: SearchOverviewTotals | null
  pending: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: SearchType]
}>()

const counts = computed<Record<SearchType, number>>(() => {
  const totals = props.totals
  if (!totals) {
    return {} as Record<SearchType, number>
  }
  const sum = Object.values(totals).reduce((acc, value) => acc + value, 0)
  return { ...totals, all: sum }
})

const countOf = (value: SearchType) => counts.value[value]

const handleSelect = (value: string) =>
  emit('update:modelValue', value as SearchType)
</script>

<template>
  <nav aria-label="搜索分类">
    <div class="hidden lg:block">
      <KunTab
        :items="SEARCH_CATEGORIES"
        :model-value="modelValue"
        orientation="vertical"
        variant="light"
        align="start"
        :full-width="true"
        @update:model-value="handleSelect"
      >
        <template #tab="{ item }">
          <span class="flex w-full items-center gap-2">
            <KunIcon :name="item.icon" class="size-4 shrink-0" />
            <span class="truncate">{{ item.textValue }}</span>
            <SearchNavCount
              :value="countOf(item.value as SearchType)"
              :pending="pending"
            />
          </span>
        </template>
      </KunTab>
    </div>

    <div class="lg:hidden">
      <KunTab
        :items="SEARCH_CATEGORIES"
        :model-value="modelValue"
        variant="light"
        size="sm"
        :scrollable="true"
        @update:model-value="handleSelect"
      >
        <template #tab="{ item }">
          <span class="flex items-center gap-1.5">
            <KunIcon :name="item.icon" class="size-4 shrink-0" />
            <span>{{ item.textValue }}</span>
            <SearchNavCount
              :value="countOf(item.value as SearchType)"
              :pending="pending"
              :inline="true"
            />
          </span>
        </template>
      </KunTab>
    </div>
  </nav>
</template>
