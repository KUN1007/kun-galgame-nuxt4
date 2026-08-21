<script setup lang="ts">
definePageMeta({
  middleware: 'permission',
  permissions: ['website.create', 'website.edit', 'website.delete']
})

useKunDisableSeo('网站资料库管理')

const tabs = [
  { value: 'category', textValue: '分类' },
  { value: 'tag', textValue: '标签与分组' }
]
const activeTab = useTabQuery('category')
</script>

<template>
  <div class="w-full space-y-6">
    <KunHeader
      name="网站资料库管理"
      description="维护 Galgame 网站资料库的分类与标签, 它们决定网站的归类、创建表单的选项以及价值精算值。"
    />

    <KunTab
      v-model="activeTab"
      :items="tabs"
      variant="underlined"
      color="primary"
      size="sm"
    />

    <AdminWebsiteTagSection v-if="activeTab === 'tag'" />
    <AdminWebsiteCategorySection v-else />
  </div>
</template>
