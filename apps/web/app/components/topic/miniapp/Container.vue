<script setup lang="ts">
import { reactive } from 'vue'
import { TOPIC_MINI_APPS, type TopicMiniAppKey } from './registry'
import { TOPIC_MINI_APP_SECTIONS } from './sections'

defineProps<{
  topicId: number
  isTopicAdmin: boolean
}>()

const createOpen = reactive(
  Object.fromEntries(TOPIC_MINI_APPS.map((app) => [app.key, false]))
) as Record<TopicMiniAppKey, boolean>
</script>

<template>
  <div class="space-y-3">
    <component
      :is="TOPIC_MINI_APP_SECTIONS[app.key]"
      v-for="app in TOPIC_MINI_APPS"
      :key="app.key"
      v-model:is-create-open="createOpen[app.key]"
      :topic-id="topicId"
      :is-topic-admin="isTopicAdmin"
    />

    <TopicMiniappPanel
      v-if="isTopicAdmin"
      @create="(key: TopicMiniAppKey) => (createOpen[key] = true)"
    />
  </div>
</template>
