<script setup lang="ts">
const { page, limit, sortField, sortOrder, releasedFrom, releasedTo } =
  useGalgameFilters('popularity')

const route = useRoute()

const { data, status, refresh } = await useKunFetch<{
  galgames: GalgameCard[]
  total: number
}>(`/galgame`, {
  method: 'GET',
  query: {
    library: 'true',
    page,
    limit,
    sort_field: sortField,
    sort_order: sortOrder,
    released_from: releasedFrom,
    released_to: releasedTo
  },
  watch: false
})

const listPath = route.path
watch(
  () => route.fullPath,
  () => {
    if (route.path === listPath) {
      refresh()
    }
  }
)

const releaseLabel = (galgame: GalgameCard) => {
  if (galgame.release_date_tba) {
    return '发售日期待定'
  }
  return galgame.release_date?.slice(0, 10) || '发售日期未知'
}
</script>

<template>
  <div class="flex flex-col gap-3">
    <template v-if="data">
      <div class="z-10">
        <KunHeader name="Galgame 资料库">
          <template #endContent>
            <GalgameLibraryNav />
          </template>

          <template #description>
            <p class="text-default-500">
              资料库收录的全部 Galgame, 无论本站是否有下载资源。想找下载请前往
              <KunLink to="/galgame">Galgame 资源</KunLink>。
            </p>
          </template>
        </KunHeader>
      </div>

      <KunLoading :loading="status === 'pending'">
        <GalgameCard
          v-if="data.galgames.length"
          layout="portrait"
          :galgames="data.galgames"
        >
          <template #meta="{ galgame }">
            <p class="text-default-500 mt-1 text-xs">
              {{ releaseLabel(galgame) }}
            </p>
          </template>
        </GalgameCard>
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
