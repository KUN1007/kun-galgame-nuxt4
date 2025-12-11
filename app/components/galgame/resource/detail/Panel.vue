<script setup lang="ts">
import {
  KUN_GALGAME_RESOURCE_LANGUAGE_MAP,
  KUN_GALGAME_RESOURCE_PLATFORM_MAP,
  KUN_GALGAME_RESOURCE_TYPE_MAP
} from '~/constants/galgame'
import {
  GALGAME_RESOURCE_TYPE_ICON_MAP,
  GALGAME_RESOURCE_PLATFORM_ICON_MAP
} from '~/constants/galgameResource'

const props = defineProps<{
  resource: GalgameResourceDetails
  refresh: () => Promise<void>
}>()

const resourceTypeLabel = computed(
  () =>
    KUN_GALGAME_RESOURCE_TYPE_MAP[props.resource.type] || props.resource.type
)
const resourceLanguageLabel = computed(
  () =>
    KUN_GALGAME_RESOURCE_LANGUAGE_MAP[props.resource.language] ||
    props.resource.language
)
const resourcePlatformLabel = computed(
  () =>
    KUN_GALGAME_RESOURCE_PLATFORM_MAP[props.resource.platform] ||
    props.resource.platform
)

const isResourceExpired = computed(() => props.resource.status === 1)
</script>

<template>
  <KunCard
    :is-hoverable="false"
    :is-transparent="false"
    content-class="space-y-4 h-full"
  >
    <KunHeader
      name="Galgame 资源详情"
      description="查看该 Galgame 的下载信息, 若失效请及时向资源发布者反馈或贡献新的下载资源。"
      scale="h3"
    />

    <div class="flex flex-wrap gap-2">
      <KunBadge color="primary">
        <KunIcon :name="GALGAME_RESOURCE_TYPE_ICON_MAP[resource.type]" />
        {{ resourceTypeLabel }}
      </KunBadge>
      <KunBadge color="secondary">
        <KunIcon name="lucide:languages" />
        {{ resourceLanguageLabel }}
      </KunBadge>
      <KunBadge color="success">
        <KunIcon
          :name="GALGAME_RESOURCE_PLATFORM_ICON_MAP[resource.platform]"
        />
        {{ resourcePlatformLabel }}
      </KunBadge>
      <KunBadge color="warning">
        <KunIcon name="lucide:database" />
        {{ resource.size }}
      </KunBadge>
      <KunBadge color="default">
        <KunIcon name="lucide:download" />
        {{ `${resource.download} 次下载` }}
      </KunBadge>
      <KunBadge :color="isResourceExpired ? 'danger' : 'success'">
        {{ isResourceExpired ? '待修复' : '可用' }}
      </KunBadge>
    </div>

    <KunDivider />

    <GalgameResourceDetailInfo :details="resource" :refresh="refresh" />
  </KunCard>
</template>
