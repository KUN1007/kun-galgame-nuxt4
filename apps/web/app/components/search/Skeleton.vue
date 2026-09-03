<script setup lang="ts">
withDefaults(
  defineProps<{
    shape?: 'row' | 'card' | 'entity'
    count?: number
  }>(),
  { shape: 'row', count: 4 }
)
</script>

<template>
  <div
    v-if="shape === 'entity'"
    class="grid gap-2 sm:grid-cols-2 xl:grid-cols-3"
    aria-hidden="true"
  >
    <KunSkeleton
      v-for="i in count * 2"
      :key="i"
      height="3.75rem"
      rounded="lg"
    />
  </div>

  <div
    v-else-if="shape === 'card'"
    class="grid grid-cols-2 gap-4 lg:grid-cols-3"
    aria-hidden="true"
  >
    <KunSkeleton v-for="i in count" :key="i" height="10rem" rounded="lg" />
  </div>

  <div v-else class="space-y-3" aria-hidden="true">
    <KunSkeleton v-for="i in count" :key="i" height="5.5rem" rounded="lg" />
  </div>
</template>
