<script setup lang="ts">
const props = defineProps<{
  value?: number
  pending?: boolean
  inline?: boolean
}>()

const label = computed(() =>
  props.value === undefined
    ? ''
    : props.value > 9999
      ? '9999+'
      : `${props.value}`
)
</script>

<template>
  <KunSkeleton
    v-if="value === undefined && pending"
    :class="inline ? '' : 'ml-auto'"
    variant="text"
    width="1.75rem"
    height="0.75rem"
  />
  <span
    v-else-if="label"
    :class="[
      'shrink-0 text-xs tabular-nums',
      value === 0 ? 'text-default-300' : 'text-default-500',
      inline ? '' : 'ml-auto'
    ]"
  >
    {{ label }}
  </span>
</template>
