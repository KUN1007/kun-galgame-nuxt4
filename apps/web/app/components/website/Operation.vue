<script setup lang="ts">
import type { CreateWebsitePayload, UpdateWebsitePayload } from './modal/types'

const props = defineProps<{
  website: WebsiteDetail
}>()

const emits = defineEmits<{
  refresh: []
}>()

type WebsiteData = (CreateWebsitePayload & { website_id?: number }) | undefined

const { id } = usePersistUserStore()
const canEditWebsite = useCan('website.edit')
const canDeleteWebsite = useCan('website.delete')

const utmLink = useUtmLink()

const isLiked = ref(props.website.is_liked)
const likeCount = ref(props.website.like_count)
type ActionType = 'update' | 'delete'
const loadingStates = reactive<Record<ActionType, boolean>>({
  update: false,
  delete: false
})

const likePending = ref(false)
const revertLike = (next: boolean) => {
  isLiked.value = !next
  likeCount.value += next ? -1 : 1
}
const onLike = async (next: boolean) => {
  if (!id) {
    useAuthModal().open()
    revertLike(next)
    return
  }
  likePending.value = true
  const result = await kunFetch(`/website/${props.website.id}/like`, {
    method: 'PUT',
    body: { website_id: props.website.id }
  })
  likePending.value = false
  if (!result) {
    revertLike(next)
    return
  }
  useMessage(next ? '点赞网站成功!' : '取消点赞成功!', 'success')
}

const showWebsiteModal = ref(false)
const editingWebsite = ref<WebsiteData>(undefined)

const handleOpenUpdateModal = () => {
  const { age_limit, category, tags, create_time, language, status, ...rest } =
    props.website
  editingWebsite.value = {
    ...rest,
    language: language as CreateWebsitePayload['language'],
    tag_ids: tags.map((t) => t.id),
    category_id: category.id,
    age_limit,
    status,
    create_time,
    website_id: props.website.id
  }
  showWebsiteModal.value = true
}

const handleUpdate = (data: CreateWebsitePayload | UpdateWebsitePayload) =>
  handleAction('update', async () => {
    const result = await kunFetch(`/website/${props.website.id}`, {
      method: 'PUT',
      body: data
    })

    if (result) {
      useMessage('重新编辑成功', 'success')
      showWebsiteModal.value = false
      emits('refresh')
    }
  })

const handleDelete = () =>
  handleAction('delete', async () => {
    const res = await useComponentMessageStore().alert(
      '您确定删除这个网站吗？',
      '这将会删除这个网站的所有信息, 该操作不可撤销'
    )
    if (!res) {
      return
    }

    const result = await kunFetch(`/website/${props.website.id}`, {
      method: 'DELETE',
      query: { website_id: props.website.id }
    })

    if (result) {
      useMessage('删除网站成功', 'success')
      await navigateTo('/website')
    }
  })

const handleAction = async (
  actionType: ActionType,
  action: () => Promise<void>
) => {
  if (loadingStates[actionType]) {
    return
  }

  if (!id) {
    useAuthModal().open()
    return
  }

  loadingStates[actionType] = true
  await action()
  loadingStates[actionType] = false
}
</script>

<template>
  <div class="flex flex-wrap gap-3">
    <KunTooltip text="点赞">
      <span class="flex">
        <KunReaction
          v-model="isLiked"
          v-model:count="likeCount"
          :disabled="likePending"
          size="lg"
          icon="lucide:thumbs-up"
          color="secondary"
          label="点赞"
          @change="onLike"
        />
      </span>
    </KunTooltip>

    <FavoriteToggle
      :favorited="website.is_favorited"
      :count="website.favorite_count"
      :endpoint="`/website/${website.id}/favorite`"
      :body="{ website_id: website.id }"
      :messages="['收藏网站成功!', '取消收藏成功!']"
      size="lg"
    />

    <KunTooltip v-if="canEditWebsite" text="编辑">
      <KunButton
        :is-icon-only="true"
        size="lg"
        variant="light"
        class-name="gap-1"
        @click="handleOpenUpdateModal"
      >
        <KunIcon name="lucide:pen" />
      </KunButton>
    </KunTooltip>

    <WebsiteModalWebsite
      v-model="showWebsiteModal"
      :initial-data="editingWebsite"
      :loading="loadingStates.update"
      @submit="handleUpdate"
    />

    <KunTooltip v-if="canDeleteWebsite" text="删除">
      <KunButton
        :is-icon-only="true"
        size="lg"
        color="danger"
        variant="light"
        class-name="gap-1"
        @click="handleDelete"
      >
        <KunIcon name="lucide:trash-2" />
      </KunButton>
    </KunTooltip>

    <KunButton
      target="_blank"
      :href="utmLink(`https://${website.url}`)"
      class-name="ml-auto"
    >
      访问网站
    </KunButton>
  </div>
</template>
