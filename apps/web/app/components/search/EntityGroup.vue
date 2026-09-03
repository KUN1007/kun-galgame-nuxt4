<script setup lang="ts">
import { SEARCH_ENTITY_FAMILY_MAP } from './items'

const props = defineProps<{
  group: SearchEntityGroup
  keywords: string
  showHeader?: boolean
  showCap?: boolean
}>()

const meta = computed(() => SEARCH_ENTITY_FAMILY_MAP[props.group.family])

// The catalog search face answers one page and carries no cursor, so what the
// limit did not fetch cannot be paged to — say so instead of implying a page 2.
const isCapped = computed(
  () => props.showCap && props.group.total > props.group.items.length
)
</script>

<template>
  <section v-if="group.items.length" class="space-y-3">
    <header v-if="showHeader" class="flex items-center gap-2">
      <KunIcon :name="meta.icon" class="text-default-500 size-4" />
      <h3 class="text-sm font-medium">{{ meta.textValue }}</h3>
      <span class="text-default-400 text-xs tabular-nums">{{
        group.total
      }}</span>
    </header>

    <div class="grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
      <SearchEntityCard
        v-for="item in group.items"
        :key="`${item.family}-${item.id}`"
        :item="item"
        :keywords="keywords"
      />
    </div>

    <p v-if="isCapped" class="text-default-400 text-xs">
      按相关度显示前 {{ group.items.length }} 条, 共 {{ group.total }} 条
    </p>
  </section>
</template>
