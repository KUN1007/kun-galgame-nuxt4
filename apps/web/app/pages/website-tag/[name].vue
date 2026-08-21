<script setup lang="ts">
definePageMeta({ key: (route) => route.path })

const route = useRoute()
const tagName = computed(() => {
  return (route.params as { name: string }).name
})

const canManageTaxonomy = useCan('website.edit')

const { data } = await useKunFetch<WebsiteTagDetail>(
  `/website-tag/${tagName.value}`,
  {
    watch: false,
    query: { name: tagName.value }
  }
)

if (data.value) {
  useKunSeoMeta({
    title: `${data.value.label}的 Galgame 网站`,
    description: data.value.description,
    articlePublishedTime: data.value.created.toString(),
    articleModifiedTime: data.value.updated.toString()
  })
} else {
  useKunDisableSeo('未找到该网站标签')
}
</script>

<template>
  <div v-if="data" class="space-y-6">
    <KunHeader
      :name="`${data.label}的 Galgame 网站`"
      :description="data.description"
    >
      <template #endContent>
        <div class="space-y-3">
          <div class="flex items-center space-x-3">
            <KunChip color="primary">标签价值 {{ data.level }}</KunChip>

            <KunChip>
              更新于 <KunTime :time="data.updated" type="date" show-year />
            </KunChip>
          </div>

          <div v-if="canManageTaxonomy" class="flex justify-end">
            <KunButton variant="light" href="/admin/website?tab=tag">
              管理标签
            </KunButton>
          </div>
        </div>
      </template>
    </KunHeader>

    <div v-if="data.websites.length">
      <div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        <WebsiteCard
          v-for="website in data.websites"
          :key="website.id"
          :website="website"
        />
      </div>
    </div>

    <KunNull v-else :description="`${data.label} 标签下暂无网站`" />
  </div>
</template>
