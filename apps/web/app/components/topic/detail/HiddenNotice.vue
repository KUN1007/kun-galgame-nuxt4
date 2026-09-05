<script setup lang="ts">
import { topicHiddenByMeta } from '~/constants/topic'

const props = defineProps<{
  hiddenBy: string
}>()

const meta = computed(() => topicHiddenByMeta(props.hiddenBy))
</script>

<template>
  <KunInfo
    variant="flat"
    icon="lucide:eye-off"
    :color="meta.color"
    :title="meta.notice"
    description="目前只有作者本人与持有「查看隐藏话题」权限的管理人员可以看到它。"
  />
</template>
