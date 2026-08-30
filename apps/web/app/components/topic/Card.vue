<script setup lang="ts">
const props = defineProps<{
  topic: TopicCard
}>()

const actionsCount = computed(
  () => props.topic.reply_count + props.topic.comment_count
)
</script>

<template>
  <NuxtLink
    :to="`/topic/${topic.id}`"
    class="group block space-y-2 py-4 first:pt-0 last:pb-0"
  >
    <h3
      class="group-hover:text-primary line-clamp-2 text-lg font-medium transition-colors"
    >
      {{ topic.title }}
    </h3>

    <TopicBadgeGroup
      :section="props.topic.section"
      :has-best-answer="topic.has_best_answer"
      :mini-apps="topic.mini_apps"
      :is-n-s-f-w-topic="topic.is_nsfw_topic"
    />

    <div class="text-default-600 flex flex-wrap items-center gap-2 text-sm">
      <KunAvatar
        :disable-floating="true"
        :user="topic.user"
        size="xs"
        :is-navigation="false"
      />
      <span>{{ topic.user.name }}</span>
      <KunTime :time="topic.created" type="relative" />

      <div class="text-default-500 ml-2 flex items-center gap-3">
        <span class="flex items-center gap-1">
          <KunIcon class="size-4" name="lucide:eye" />
          {{ formatNumber(props.topic.view) }}
        </span>
        <span v-if="props.topic.like_count" class="flex items-center gap-1">
          <KunIcon class="size-4" name="lucide:thumbs-up" />
          {{ props.topic.like_count }}
        </span>
        <span v-if="actionsCount" class="flex items-center gap-1">
          <KunIcon class="size-4" name="carbon:reply" />
          {{ actionsCount }}
        </span>
      </div>
    </div>
  </NuxtLink>
</template>
