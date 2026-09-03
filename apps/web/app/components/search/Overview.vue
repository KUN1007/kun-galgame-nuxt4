<script setup lang="ts">
const props = defineProps<{
  keywords: string
  overview: SearchOverviewResult | null
  pending: boolean
  failed: boolean
}>()

const emit = defineEmits<{
  open: [value: SearchType]
}>()

const entityGroups = computed(
  () => props.overview?.entities.filter((group) => group.items.length) ?? []
)

const isEmpty = computed(() => {
  const totals = props.overview?.totals
  return (
    !props.pending &&
    !!totals &&
    Object.values(totals).every((value) => value === 0)
  )
})
</script>

<template>
  <div v-if="pending" class="space-y-8">
    <SearchSkeleton shape="card" :count="3" />
    <SearchSkeleton shape="row" :count="3" />
  </div>

  <KunNull v-else-if="failed" description="搜索没能完成, 请稍后重试" />

  <KunNull v-else-if="isEmpty" description="杂鱼杂鱼杂鱼~什么也没有搜索到" />

  <div v-else-if="overview" class="space-y-10">
    <SearchSection
      v-if="overview.galgames.length"
      type="galgame"
      :total="overview.totals.galgame"
      :shown="overview.galgames.length"
      @open="emit('open', $event)"
    >
      <GalgameCard :is-transparent="true" :galgames="overview.galgames" />
    </SearchSection>

    <SearchSection
      v-if="overview.topics.length"
      type="topic"
      :total="overview.totals.topic"
      :shown="overview.topics.length"
      @open="emit('open', $event)"
    >
      <div class="space-y-2">
        <KunCard v-for="topic in overview.topics" :key="topic.id" padding="sm">
          <SearchTopicCard :topic="topic" :keywords="keywords" />
        </KunCard>
      </div>
    </SearchSection>

    <SearchSection
      v-if="entityGroups.length"
      type="entity"
      :total="overview.totals.entity"
      :shown="entityGroups.reduce((sum, group) => sum + group.items.length, 0)"
      @open="emit('open', $event)"
    >
      <div class="space-y-5">
        <SearchEntityGroup
          v-for="group in entityGroups"
          :key="group.family"
          :group="group"
          :keywords="keywords"
          :show-header="true"
        />
      </div>
    </SearchSection>

    <SearchSection
      v-if="overview.resources.length"
      type="resource"
      :total="overview.totals.resource"
      :shown="overview.resources.length"
      @open="emit('open', $event)"
    >
      <div class="space-y-2">
        <KunCard
          v-for="resource in overview.resources"
          :key="resource.id"
          padding="sm"
        >
          <SearchResourceCard :resource="resource" :keywords="keywords" />
        </KunCard>
      </div>
    </SearchSection>

    <SearchSection
      v-if="overview.users.length"
      type="user"
      :total="overview.totals.user"
      :shown="overview.users.length"
      @open="emit('open', $event)"
    >
      <div class="grid gap-2 sm:grid-cols-2">
        <SearchUserCard
          v-for="user in overview.users"
          :key="user.id"
          :user="user"
          :keywords="keywords"
        />
      </div>
    </SearchSection>

    <SearchSection
      v-if="overview.replies.length"
      type="reply"
      :total="overview.totals.reply"
      :shown="overview.replies.length"
      @open="emit('open', $event)"
    >
      <div class="space-y-2">
        <KunCard v-for="reply in overview.replies" :key="reply.id" padding="sm">
          <SearchReplyCard :reply="reply" :keywords="keywords" />
        </KunCard>
      </div>
    </SearchSection>

    <SearchSection
      v-if="overview.comments.length"
      type="comment"
      :total="overview.totals.comment"
      :shown="overview.comments.length"
      @open="emit('open', $event)"
    >
      <div class="space-y-2">
        <KunCard
          v-for="comment in overview.comments"
          :key="comment.id"
          padding="sm"
        >
          <SearchCommentCard :comment="comment" :keywords="keywords" />
        </KunCard>
      </div>
    </SearchSection>

    <SearchSection
      v-if="overview.toolsets.length"
      type="toolset"
      :total="overview.totals.toolset"
      :shown="overview.toolsets.length"
      @open="emit('open', $event)"
    >
      <ToolsetCard :items="overview.toolsets" :keywords="keywords" />
    </SearchSection>
  </div>
</template>
