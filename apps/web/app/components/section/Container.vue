<script setup lang="ts">
import { KUN_TOPIC_CATEGORY, KUN_TOPIC_SECTION } from '~/constants/topic'
import { KUN_TOPIC_SECTION_DESCRIPTION_MAP } from '~/constants/section'

const props = defineProps<{
  section: string
}>()
const page = ref(1)

const categoryMap: Record<string, string> = {
  g: 'galgame',
  t: 'technique',
  o: 'others'
}
const category = computed(
  () => KUN_TOPIC_CATEGORY[categoryMap[props.section[0]!]!]!
)

const { data, status } = await useKunFetch<SectionTopicList>('/section', {
  query: {
    section: props.section,
    sort_order: 'desc',
    page,
    limit: 30
  }
})

watch(
  () => status.value,
  () => {
    if (status.value === 'success') {
      window?.scrollTo({
        top: 0,
        behavior: 'smooth'
      })
    }
  }
)
</script>

<template>
  <div class="space-y-6">
    <KunHeader :description="KUN_TOPIC_SECTION_DESCRIPTION_MAP[section]">
      <template #title>
        <div class="flex items-center gap-2">
          <KunLink
            underline="hover"
            :to="`/category/${categoryMap[props.section[0]!]}`"
            class-name="text-2xl font-medium"
          >
            {{ category }}
          </KunLink>
          /
          <span class="text-lg">{{ KUN_TOPIC_SECTION[section] }}</span>
        </div>
      </template>
    </KunHeader>

    <KunCard
      :is-hoverable="true"
      :is-transparent="false"
      :is-pressable="true"
      content-class="items-start flex flex-row gap-3 flex-nowrap"
      v-for="(topic, index) in data?.topics"
      :key="index"
      :href="`/topic/${topic.id}`"
    >
      <KunAvatar
        :disable-floating="true"
        :user="topic.user"
        :is-navigation="false"
      />

      <div class="w-full space-y-2">
        <div class="flex items-center">
          <div class="mr-2 font-bold">{{ topic.user.name }}</div>
          <div class="text-default-500 text-sm">
            <KunTime :time="topic.created" type="datetime" show-year />
          </div>
        </div>

        <h2 class="hover:text-primary text-lg transition-colors">
          {{ topic.title }}
        </h2>

        <TopicBadgeGroup
          :section="[]"
          :has-best-answer="topic.has_best_answer"
          :mini-apps="topic.mini_apps"
          :is-n-s-f-w-topic="topic.is_nsfw_topic"
        />

        <div class="text-default-500 line-clamp-2 text-sm break-all">
          {{ markdownToText(topic.content) }}
        </div>

        <div class="text-default-700 flex gap-4 text-sm">
          <div class="flex items-center gap-2 text-inherit">
            <KunIcon name="lucide:eye" />
            {{ topic.view }}
          </div>
          <div class="flex items-center gap-2 text-inherit">
            <KunIcon name="lucide:thumbs-up" />
            {{ topic.like_count }}
          </div>
          <div class="flex items-center gap-2 text-inherit">
            <KunIcon name="carbon:reply" />
            {{ topic.reply_count }}
          </div>
        </div>
      </div>
    </KunCard>

    <KunPagination
      v-if="data"
      v-model:current-page="page"
      :total-page="Math.ceil(data.total / 30)"
      :is-loading="status === 'pending'"
    />
  </div>
</template>
