<script setup lang="ts">
import {
  kunUserTopicNavItem,
  type KUN_USER_PAGE_TOPIC_TYPE
} from '~/constants/user'

const props = defineProps<{
  userId: number
  type: (typeof KUN_USER_PAGE_TOPIC_TYPE)[number]
}>()

const { id: currentUserId } = usePersistUserStore()
const canViewHiddenTopic = useCan('topic.view_hidden')
const canSeeHidden = computed(
  () => currentUserId === props.userId || canViewHiddenTopic.value
)

const activeTab = ref(props.type)
const pageData = reactive({
  page: 1,
  limit: 50,
  type: props.type,
  userId: props.userId
})

const { data, status } = await useKunFetch<{
  topics: UserTopic[]
  total: number
}>(`/user/${props.userId}/topics`, { query: pageData })

const handleUpdateTopicHideStatus = async (topicId: number) => {
  const res = await useComponentMessageStore().alert(
    '八嘎杂鱼笨蛋萝莉, 你要取消隐藏该话题吗, 取消隐藏后该话题将对所有人可见'
  )
  if (!res) {
    return
  }

  const result = await kunFetch<string>(`/topic/${topicId}/hide`, {
    method: 'PUT',
    body: { topicId }
  })

  if (result) {
    useMessage('取消隐藏话题成功', 'success')
    await navigateTo(`/topic/${topicId}`)
  }
}
</script>

<template>
  <div class="space-y-3">
    <KunTab
      :items="kunUserTopicNavItem(userId, canSeeHidden)"
      :model-value="activeTab"
      variant="light"
      size="sm"
      scrollable
    />

    <div class="flex flex-col space-y-3" v-if="data && data.topics.length">
      <template v-if="pageData.type !== 'topic_hide'">
        <KunCard
          v-for="(topic, index) in data.topics"
          :key="index"
          :href="`/topic/${topic.id}`"
        >
          <div>
            {{ topic.title }}
          </div>
          <div class="text-default-500 text-sm">
            <KunTime :time="topic.created" type="date" show-year />
          </div>
        </KunCard>
      </template>

      <template v-else>
        <KunCard
          :is-hoverable="false"
          :is-transparent="true"
          v-for="(topic, index) in data.topics"
          :key="index"
        >
          <KunLink :to="`/topic/${topic.id}`">
            {{ topic.title }}
          </KunLink>
          <div
            class="text-default-500 flex items-center justify-between text-sm"
          >
            <KunTime :time="topic.created" type="date" show-year />
            <KunButton
              @click="handleUpdateTopicHideStatus(topic.id)"
              size="sm"
              variant="flat"
              color="primary"
            >
              取消隐藏
            </KunButton>
          </div>
        </KunCard>
      </template>
    </div>

    <KunPagination
      v-if="data && data.total > pageData.limit"
      v-model:current-page="pageData.page"
      :total-page="Math.ceil(data.total / pageData.limit)"
      :is-loading="status === 'pending'"
    />

    <KunNull
      v-if="data && !data.topics.length"
      description="这只笨蛋萝莉没有发布过任何话题"
    />
  </div>
</template>
