<script setup lang="ts">
import { SEARCH_CATEGORY_MAP } from './items'

const props = defineProps<{
  type: SearchType
  total: number
  shown: number
}>()

const emit = defineEmits<{
  open: [value: SearchType]
}>()

const meta = computed(() => SEARCH_CATEGORY_MAP[props.type])
const hasMore = computed(() => props.total > props.shown)
</script>

<template>
  <section class="space-y-4">
    <header class="border-default-200 flex items-center gap-2 border-b pb-2">
      <KunIcon :name="meta.icon" class="text-primary size-4.5 shrink-0" />
      <h2 class="font-medium">{{ meta.textValue }}</h2>
      <span class="text-default-400 text-sm tabular-nums">{{ total }}</span>

      <KunButton
        v-if="hasMore"
        class-name="ml-auto"
        variant="light"
        size="sm"
        color="default"
        @click="emit('open', type)"
      >
        查看全部
        <KunIcon name="lucide:chevron-right" class="size-4" />
      </KunButton>
    </header>

    <slot />
  </section>
</template>
