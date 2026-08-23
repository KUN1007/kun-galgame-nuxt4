<script setup lang="ts">
import { storeToRefs } from 'pinia'

const {
  page,
  limit,
  type,
  language,
  platform,
  gameType,
  sortField,
  sortOrder,
  releasedFrom,
  releasedTo,
  releasedMonths,
  includeProviders,
  excludeOnlyProviders,
  minRatingCount,
  minRating
} = useGalgameFilters()

const { showKUNGalgameNoResource } = storeToRefs(usePersistSettingsStore())

const route = useRoute()

const { data, status, refresh } = await useKunFetch<{
  galgames: GalgameCard[]
  total: number
}>(`/galgame`, {
  method: 'GET',
  query: {
    page,
    limit,
    type,
    language,
    platform,
    game_type: gameType,
    sort_field: sortField,
    sort_order: sortOrder,
    released_from: releasedFrom,
    released_to: releasedTo,
    released_months: releasedMonths,
    include_providers: includeProviders,
    exclude_only_providers: excludeOnlyProviders,
    min_rating_count: minRatingCount,
    min_rating: minRating,
    show_no_resource: showKUNGalgameNoResource
  },
  watch: false
})

const listPath = route.path
watch([() => route.fullPath, showKUNGalgameNoResource], () => {
  if (route.path === listPath) {
    refresh()
  }
})
</script>

<template>
  <div class="flex flex-col gap-3">
    <template v-if="data">
      <div class="z-10">
        <KunHeader name="Galgame 资源资料库">
          <template #endContent>
            <GalgameCardNav :is-show-advanced="true" />
          </template>

          <template #description>
            <p class="text-default-500">
              浏览资料库中的全部 Galgame,
              打开任意作品即可发布资源。我们不是资源的提供者,
              我们只是资源的指路人。
            </p>
          </template>
        </KunHeader>
      </div>

      <KunLoading :loading="status === 'pending'">
        <GalgameCard v-if="data.galgames.length" :galgames="data.galgames" />
        <KunNull v-else description="没有找到符合条件的 Galgame" />
      </KunLoading>

      <KunCard
        v-if="data.galgames.length"
        :is-hoverable="false"
        :is-transparent="false"
        content-class="gap-3"
      >
        <KunPagination
          v-model:current-page="page"
          :total-page="Math.ceil(data.total / limit)"
          :is-loading="status === 'pending'"
        />
      </KunCard>
    </template>
  </div>
</template>
