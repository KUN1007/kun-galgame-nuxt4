<script setup lang="ts">
const props = defineProps<{
  text?: string
  keywords?: string
}>()

const escapeRegExp = (value: string) =>
  value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

const segments = computed(() => {
  const text = props.text ?? ''
  const tokens = (props.keywords ?? '').trim().split(/\s+/).filter(Boolean)
  if (!text || !tokens.length) {
    return [{ text, hit: false }]
  }

  const pattern = new RegExp(tokens.map(escapeRegExp).join('|'), 'gi')
  const parts: { text: string; hit: boolean }[] = []
  let cursor = 0
  for (const match of text.matchAll(pattern)) {
    const start = match.index ?? 0
    if (start > cursor) {
      parts.push({ text: text.slice(cursor, start), hit: false })
    }
    parts.push({ text: match[0], hit: true })
    cursor = start + match[0].length
  }
  if (cursor < text.length) {
    parts.push({ text: text.slice(cursor), hit: false })
  }
  return parts
})
</script>

<template>
  <span>
    <template v-for="(segment, index) in segments" :key="index">
      <mark
        v-if="segment.hit"
        class="bg-primary/15 text-primary rounded-sm px-px font-medium"
        >{{ segment.text }}</mark
      >
      <template v-else>{{ segment.text }}</template>
    </template>
  </span>
</template>
