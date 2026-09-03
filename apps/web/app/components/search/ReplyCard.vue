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
    class-name="flex-col items-start w-full gap-1.5"
  >
    <div class="flex w-full items-baseline gap-2">
      <KunIcon
        class="text-primary size-3.5 shrink-0 self-center"
        name="carbon:reply"
      />
      <span class="min-w-0 flex-1 truncate text-sm font-medium">
        <SearchHighlight :text="reply.topic_title" :keywords="keywords" />
      </span>
      <span class="text-default-400 shrink-0 text-xs tabular-nums">
        #{{ reply.floor }}
      </span>
    </div>

    <p v-if="snippet" class="text-default-700 line-clamp-2 w-full text-sm">
      <SearchHighlight :text="snippet" :keywords="keywords" />
    </p>

    <div class="text-default-500 flex w-full items-center gap-1.5 text-xs">
      <KunAvatar size="xs" :user="reply.user" :is-navigation="false" />
      <span class="truncate">{{ reply.user.name }}</span>
      <span class="ml-auto shrink-0">
        <KunTime :time="reply.created" />
      </span>
    </div>
  </KunLink>
</template>
