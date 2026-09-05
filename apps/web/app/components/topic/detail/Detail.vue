<script setup lang="ts">
import { useTopicReplies } from '~/composables/topic/useTopicReplies'
import { useTopicScroll } from '~/composables/topic/useTopicScroll'
import { TOPIC_TOC_SOURCE } from '~/composables/topic/useTopicTOC'

const props = defineProps<{
  topic: TopicDetail
}>()

const { id } = usePersistUserStore()
const canCreateAnyPoll = useCan('poll.create_any')
const canEditAnyPoll = useCan('poll.edit_any')
const tempReplyStore = useTempReplyStore()
const { lastSuccessfulReply } = storeToRefs(tempReplyStore)
const isTopicAdmin = computed(
  () =>
    props.topic.user.id === id || canCreateAnyPoll.value || canEditAnyPoll.value
)

const {
  replies,
  status,
  isComplete,
  hasEarlier,
  sortOrder,
  loadInitialReplies,
  loadMore,
  loadEarlier,
  setSort,
  addNewReply,
  updateReply,
  removeReply
} = useTopicReplies(props.topic.id)

const route = useRoute()
const { scrollToFloor, scrollToComment } = useTopicScroll()

const targetFloor = Number(route.query.reply) || 0
const targetCommentId = Number(route.query.comment) || 0

let startPage = 1
let targetReplyFloor = targetFloor
if (targetFloor > 0 || targetCommentId > 0) {
  const located = await kunFetch<{
    page: number
    floor: number
    reply_id: number
    comment_id: number
  }>(`/topic/${props.topic.id}/reply/locate`, {
    query:
      targetCommentId > 0
        ? { comment: targetCommentId }
        : { reply: targetFloor }
  })
  if (located?.page) {
    startPage = located.page
  }
  if (located?.floor) {
    targetReplyFloor = located.floor
  }
}

const activeFloor = ref(targetReplyFloor)
watch(
  () => route.query.reply,
  (value) => {
    const floor = Number(value) || 0
    if (floor > 0) {
      activeFloor.value = floor
      nextTick(() => scrollToFloor(floor, false))
    }
  }
)

await loadInitialReplies(startPage)

onMounted(() => {
  if (!targetFloor && !targetCommentId) {
    return
  }
  setTimeout(() => {
    const ok =
      targetCommentId > 0
        ? scrollToComment(targetCommentId, false)
        : scrollToFloor(targetFloor, false)
    if (!ok) {
      useMessage('目标回复或评论可能已被删除', 'info')
    }
  }, 300)
})

provide('topicUserId', props.topic.user.id)
provide('activeReplyFloor', activeFloor)

provide(TOPIC_TOC_SOURCE, {
  getContentHtml: () => props.topic.content_html,
  getReplies: () => replies.value,
  getTargetFloor: () => activeFloor.value
})

watch(
  lastSuccessfulReply,
  (event) => {
    if (!event) {
      return
    }

    switch (event.type) {
      case 'created':
        addNewReply(event.data)

        nextTick(() => {
          scrollToFloor(event.data.floor)
        })
        break
      case 'updated':
        updateReply(event.data)

        nextTick(() => {
          scrollToFloor(event.data.floor)
        })
        break
      case 'deleted':
        removeReply(event.data.id)
        break
    }

    tempReplyStore.clearSuccessfulReply()
  },
  { deep: true }
)
</script>

<template>
  <div class="flex flex-col gap-4 lg:flex-row lg:items-start">
    <TopicDetailMasterUser v-if="topic.user" :user="topic.user" />

    <div class="min-w-0 flex-1 space-y-4">
      <TopicDetailHiddenNotice
        v-if="topic.status === 1"
        :hidden-by="topic.hidden_by"
      />

      <TopicDetailScopeNotice
        v-if="topic.access_scope && topic.access_scope !== 'public'"
        :scope="topic.access_scope"
      />

      <TopicDetailMaster :topic="topic" />

      <TopicMiniappContainer
        :topic-id="topic.id"
        :is-topic-admin="isTopicAdmin"
      />

      <div id="comments-anchor" class="scroll-mt-20">
        <TopicDetailTool
          :reply-count="topic.reply_count"
          :status="status"
          :sort-order="sortOrder"
          @set-sort-order="setSort"
        />
      </div>

      <section id="reply-section" class="space-y-4">
        <div v-if="hasEarlier && status !== 'pending'" class="text-center">
          <KunButton size="lg" variant="flat" @click="loadEarlier">
            加载更早的回复
          </KunButton>
        </div>

        <div
          v-if="status === 'pending' && replies.length === 0"
          class="flex justify-center py-16"
        >
          <KunLoading description="少女祈祷中..." />
        </div>

        <TopicReplyList
          v-else-if="replies.length > 0"
          :initial-replies="replies"
          :topic-id="topic.id"
          :title="topic.title"
        />

        <div class="py-6 text-center">
          <KunButton
            v-if="!isComplete && status !== 'pending'"
            size="lg"
            variant="flat"
            @click="loadMore"
          >
            加载更多
          </KunButton>
          <KunLoading v-if="status === 'pending' && replies.length > 0" />
          <p v-if="isComplete" class="text-default-500">
            {{ `(｡>︿<｡) 已经一滴回复都不剩了哦~` }}
          </p>
        </div>
      </section>

      <TopicDetailActionBar :topic="topic" />
    </div>
  </div>
</template>
