<script setup lang="ts">
import { watchDebounced } from '@vueuse/core'
import type { KunTabItem } from '@kungal/ui-vue'
import { topicHiddenByMeta } from '~/constants/topic'

definePageMeta({
  middleware: 'permission',
  permissions: ['topic.view_hidden']
})

useKunDisableSeo('隐藏话题管理')

interface HiddenTopic {
  id: number
  title: string
  hidden_by: string
  reply_count: number
  status_update_time: Date | string
  created: Date | string
  user: KunUser
}

interface HiddenTopicList {
  topics: HiddenTopic[]
  total: number
}

interface TopicPurgeStats {
  id: number
  title: string
  status: number
  hidden_by: string
  user: KunUser
  replies: number
  comments: number
  polls: number
  lotteries: number
  drawn_lotteries: number
  favorites: number
}

type PurgeCountKey =
  | 'replies'
  | 'comments'
  | 'polls'
  | 'lotteries'
  | 'drawn_lotteries'
  | 'favorites'

const PURGE_COUNT_LABELS: { key: PurgeCountKey; label: string }[] = [
  { key: 'replies', label: '回复' },
  { key: 'comments', label: '评论' },
  { key: 'polls', label: '投票' },
  { key: 'lotteries', label: '抽奖' },
  { key: 'drawn_lotteries', label: '已开奖抽奖' },
  { key: 'favorites', label: '收藏' }
]

const HIDDEN_BY_TABS: KunTabItem[] = [
  { value: 'all', textValue: '全部' },
  { value: 'author', textValue: '作者隐藏' },
  { value: 'moderator', textValue: '管理员隐藏' },
  { value: 'trust', textValue: '风纪隐藏' }
]

const canDeleteTopic = useCan('topic.delete_any')

const activeFilter = ref('all')
const searchQuery = ref('')

const pageData = reactive({
  page: 1,
  limit: 30,
  hidden_by: '',
  keywords: ''
})

watch(activeFilter, (value) => {
  pageData.hidden_by = value === 'all' ? '' : value
  pageData.page = 1
})

watchDebounced(
  searchQuery,
  (value) => {
    pageData.keywords = value.trim()
    pageData.page = 1
  },
  { debounce: 500, maxWait: 1000 }
)

const { data, status } = await useKunFetch<HiddenTopicList>(
  '/admin/topic/hidden',
  { query: pageData }
)

const isPurgeOpen = ref(false)
const isLoadingStats = ref(false)
const isPurging = ref(false)
const purgeTarget = ref<TopicPurgeStats | null>(null)

const purgeTotal = computed(() => {
  const target = purgeTarget.value
  if (!target) {
    return 0
  }
  // drawn_lotteries is a subset of lotteries; summing both counts them twice
  return PURGE_COUNT_LABELS.reduce(
    (sum, item) =>
      item.key === 'drawn_lotteries' ? sum : sum + target[item.key],
    0
  )
})

const openPurge = async (topic: HiddenTopic) => {
  isPurgeOpen.value = true
  isLoadingStats.value = true
  purgeTarget.value = null
  purgeTarget.value = await kunFetch<TopicPurgeStats>(
    `/admin/topic/${topic.id}/purge-stats`
  )
  isLoadingStats.value = false
  if (!purgeTarget.value) {
    isPurgeOpen.value = false
  }
}

const handlePurge = async () => {
  const target = purgeTarget.value
  if (!target) {
    return
  }
  isPurging.value = true
  const deleted = await kunFetch<TopicPurgeStats>(`/admin/topic/${target.id}`, {
    method: 'DELETE'
  })
  isPurging.value = false
  if (!deleted) {
    return
  }

  isPurgeOpen.value = false
  purgeTarget.value = null
  if (data.value) {
    data.value.topics = data.value.topics.filter((t) => t.id !== deleted.id)
    data.value.total = Math.max(0, data.value.total - 1)
  }
  useMessage(`已彻底删除话题《${deleted.title}》`, 'success')
}
</script>

<template>
  <div class="w-full space-y-4">
    <KunHeader
      name="隐藏话题管理"
      description="此处汇总全站被隐藏的话题, 包含作者自行隐藏、管理员隐藏与风纪处置隐藏三种来源。隐藏只是让话题从列表与搜索中消失, 作者与持有查看权限的管理人员仍可访问; 若要连同回复、评论、投票、抽奖与收藏一并抹除, 请使用彻底删除。"
    />

    <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
      <KunTab
        v-model="activeFilter"
        :items="HIDDEN_BY_TABS"
        variant="underlined"
        color="primary"
        size="sm"
        class-name="min-w-0 flex-1"
      />
      <!-- KunInput's root keeps w-full regardless of class-name (it lands on
           the inner input), so an unwrapped KunInput takes the whole flex row
           and crushes the KunTab beside it to zero width. -->
      <div class="w-full shrink-0 sm:w-72">
        <KunInput
          v-model="searchQuery"
          type="text"
          placeholder="输入话题标题以搜索"
        />
      </div>
    </div>

    <KunDivider />

    <KunLoading v-if="status === 'pending'" />

    <KunNull v-else-if="!data?.topics.length" description="暂无被隐藏的话题" />

    <div v-else class="flex flex-col gap-3">
      <div
        v-for="topic in data.topics"
        :key="topic.id"
        class="dark:border-default-200 flex flex-col gap-3 rounded-lg border border-transparent p-3 sm:flex-row sm:items-start"
      >
        <div class="min-w-0 flex-1 space-y-2">
          <div class="flex flex-wrap items-center gap-2">
            <KunLink :to="`/topic/${topic.id}`" target="_blank">
              {{ topic.title }}
            </KunLink>
            <KunChip
              size="xs"
              variant="flat"
              :color="topicHiddenByMeta(topic.hidden_by).color"
            >
              {{ topicHiddenByMeta(topic.hidden_by).label }}
            </KunChip>
          </div>

          <KunUserChip :user="topic.user" size="xs" />

          <div
            class="text-default-500 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm"
          >
            <span class="flex items-center gap-1">
              <KunIcon name="carbon:reply" class="size-4" />
              {{ topic.reply_count }}
            </span>
            <span class="flex items-center gap-1">
              <KunIcon name="lucide:eye-off" class="size-4" />
              隐藏于
              <KunTime
                :time="topic.status_update_time"
                type="datetime"
                show-year
              />
            </span>
            <span class="flex items-center gap-1">
              <KunIcon name="lucide:clock" class="size-4" />
              发布于
              <KunTime :time="topic.created" type="date" show-year />
            </span>
          </div>
        </div>

        <KunButton
          v-if="canDeleteTopic"
          size="sm"
          color="danger"
          variant="flat"
          class-name="shrink-0"
          @click="openPurge(topic)"
        >
          <KunIcon name="lucide:trash-2" />
          彻底删除
        </KunButton>
      </div>
    </div>

    <KunPagination
      v-if="data && data.total > pageData.limit"
      v-model:current-page="pageData.page"
      :total-page="Math.ceil(data.total / pageData.limit)"
      :is-loading="status === 'pending'"
    />

    <KunModal
      v-model="isPurgeOpen"
      role="alertdialog"
      title="彻底删除这个话题"
      description="删除后无法恢复, 话题及其下的全部内容都会从数据库中抹除, 已发放的萌萌点不会退回。"
      inner-class-name="w-full max-w-lg"
    >
      <KunLoading v-if="isLoadingStats" />

      <div v-else-if="purgeTarget" class="space-y-4">
        <KunInfo
          color="danger"
          variant="flat"
          icon="lucide:triangle-alert"
          title="这是不可撤销的操作"
          description="这不是隐藏, 也不是软删除。确认后话题连同下列内容一起从数据库消失, 没有任何恢复手段。"
        />

        <div class="space-y-1">
          <p class="font-medium break-words">{{ purgeTarget.title }}</p>
          <KunUserChip :user="purgeTarget.user" size="xs" />
        </div>

        <div class="flex flex-wrap gap-2 text-sm">
          <KunChip
            v-for="item in PURGE_COUNT_LABELS"
            :key="item.key"
            size="sm"
            variant="flat"
            :color="purgeTarget[item.key] ? 'danger' : 'default'"
          >
            {{ item.label }} {{ purgeTarget[item.key] }}
          </KunChip>
        </div>

        <p class="text-default-500 text-sm">
          共 {{ purgeTotal }} 项关联数据将一并删除。
        </p>

        <div class="flex justify-end gap-2">
          <KunButton
            variant="light"
            :disabled="isPurging"
            @click="isPurgeOpen = false"
          >
            取消
          </KunButton>
          <KunButton
            color="danger"
            :loading="isPurging"
            :disabled="isPurging"
            @click="handlePurge"
          >
            确认彻底删除
          </KunButton>
        </div>
      </div>
    </KunModal>
  </div>
</template>
