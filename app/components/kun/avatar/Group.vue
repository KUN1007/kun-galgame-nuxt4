<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    users: KunUser[]
    ellipsis?: boolean
    visibleCount?: number
    total?: number
  }>(),
  { ellipsis: true, visibleCount: 5, total: 0 }
)

const users = computed(() => {
  if (!props.ellipsis) {
    return props.users
  }
  return props.users.slice(0, props.visibleCount)
})
</script>

<template>
  <div class="flex flex-wrap items-center gap-1">
    <KunAvatar v-for="(user, index) in users" :key="index" :user="user" />

    <div
      class="bg-default-500/20 flex items-center justify-center rounded-full px-2 py-2 text-xs"
      v-if="props.total && props.total > props.visibleCount"
    >
      {{ `+ ${props.total - props.visibleCount}` }}
    </div>

    <slot />
  </div>
</template>
