<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    content?: string
    className?: string
  }>(),
  { content: '', className: '' }
)

const processedText = computed(() =>
  props.content.replace(/([_/])/g, '$1\u200B')
)
</script>

<template>
  <span :class="cn('kun-text-block', props.className)">
    {{ processedText }}
  </span>
</template>

<style scoped>
.kun-text-block {
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  word-break: break-word;
}
</style>
