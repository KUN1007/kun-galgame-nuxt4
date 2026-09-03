<script setup lang="ts">
defineProps<{
  comment: SearchResultComment
  keywords?: string
}>()
</script>

<template>
  <KunLink
    color="default"
    underline="none"
    :to="commentPermalink(`/topic/${comment.topic_id}`, comment.id)"
    class-name="flex-col items-start w-full"
  >
    <div class="flex w-full items-center gap-2">
      <KunIcon class="text-primary size-5 shrink-0" name="uil:comment-dots" />
      <span class="truncate text-lg">
        <SearchHighlight :text="comment.topic_title" :keywords="keywords" />
      </span>
      <span class="text-default-500 ml-auto shrink-0 text-sm">
        <KunTime :time="comment.created" />
      </span>
    </div>

    <div
      class="border-primary bg-primary/10 my-2 w-full rounded border-l-3 p-2 text-sm"
    >
      <SearchHighlight :text="comment.content" :keywords="keywords" />
    </div>

    <div class="flex items-center">
      <KunAvatar :user="comment.user" :is-navigation="false" />
      <span class="ml-2 text-sm">{{ comment.user.name }}</span>
    </div>
  </KunLink>
</template>
