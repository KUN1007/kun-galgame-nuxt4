<script setup lang="ts">
import { topicAccessScopeMeta } from '~/constants/topic'

const props = defineProps<{
  scope: string
}>()

const meta = computed(() => topicAccessScopeMeta(props.scope))
</script>

<template>
  <KunInfo
    variant="flat"
    color="default"
    :icon="meta.icon"
    :title="meta.notice"
    :description="meta.description"
  />
</template>
