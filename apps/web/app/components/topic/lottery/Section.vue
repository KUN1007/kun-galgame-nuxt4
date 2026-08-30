<script setup lang="ts">
import { ref } from 'vue'
import { useLottery } from '~/composables/topic/useLottery'

const props = defineProps<{
  topicId: number
  isTopicAdmin: boolean
}>()

const isCreateOpen = defineModel<boolean>('isCreateOpen', { default: false })

const { getLotteries } = useLottery(props.topicId)

const isModalOpen = ref(false)
const lotteryToEdit = ref<TopicLottery | undefined>(undefined)

const { data, refresh } = await getLotteries()
const lotteries = computed(() => data.value || [])

watch(isCreateOpen, (open) => {
  if (!open) {
    return
  }
  lotteryToEdit.value = undefined
  isModalOpen.value = true
  isCreateOpen.value = false
})

const openEditModal = (lottery: TopicLottery) => {
  lotteryToEdit.value = lottery
  isModalOpen.value = true
}
</script>

<template>
  <div class="space-y-3">
    <TopicLotteryCard
      v-for="lottery in lotteries"
      :key="lottery.id"
      :lottery="lottery"
      :is-topic-admin="isTopicAdmin"
      @edit="openEditModal"
      @refresh="refresh"
    />

    <TopicLotteryModal
      v-if="isTopicAdmin"
      v-model="isModalOpen"
      :topic-id="topicId"
      :initial-data="lotteryToEdit"
      @refresh="refresh"
    />
  </div>
</template>
