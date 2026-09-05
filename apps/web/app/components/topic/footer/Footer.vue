<script setup lang="ts">
defineProps<{
  topic: TopicDetail
}>()

const { id } = usePersistUserStore()
</script>

<template>
  <div class="mt-auto hidden items-center justify-between leading-none md:flex">
    <div class="flex items-center gap-1">
      <TopicFooterUpvote
        :topic-id="topic.id"
        :target-user-id="topic.user.id"
        :upvote-count="topic.upvote_count"
        :is-upvoted="topic.is_upvoted"
      />

      <TopicFooterFavorite
        :topic-id="topic.id"
        :target-user-id="topic.user.id"
        :favorite-count="topic.favorite_count"
        :is-favorite="topic.is_favorited"
      />

      <TopicReactionTrigger />
    </div>

    <div class="flex items-center gap-1">
      <TopicFooterReply
        :target-user-name="topic.user.name"
        :target-user-id="topic.user.id"
        :target-floor="0"
      />

      <TopicFooterRewrite :topic="topic" />

      <KunPopover position="top-end">
        <template #trigger>
          <KunReaction :toggle="false" icon="lucide:ellipsis" label="更多" />
        </template>

        <div class="flex w-54 flex-col gap-2 p-2">
          <KunButton
            variant="light"
            color="default"
            size="sm"
            class-name="w-full justify-start gap-2 whitespace-nowrap"
            @click="
              useKunCopy(
                `${topic.title}: https://www.kungal.com/topic/${topic.id}`
              )
            "
          >
            <KunIcon class-name="text-lg" name="lucide:share-2" />
            分享
          </KunButton>
          <TopicFooterHide
            v-if="id"
            :topic-id="topic.id"
            :status="topic.status"
            :hidden-by="topic.hidden_by"
          />
          <ReportButton
            v-if="topic.user.id !== id"
            menu
            subject-kind="forum_topic"
            :subject-id="topic.id"
            :snapshot="topic.title"
            :subject-url="`${kungal.domain.main}/topic/${topic.id}`"
          />
        </div>
      </KunPopover>
    </div>
  </div>
</template>
