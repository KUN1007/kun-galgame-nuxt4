<script setup lang="ts">
import { watchDebounced } from '@vueuse/core'

const props = defineProps<{
  keywords: string
}>()

const emit = defineEmits<{
  submit: [value: string]
  remember: [value: string]
}>()

const draft = ref(props.keywords)

watch(
  () => props.keywords,
  (value) => {
    if (value !== draft.value.trim()) {
      draft.value = value
    }
  }
)

watchDebounced(draft, (value) => emit('submit', value.trim()), {
  debounce: 400
})

// Only an explicit Enter is worth remembering. The debounced watcher above sees
// every prefix of what is being typed, so recording there fills the history
// with 汉, 汉化, 汉化补.
const handleEnter = () => {
  const value = draft.value.trim()
  emit('submit', value)
  if (value) {
    emit('remember', value)
  }
}
</script>

<template>
  <KunInput
    v-model="draft"
    type="text"
    size="lg"
    :autofocus="true"
    :is-clearable="true"
    placeholder="搜索 Galgame, 话题, 角色, 会社, 下载资源, 用户, 回复与评论…"
    @keydown.enter="handleEnter"
  >
    <template #prefix>
      <KunIcon name="lucide:search" class="text-default-400 size-5" />
    </template>
  </KunInput>
</template>
