<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  topicId: number
  isTopicAdmin: boolean
}>()

const isPollCreateOpen = ref(false)
const isLotteryCreateOpen = ref(false)

const handleCreate = (key: 'poll' | 'lottery') => {
  if (key === 'poll') {
    isPollCreateOpen.value = true
    return
  }
  isLotteryCreateOpen.value = true
}
</script>

<template>
  <div class="space-y-3">
    <TopicPollSection
      v-model:is-create-open="isPollCreateOpen"
      :topic-id="topicId"
      :is-topic-admin="isTopicAdmin"
    />

    <TopicLotterySection
      v-model:is-create-open="isLotteryCreateOpen"
      :topic-id="topicId"
      :is-topic-admin="isTopicAdmin"
    />

    <TopicMiniappPanel v-if="isTopicAdmin" @create="handleCreate" />
  </div>
</template>
