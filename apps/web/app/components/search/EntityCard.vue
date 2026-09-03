<script setup lang="ts">
import { SEARCH_ENTITY_FAMILY_MAP, searchEntityPath } from './items'

const props = defineProps<{
  item: SearchEntityItem
  keywords?: string
}>()

const meta = computed(() => SEARCH_ENTITY_FAMILY_MAP[props.item.family])
</script>

<template>
  <KunCard :href="searchEntityPath(item)" :is-hoverable="true" padding="sm">
    <div class="flex w-full items-center gap-3">
      <span
        class="bg-default-100 text-default-600 flex size-9 shrink-0 items-center justify-center rounded-lg"
      >
        <KunIcon :name="meta.icon" class="size-4.5" />
      </span>

      <span class="min-w-0 flex-1">
        <span class="block truncate text-sm font-medium">
          <SearchHighlight :text="item.name" :keywords="keywords" />
        </span>
        <span v-if="item.alias" class="text-default-500 block truncate text-xs">
          {{ item.alias }}
        </span>
      </span>
    </div>
  </KunCard>
</template>
