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
const isResourceResults = (
  results: unknown[]
): results is SearchResultResource[] => props.type === 'resource'
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
    <GalgameCard
      v-if="isGalgameResults(results)"
      :is-transparent="true"
      :galgames="results"
    />

    <div v-if="isTopicResults(results)" class="space-y-2">
      <KunCard v-for="topic in results" :key="topic.id" padding="sm">
        <SearchTopicCard :topic="topic" :keywords="keywords" />
      </KunCard>
    </div>

    <div v-if="isResourceResults(results)" class="space-y-2">
      <KunCard v-for="resource in results" :key="resource.id" padding="sm">
        <SearchResourceCard :resource="resource" :keywords="keywords" />
      </KunCard>
    </div>

    <ToolsetCard
      v-if="isToolsetResults(results)"
      :items="results"
      :keywords="keywords"
    />

    <div v-if="isUserResults(results)" class="grid gap-2 sm:grid-cols-2">
      <SearchUserCard
        v-for="user in results"
        :key="user.id"
        :user="user"
        :keywords="keywords"
      />
    </div>

    <div v-if="isReplyResults(results)" class="space-y-2">
      <KunCard v-for="reply in results" :key="reply.id" padding="sm">
        <SearchReplyCard :reply="reply" :keywords="keywords" />
      </KunCard>
    </div>

    <div v-if="isCommentResults(results)" class="space-y-2">
      <KunCard v-for="comment in results" :key="comment.id" padding="sm">
        <SearchCommentCard :comment="comment" :keywords="keywords" />
      </KunCard>
    </div>
  </div>
</template>
