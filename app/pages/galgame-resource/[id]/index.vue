<script setup lang="ts">
import {
  KUN_GALGAME_RESOURCE_LANGUAGE_MAP,
  KUN_GALGAME_RESOURCE_PLATFORM_MAP,
  KUN_GALGAME_RESOURCE_TYPE_MAP
} from '~/constants/galgame'

const route = useRoute()
const resourceId = computed(() => Number((route.params as { id: string }).id))

const { data, refresh } = await useFetch(
  `/api/galgame-resource/${resourceId.value}`,
  {
    query: { resourceId },
    ...kungalgameResponseHandler
  }
)

const getResourceDescription = (value: GalgameResourcePageData) => {
  const typeLabel =
    KUN_GALGAME_RESOURCE_TYPE_MAP[value.resource.type] || value.resource.type
  const languageLabel =
    KUN_GALGAME_RESOURCE_LANGUAGE_MAP[value.resource.language] ||
    value.resource.language
  const platformLabel =
    KUN_GALGAME_RESOURCE_PLATFORM_MAP[value.resource.platform] ||
    value.resource.platform

  return `${typeLabel} · ${languageLabel} · ${platformLabel} · ${value.resource.size}`
}

if (data.value && data.value !== 'not found') {
  const titleBase = getPreferredLanguageText(data.value.galgame.name)
  if (data.value.galgame.contentLimit === 'nsfw') {
    useKunDisableSeo(titleBase)
  }

  const description = getResourceDescription(data.value)

  useKunSeoMeta({
    title: `${titleBase} | 资源详情`,
    description,
    ogImage: data.value.galgame.banner
  })
} else {
  useKunDisableSeo('未找到 Galgame 资源')
}
</script>

<template>
  <div v-if="data" class="space-y-3">
    <template v-if="data !== 'not found'">
      <GalgameResourceDetailHero :galgame="data.galgame" />

      <div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
        <GalgameResourceDetailPanel
          class="lg:col-span-2"
          :resource="data.resource"
          :refresh="refresh"
        />

        <GalgameResourceDetailRecommendations
          :recommendations="data.recommendations"
        />
      </div>
    </template>

    <KunNull
      v-else
      description="未找到对应的 Galgame 资源，或该资源已被移除。"
    />
  </div>
</template>
