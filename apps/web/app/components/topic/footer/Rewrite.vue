<script setup lang="ts">
import { KunTooltip } from '#components'
import { toTopicAccessRoles, toTopicAccessScope } from '~/constants/topic'

const props = defineProps<{
  topic: TopicDetail
  menu?: boolean
}>()

const {
  id,
  title,
  content,
  category,
  section,
  isNSFW,
  coverImages,
  accessScope,
  accessRoles,
  accessUserIds,
  isTopicRewriting
} = storeToRefs(useTempEditStore())
const { id: userId } = usePersistUserStore()
const canEditAnyTopic = useCan('topic.edit_any')
const isShowRewrite = computed(
  () => userId === props.topic.user.id || canEditAnyTopic.value
)

const rewriteTopic = async () => {
  id.value = props.topic.id
  title.value = props.topic.title
  content.value = props.topic.content_markdown
  category.value = props.topic.category
  section.value = props.topic.section ?? []
  isNSFW.value = !!props.topic.is_nsfw
  coverImages.value = props.topic.cover_images ?? []
  accessScope.value = toTopicAccessScope(props.topic.access_scope)
  accessRoles.value = toTopicAccessRoles(props.topic.access_grants?.roles)
  accessUserIds.value = props.topic.access_grants?.user_ids ?? []
  isTopicRewriting.value = true

  await navigateTo('/edit/topic')
}
</script>

<template>
  <template v-if="isShowRewrite">
    <KunButton
      v-if="menu"
      variant="light"
      color="default"
      size="sm"
      class-name="w-full justify-start gap-2 whitespace-nowrap"
      @click="rewriteTopic"
    >
      <KunIcon class-name="text-lg" name="lucide:pencil" />
      重新编辑
    </KunButton>

    <KunTooltip v-else text="重新编辑">
      <KunReaction
        :toggle="false"
        icon="lucide:pencil"
        label="重新编辑"
        @click="rewriteTopic"
      />
    </KunTooltip>
  </template>
</template>
