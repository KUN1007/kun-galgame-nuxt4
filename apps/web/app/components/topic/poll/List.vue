<script setup lang="ts">
import { ref, computed } from 'vue'
import { usePoll } from '~/composables/topic/usePoll'
import { TOPIC_POLL_VISIBILITY_MAP } from '~/constants/topic'

const props = defineProps<{
  poll: TopicPoll
  isTopicAdmin: boolean
}>()

const emits = defineEmits<{
  edit: [poll: TopicPoll]
  refresh: []
}>()

const { submitVote, deletePoll } = usePoll(props.poll.topic_id)
const selectedOptions = ref<number[]>([])
const isLoading = ref(false)
const isLogModalOpen = ref(false)

const nowMs = useState(`kun-poll-now-${useId()}`, () => Date.now())
onMounted(() => {
  nowMs.value = Date.now()
})

const isPollEnded = computed(() => {
  if (props.poll.status === 'closed') {
    return true
  }
  if (!props.poll.deadline) {
    return false
  }
  return new Date(props.poll.deadline).getTime() < nowMs.value
})

const canViewResults = computed(() => {
  if (isPollEnded.value) {
    return true
  }
  const visibility = props.poll.result_visibility
  if (visibility === 'always') {
    return true
  }
  if (visibility === 'after_vote' && props.poll.has_voted) {
    return true
  }
  return false
})

const percentOf = (voteCount: number | null) =>
  ((voteCount || 0) / (props.poll.vote_count || 1)) * 100

const metaLine = computed(() => [
  props.poll.min_choice === props.poll.max_choice
    ? `必选 ${props.poll.max_choice} 项`
    : `可选 ${props.poll.min_choice}-${props.poll.max_choice} 项`,
  TOPIC_POLL_VISIBILITY_MAP[props.poll.result_visibility],
  props.poll.can_change_vote ? '可修改投票' : '投出后不可修改',
  props.poll.is_anonymous ? '匿名投票' : '实名投票',
  props.poll.vote_count ? `共 ${props.poll.vote_count} 票` : ''
])

const handleOptionClick = (optionId: number) => {
  if (props.poll.type === 'single') {
    selectedOptions.value = [optionId]
  } else {
    const index = selectedOptions.value.indexOf(optionId)
    if (index > -1) {
      selectedOptions.value.splice(index, 1)
    } else {
      if (
        props.poll.max_choice &&
        selectedOptions.value.length >= props.poll.max_choice
      ) {
        selectedOptions.value.shift()
      }
      selectedOptions.value.push(optionId)
    }
  }
}

const handleVote = async () => {
  if (!requireLogin()) return
  if (selectedOptions.value.length === 0) return
  isLoading.value = true
  await submitVote(props.poll.id, selectedOptions.value)
  isLoading.value = false
  emits('refresh')
}

const handleDelete = async () => {
  await deletePoll(props.poll.id)
  emits('refresh')
}
</script>

<template>
  <KunCard
    :is-hoverable="false"
    :is-transparent="false"
    content-class="space-y-3"
  >
    <TopicMiniappHeader
      app-key="poll"
      :title="poll.title"
      :meta="metaLine"
      :status="isPollEnded ? '已结束' : '进行中'"
      :status-color="isPollEnded ? 'default' : 'success'"
    />

    <p v-if="poll.description" class="text-default-500 text-sm">
      {{ poll.description }}
    </p>

    <div class="flex flex-col gap-2">
      <div
        v-for="option in poll.option"
        :key="option.id"
        :class="
          cn(
            'border-default-200 relative overflow-hidden rounded-lg border transition-colors',
            !isPollEnded && 'hover:border-primary/60 cursor-pointer',
            selectedOptions.includes(option.id) && 'border-primary bg-primary/5'
          )
        "
        @click="handleOptionClick(option.id)"
      >
        <div
          v-if="canViewResults"
          class="bg-primary/15 absolute inset-y-0 left-0 transition-all duration-500"
          :style="{ width: `${percentOf(option.vote_count)}%` }"
        />

        <div
          class="relative flex items-center justify-between gap-3 px-3 py-2.5"
        >
          <div class="flex min-w-0 items-center gap-3">
            <KunCheckBox
              v-if="poll.type === 'multiple'"
              color="primary"
              :model-value="selectedOptions.includes(option.id)"
              :disabled="isPollEnded"
              @click.stop
              @change="handleOptionClick(option.id)"
            />

            <KunCheckBox
              v-else
              color="primary"
              type="single"
              :model-value="selectedOptions.includes(option.id)"
              :disabled="isPollEnded"
              @click.stop
              @change="handleOptionClick(option.id)"
            />

            <span class="truncate text-sm">{{ option.text }}</span>
            <KunIcon
              v-if="option.is_voted"
              name="lucide:check-circle-2"
              class="text-primary shrink-0"
            />
          </div>

          <div
            v-if="canViewResults"
            class="flex shrink-0 items-baseline gap-2 tabular-nums"
          >
            <span class="text-default-500 text-xs">
              {{ option.vote_count || 0 }} 票
            </span>
            <span class="w-14 text-right text-sm font-semibold">
              {{ percentOf(option.vote_count).toFixed(1) }}%
            </span>
          </div>
        </div>
      </div>
    </div>

    <div
      class="border-default-200 flex flex-wrap items-center justify-between gap-3 border-t pt-3"
    >
      <span v-if="!canViewResults" class="text-default-500 text-sm">
        {{
          poll.result_visibility === 'after_vote'
            ? '投票后可以看到结果'
            : '结束后才会公开结果'
        }}
      </span>
      <KunAvatarGroup
        v-else-if="!poll.is_anonymous && poll.vote_count"
        :users="poll.voters"
        :total="poll.vote_count"
      />

      <div class="ml-auto flex items-center gap-2">
        <KunButton
          v-if="!isPollEnded && (!poll.has_voted || poll.can_change_vote)"
          color="primary"
          size="sm"
          :loading="isLoading"
          :disabled="selectedOptions.length === 0"
          @click="handleVote"
        >
          {{ poll.has_voted ? '修改投票' : '投票' }}
        </KunButton>

        <template v-if="canViewResults && isTopicAdmin">
          <KunTooltip text="查看投票日志">
            <KunButton
              variant="light"
              color="default"
              size="sm"
              :is-icon-only="true"
              @click="isLogModalOpen = true"
            >
              <KunIcon name="lucide:history" />
            </KunButton>
          </KunTooltip>
          <KunTooltip text="编辑投票">
            <KunButton
              variant="light"
              color="default"
              size="sm"
              :is-icon-only="true"
              @click="emits('edit', poll)"
            >
              <KunIcon name="lucide:pencil" />
            </KunButton>
          </KunTooltip>
          <KunTooltip text="删除投票">
            <KunButton
              variant="light"
              color="danger"
              size="sm"
              :is-icon-only="true"
              @click="handleDelete"
            >
              <KunIcon name="lucide:trash-2" />
            </KunButton>
          </KunTooltip>
        </template>
      </div>
    </div>

    <TopicPollLog
      v-model="isLogModalOpen"
      :poll-id="poll.id"
      :topic-id="poll.topic_id"
    />
  </KunCard>
</template>
