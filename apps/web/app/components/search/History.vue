<script setup lang="ts">
const emit = defineEmits<{
  select: [value: string]
}>()

const { searchHistory } = storeToRefs(usePersistKUNGalgameSearchStore())

const recent = computed(() => searchHistory.value.slice().reverse())

const handleDelete = (value: string) => {
  searchHistory.value = searchHistory.value.filter((item) => item !== value)
}
</script>

<template>
  <div class="space-y-3">
    <div class="flex items-center justify-between">
      <h2 class="text-default-700 flex items-center gap-2 text-sm font-medium">
        <KunIcon name="lucide:history" class="size-4" />
        搜索历史
      </h2>
      <KunButton
        v-if="recent.length"
        variant="light"
        size="sm"
        color="default"
        @click="searchHistory = []"
      >
        清除全部
      </KunButton>
    </div>

    <div v-if="recent.length" class="flex flex-wrap gap-2">
      <div
        v-for="value in recent"
        :key="value"
        class="border-default-200 bg-default-50 hover:border-primary/50 hover:bg-primary/5 group flex items-center gap-1 rounded-full border py-1 pr-1 pl-3 text-sm transition-colors"
      >
        <button
          type="button"
          class="max-w-60 truncate"
          @click="emit('select', value)"
        >
          {{ value }}
        </button>
        <button
          type="button"
          class="text-default-400 hover:text-danger rounded-full p-1 transition-colors"
          :aria-label="`删除搜索历史 ${value}`"
          @click="handleDelete(value)"
        >
          <KunIcon name="lucide:x" class="size-3.5" />
        </button>
      </div>
    </div>

    <KunNull v-else description="还没有搜索历史" />

    <p class="text-default-400 text-sm">
      输入关键字即可搜索整个论坛, 在任何页面按 Ctrl / ⌘ + K 也能打开快速搜索。
    </p>
  </div>
</template>
