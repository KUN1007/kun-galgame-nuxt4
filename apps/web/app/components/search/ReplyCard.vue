<script setup lang="ts">
const props = defineProps<{
  reply: SearchResultReply
  keywords?: string
}>()

const snippet = computed(() => markdownToText(props.reply.content))
</script>

<template>
  <KunLink
    color="default"
    underline="none"
    :to="replyPermalink(`/topic/${reply.topic_id}`, reply.floor)"
    class-name="flex-col items-start w-full"
  >
    <div class="flex w-full items-center gap-2">
      <KunIcon class="text-primary size-5 shrink-0" name="carbon:reply" />
      <span class="truncate text-lg">
        <SearchHighlight :text="reply.topic_title" :keywords="keywords" />
      </span>
      <span class="text-default-500 ml-auto shrink-0 text-sm">
        <KunTime :time="reply.created" />
      </span>
    </div>

    <div v-if="snippet" class="mt-2 w-full">
      <div class="mb-2 flex items-center">
        <KunAvatar :user="reply.user" :is-navigation="false" />
        <span class="ml-2 text-sm">{{ reply.user.name }}</span>
        <span class="text-default-400 ml-2 text-xs">#{{ reply.floor }}</span>
      </div>
      <p class="text-default-700 line-clamp-3 text-sm">
        <SearchHighlight :text="snippet" :keywords="keywords" />
      </p>
    </div>
  </KunLink>
</template>
