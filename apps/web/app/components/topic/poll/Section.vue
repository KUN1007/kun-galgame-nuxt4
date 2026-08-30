<script setup lang="ts">
import { ref } from 'vue'
import { usePoll } from '~/composables/topic/usePoll'

const props = defineProps<{
  topicId: number
  isTopicAdmin: boolean
}>()

const isCreateOpen = defineModel<boolean>('isCreateOpen', { default: false })

const { getPoll } = usePoll(props.topicId)

const isModalOpen = ref(false)
const pollToEdit = ref<TopicPoll | undefined>(undefined)

const { data, refresh } = await getPoll()
const polls = computed(() => data.value || [])

watch(isCreateOpen, (open) => {
  if (!open) {
    return
  }
  pollToEdit.value = undefined
  isModalOpen.value = true
  isCreateOpen.value = false
})

const openEditModal = (pollData: TopicPoll) => {
  pollToEdit.value = pollData
  isModalOpen.value = true
}
</script>

<template>
  <div class="space-y-3">
    <TopicPollList
      v-for="poll in polls"
      :key="poll.id"
      :poll="poll"
      :is-topic-admin="isTopicAdmin"
      @edit="openEditModal"
      @refresh="refresh"
    />

    <TopicPollModal
      v-if="isTopicAdmin"
      v-model="isModalOpen"
      :topic-id="topicId"
      :initial-data="pollToEdit"
      @refresh="refresh"
    />
  </div>
</template>
