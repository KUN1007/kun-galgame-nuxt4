<script setup lang="ts">
defineProps<{
  topic: SearchResultTopic
  keywords?: string
}>()
</script>

<template>
  <KunLink
    color="default"
    underline="none"
    class-name="flex-col items-start w-full gap-1.5"
    :to="`/topic/${topic.id}`"
  >
    <div class="flex w-full items-baseline gap-2">
      <h3 class="hover:text-primary min-w-0 flex-1 truncate font-medium">
        <SearchHighlight :text="topic.title" :keywords="keywords" />
      </h3>
      <span class="text-default-400 shrink-0 text-xs">
        <KunTime :time="topic.status_update_time" />
      </span>
    </div>

    <div class="flex w-full flex-wrap items-center gap-x-3 gap-y-1 text-xs">
      <TopicBadgeGroup
        :section="topic.section"
        :upvote-time="topic.upvote_time"
        :has-best-answer="topic.has_best_answer"
        :mini-apps="topic.mini_apps"
        :is-n-s-f-w-topic="topic.is_nsfw_topic"
      />

      <span class="text-default-500 flex items-center gap-1">
        <KunAvatar size="xs" :user="topic.user" :is-navigation="false" />
        {{ topic.user.name }}
      </span>

      <span
        class="text-default-500 ml-auto flex items-center gap-3 tabular-nums"
      >
        <span class="flex items-center gap-1">
          <KunIcon name="lucide:eye" class="size-3.5" />
          {{ formatNumber(topic.view) }}
        </span>
        <span class="flex items-center gap-1">
          <KunIcon name="lucide:thumbs-up" class="size-3.5" />
          {{ topic.like_count }}
        </span>
        <span class="flex items-center gap-1">
          <KunIcon name="carbon:reply" class="size-3.5" />
          {{ topic.reply_count + topic.comment_count }}
        </span>
      </span>
    </div>
  </KunLink>
</template>
