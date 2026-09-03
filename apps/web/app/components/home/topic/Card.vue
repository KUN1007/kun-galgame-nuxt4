<script setup lang="ts">
defineProps<{
  topic: HomeTopic
  isTransparent?: boolean
  /** Search only: highlights the matched run inside the title. */
  keywords?: string
}>()
</script>

<template>
  <KunLink
    underline="none"
    class-name="flex-col items-start w-full"
    :to="`/topic/${topic.id}`"
  >
    <div class="flex w-full items-center justify-between gap-4">
      <h3
        class="hover:text-primary line-clamp-2 text-base font-medium transition-colors sm:text-lg"
      >
        <SearchHighlight :text="topic.title" :keywords="keywords" />
      </h3>

      <span class="text-default-500 shrink-0 text-sm">
        <KunTime :time="topic.status_update_time" />
      </span>
    </div>

    <TopicCoverGrid
      v-if="topic.cover_images?.length"
      :images="topic.cover_images"
      :meta="topic.cover_image_meta"
      class="my-1"
    />

    <div class="flex w-full flex-wrap items-center justify-between gap-2">
      <TopicBadgeGroup
        :section="topic.section"
        :upvote-time="topic.upvote_time"
        :has-best-answer="topic.has_best_answer"
        :mini-apps="topic.mini_apps"
        :is-n-s-f-w-topic="topic.is_nsfw_topic"
      />

      <div class="text-default-700 flex items-center gap-4 text-sm">
        <span class="flex items-center gap-1">
          <KunIcon name="lucide:eye" class="h-4 w-4" />
          {{ formatNumber(topic.view) }}
        </span>
        <span class="flex items-center gap-1">
          <KunIcon name="lucide:thumbs-up" class="h-4 w-4" />
          {{ topic.like_count }}
        </span>
        <span class="flex items-center gap-1">
          <KunIcon name="carbon:reply" class="h-4 w-4" />
          {{ topic.reply_count + topic.comment_count }}
        </span>
      </div>
    </div>
  </KunLink>
</template>
