<script setup lang="ts">
import {
  GALGAME_RESOURCE_PROVIDER_MAP,
  GALGAME_RESOURCE_TYPE_ICON_MAP,
  GALGAME_RESOURCE_PLATFORM_ICON_MAP
} from '~/constants/galgameResource'
import {
  KUN_GALGAME_RESOURCE_TYPE_MAP,
  KUN_GALGAME_RESOURCE_LANGUAGE_MAP,
  KUN_GALGAME_RESOURCE_PLATFORM_MAP
} from '~/constants/galgame'
import { fetchTitle } from './_fetchTitle'

const props = defineProps<{
  resource: GalgameResource
  refresh: () => void
}>()

const providerName = ref('')
const details = ref<GalgameResourceDetails>()
const isModalOpen = ref(false)
const { id } = usePersistUserStore()
const { rewriteResourceId } = storeToRefs(useTempGalgameResourceStore())
const isFetching = ref(false)

const handleMarkValid = async (
  galgameId: number,
  galgameResourceId: number
) => {
  const res = await useComponentMessageStore().alert(
    '您确定重新标记资源链接有效吗？',
    '若您修复了资源链接，您可以重新标记资源链接有效。'
  )
  if (!res) {
    return
  }

  const result = await $fetch(`/api/galgame/${galgameId}/resource/valid`, {
    method: 'PUT',
    body: { galgameResourceId },
    watch: false,
    ...kungalgameResponseHandler
  })

  if (result) {
    useMessage(10548, 'success')
    props.refresh()
  }
}

watch(
  () => [rewriteResourceId.value, props.resource],
  () => {
    details.value = undefined
    isModalOpen.value = false
  }
)

const findKnownProvider = (domain: string): string | null => {
  if (GALGAME_RESOURCE_PROVIDER_MAP[domain]) {
    return GALGAME_RESOURCE_PROVIDER_MAP[domain]
  }
  // 'lanzou' - lanzouj, lanzouq, lanzouw
  if (domain.includes('lanzou')) {
    return GALGAME_RESOURCE_PROVIDER_MAP.lanzou!
  }
  return null
}

onMounted(async () => {
  const knownName = findKnownProvider(props.resource.linkDomain)

  if (knownName) {
    providerName.value = knownName
  } else {
    providerName.value = '正在获取资源提供方信息...'
    try {
      const result = await fetchTitle(props.resource.linkDomain)
      providerName.value = result.title || props.resource.linkDomain
    } catch (error) {
      providerName.value = props.resource.linkDomain
    }
  }
})
</script>

<template>
  <div class="space-y-2">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <div class="flex flex-wrap items-center gap-1 rounded-lg">
        <KunBadge color="primary">
          <KunIcon :name="GALGAME_RESOURCE_TYPE_ICON_MAP[resource.type]" />
          {{ KUN_GALGAME_RESOURCE_TYPE_MAP[resource.type] }}
        </KunBadge>
        <KunBadge color="warning">
          <KunIcon name="lucide:database" />
          {{ resource.size }}
        </KunBadge>
        <KunBadge color="success">
          <KunIcon
            :name="GALGAME_RESOURCE_PLATFORM_ICON_MAP[resource.platform]"
          />
          {{ KUN_GALGAME_RESOURCE_PLATFORM_MAP[resource.platform] }}
        </KunBadge>
        <KunBadge color="secondary">
          {{ KUN_GALGAME_RESOURCE_LANGUAGE_MAP[resource.language] }}
        </KunBadge>
      </div>

      <div class="ml-auto flex items-center gap-1">
        <span class="text-default-500 mr-2 text-sm">{{ providerName }}</span>

        <KunButton
          size="sm"
          v-if="id === resource.user.id && resource.status === 1"
          @click="handleMarkValid(resource.galgameId, resource.id)"
          :loading="isFetching"
        >
          重新标记有效
        </KunButton>

        <KunTooltip text="资源下载数">
          <KunButton
            :is-icon-only="true"
            variant="light"
            color="default"
            size="sm"
            class-name="gap-1"
          >
            <KunIcon name="lucide:download" />
            <span>{{ resource.download }}</span>
          </KunButton>
        </KunTooltip>

        <GalgameResourceLike
          v-if="id !== resource.user.id"
          :galgame-id="resource.galgameId"
          :galgame-resource-id="resource.id"
          :target-user-id="resource.user.id"
          :is-liked="resource.isLiked"
          :like-count="resource.likeCount"
        />

        <KunButton
          size="sm"
          variant="flat"
          v-if="resource.id !== rewriteResourceId"
          :href="`/galgame-resource/${resource.id}`"
          :loading="isFetching"
        >
          下载资源
        </KunButton>

        <KunTooltip text="举报违规">
          <KunButton
            :is-icon-only="true"
            color="danger"
            variant="light"
            v-if="id !== resource.user.id"
            href="/report"
          >
            <KunIcon name="lucide:triangle-alert" />
          </KunButton>
        </KunTooltip>

        <KunTooltip
          position="left"
          :text="resource.status ? '链接过期' : '链接有效'"
        >
          <div
            :class="
              cn(
                'h-3 w-3 shrink-0 rounded-full',
                resource.status ? 'bg-danger' : 'bg-success'
              )
            "
          />
        </KunTooltip>
      </div>
    </div>

    <KunDivider margin="0 0 17px 0" />
  </div>
</template>
