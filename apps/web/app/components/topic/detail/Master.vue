<script setup lang="ts">
const props = defineProps<{
  topic: TopicDetail
}>()

const unseenCovers = computed(() => {
  const body = props.topic.content_markdown ?? ''
  return (props.topic.cover_images ?? []).filter((token) => {
    const hash = token.split('/').pop()
    return hash ? !body.includes(hash) : false
  })
})

provide(
  reactionsKey,
  useReactions({
    topicId: props.topic.id,
    targetUserId: props.topic.user.id,
    reactions: props.topic.reactions,
    showReactors: true
  })
)
</script>

<template>
  <div id="0" class="outline-primary rounded-lg outline-offset-2">
    <KunCard
      :is-transparent="false"
      :is-hoverable="false"
      class-name="w-full min-w-0"
      content-class="gap-4 justify-start"
    >
      <header class="space-y-3">
        <h1
          class="text-3xl leading-tight font-bold tracking-tight break-words lg:text-4xl"
        >
          {{ topic.title }}
        </h1>

        <TopicBadgeGroup
          :section="topic.section"
          :upvote-time="topic.upvote_time"
          :has-best-answer="false"
          :mini-apps="topic.mini_apps"
          :is-n-s-f-w-topic="topic.is_nsfw"
          :is-nav-to-section="true"
        />

        <div
          class="text-default-500 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm"
        >
          <span class="flex items-center gap-1.5">
            <KunIcon name="lucide:eye" class="size-4" />
            {{ topic.view }}
          </span>
          <span class="flex items-center gap-1.5">
            <KunIcon name="lucide:clock" class="size-4" />
            <KunTime :time="topic.created" type="datetime" show-year />
          </span>
          <span v-if="topic.edited" class="flex items-center gap-1.5">
            <KunIcon name="lucide:pencil-line" class="size-4" />
            编辑于 <KunTime :time="topic.edited" type="datetime" show-year />
          </span>
        </div>
      </header>

      <TopicDetailBestAnswer
        v-if="topic.best_answer"
        :best-answer="topic.best_answer"
      />

      <TopicDetailUser
        class-name="lg:hidden"
        :user="topic.user"
        :created="topic.created"
        :edited="topic.edited"
        :topic-id="topic.id"
        :floor="0"
        :show-addition="false"
      />

      <KunDivider />

      <TopicCoverGrid
        v-if="unseenCovers.length"
        :images="unseenCovers"
        :meta="topic.cover_image_meta"
        zoomable
      />

      <KunContent
        class="kun-master"
        :content="renderKatex(topic.content_html)"
      />

      <KunDivider />

      <TopicUpvoteRecords :topic-id="topic.id" />

      <div class="flex flex-wrap items-center gap-1.5">
        <TopicReactionBar />
        <span class="md:hidden">
          <TopicReactionTrigger />
        </span>
      </div>

      <p class="text-default-500 ml-auto text-sm">
        本文版权遵循
        <KunLink
          underline="hover"
          size="sm"
          class-name="text-default-500"
          target="_blank"
          rel="noopener noreferrer"
          to="https://creativecommons.org/licenses/by-nc/4.0/deed.en"
        >
          CC BY-NC 协议
        </KunLink>
        和
        <KunLink
          underline="hover"
          size="sm"
          class-name="text-default-500"
          to="/doc/article-copyright"
        >
          本站版权政策
        </KunLink>
      </p>

      <TopicFooter :topic="topic" />
    </KunCard>
  </div>
</template>
