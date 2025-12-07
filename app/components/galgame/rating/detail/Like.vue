<script setup lang="ts">
const props = defineProps<{
  ratingId?: number
  targetUserId: number
  likeCount: number
  isLiked: boolean
}>()

const { id } = usePersistUserStore()
const isLiked = ref(id && props.isLiked)
const likeCount = ref(props.likeCount)

const toggleLike = async () => {
  const res = await $fetch(`/api/galgame-rating/${props.ratingId}/like`, {
    method: 'PUT',
    body: { galgameRatingId: props.ratingId },
    watch: false,
    ...kungalgameResponseHandler
  })

  if (res) {
    likeCount.value += isLiked.value ? -1 : 1

    if (!isLiked.value) {
      useMessage(10233, 'success')
    } else {
      useMessage(10234, 'success')
    }

    isLiked.value = !isLiked.value
  }
}

const handleClickLikeThrottled = throttle(toggleLike, 1007, () =>
  useMessage(10227, 'warn')
)

const handleClickLike = () => {
  if (!id) {
    useMessage(10235, 'warn', 5000)
    return
  }
  if (id === props.targetUserId) {
    useMessage(10236, 'warn')
    return
  }
  handleClickLikeThrottled()
}
</script>

<template>
  <KunButton
    :is-icon-only="true"
    :variant="isLiked ? 'flat' : 'light'"
    :color="isLiked ? 'secondary' : 'default'"
    size="lg"
    class-name="gap-1"
    @click="handleClickLike"
  >
    <KunIcon name="lucide:thumbs-up" />
    <span class="text-sm">点赞</span>
    <span class="text-sm">{{ likeCount }}</span>
  </KunButton>
</template>
