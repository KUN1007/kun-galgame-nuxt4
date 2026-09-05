<script setup lang="ts">
const props = defineProps<{
  topicId: number
  status: number
  hiddenBy: string
}>()

const { id } = usePersistUserStore()
const canHideTopic = useCan('topic.hide')
const topicUserId = inject<number>('topicUserId')

const isAuthor = computed(() => !!id && topicUserId === id)
const isHidden = computed(() => props.status === 1)

type HideMode = 'hide' | 'unhide' | 'blocked' | 'none'

const mode = computed<HideMode>(() => {
  if (!isHidden.value) {
    return isAuthor.value || canHideTopic.value ? 'hide' : 'none'
  }
  if (canHideTopic.value || (isAuthor.value && props.hiddenBy === 'author')) {
    return 'unhide'
  }
  return isAuthor.value ? 'blocked' : 'none'
})

const confirmCopy = computed(() => {
  if (mode.value === 'unhide') {
    return {
      title: '确认取消隐藏这个话题吗',
      message: '取消隐藏后, 这个话题会重新回到列表与搜索中, 对所有人可见。'
    }
  }
  if (isAuthor.value) {
    return {
      title: '八嘎杂鱼笨蛋萝莉, 你要隐藏这个话题吗',
      message:
        '隐藏后话题会从列表与搜索中消失, 但您自己以及持有「查看隐藏话题」权限的管理人员仍然看得到它。您可以随时在个人主页的「已隐藏」中取消隐藏。'
    }
  }
  return {
    title: '确认隐藏这个话题吗',
    message:
      '隐藏后话题会从列表与搜索中消失, 仅作者本人与持有「查看隐藏话题」权限的管理人员仍可访问, 并且作者无法自行取消隐藏。'
  }
})

const isPending = ref(false)

const handleUpdateTopicHideStatus = async () => {
  const copy = confirmCopy.value
  const confirmed = await useComponentMessageStore().alert(
    copy.title,
    copy.message
  )
  if (!confirmed) {
    return
  }

  const wasHidden = isHidden.value
  isPending.value = true
  const result = await kunFetch<string>(`/topic/${props.topicId}/hide`, {
    method: 'PUT'
  })
  isPending.value = false

  if (result) {
    useMessage(wasHidden ? '取消隐藏话题成功' : '隐藏话题成功', 'success')
    await refreshNuxtData(`topic-detail-${props.topicId}`)
  }
}
</script>

<template>
  <KunButton
    v-if="mode === 'hide' || mode === 'unhide'"
    variant="light"
    :color="mode === 'unhide' ? 'primary' : 'danger'"
    size="sm"
    :loading="isPending"
    :disabled="isPending"
    @click="handleUpdateTopicHideStatus"
    class-name="whitespace-nowrap gap-2 justify-start"
  >
    <KunIcon
      class-name="text-lg"
      :name="mode === 'unhide' ? 'lucide:eye' : 'lucide:eye-off'"
    />
    {{ mode === 'unhide' ? '取消隐藏该话题' : '隐藏该话题' }}
  </KunButton>

  <div
    v-else-if="mode === 'blocked'"
    class="text-default-500 flex items-start gap-2 px-2 py-1 text-xs"
  >
    <KunIcon class-name="text-base shrink-0" name="lucide:shield-alert" />
    <span>该话题已被管理员隐藏, 无法自行取消</span>
  </div>
</template>
