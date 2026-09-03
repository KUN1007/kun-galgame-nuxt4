<script setup lang="ts">
const props = defineProps<{
  type: SearchPagedType
  results: SearchResult[]
  keywords: string
}>()

const isTopicResults = (results: unknown[]): results is SearchResultTopic[] =>
  props.type === 'topic'
const isGalgameResults = (
  results: unknown[]
): results is SearchResultGalgame[] => props.type === 'galgame'
const isToolsetResults = (
  results: unknown[]
): results is SearchResultToolset[] => props.type === 'toolset'
const isUserResults = (results: unknown[]): results is SearchResultUser[] =>
  props.type === 'user'
const isReplyResults = (results: unknown[]): results is SearchResultReply[] =>
  props.type === 'reply'
const isCommentResults = (
  results: unknown[]
): results is SearchResultComment[] => props.type === 'comment'
</script>

<template>
  <div>
    <div v-if="isTopicResults(results)" class="space-y-3">
      <KunCard v-for="topic in results" :key="topic.id">
        <HomeTopicCard :topic="topic" :keywords="keywords" />
      </KunCard>
    </div>

    <GalgameCard
      v-if="isGalgameResults(results)"
      :is-transparent="true"
      :galgames="results"
    />

    <ToolsetCard v-if="isToolsetResults(results)" :items="results" />

    <div v-if="isUserResults(results)" class="grid gap-3 sm:grid-cols-2">
      <SearchUserCard
        v-for="user in results"
        :key="user.id"
        :user="user"
        :keywords="keywords"
      />
    </div>

    <div v-if="isReplyResults(results)" class="space-y-3">
      <KunCard v-for="reply in results" :key="reply.id">
        <SearchReplyCard :reply="reply" :keywords="keywords" />
      </KunCard>
    </div>

    <div v-if="isCommentResults(results)" class="space-y-3">
      <KunCard v-for="comment in results" :key="comment.id">
        <SearchCommentCard :comment="comment" :keywords="keywords" />
      </KunCard>
    </div>
  </div>
</template>
