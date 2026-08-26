<script setup lang="ts">
import { storeToRefs } from 'pinia'

const {
  showPlatform,
  showRating,
  showViewLike,
  showUpdateTime,
  showNsfwBadge,
  showCompany,
  showJapaneseName,
  isOpenInNewTab
} = storeToRefs(usePersistGalgameCardStore())
</script>

<template>
  <div class="bg-default-50 space-y-4 rounded-lg border p-3">
    <div class="text-primary border-b pb-1 text-sm font-semibold">封面四角</div>
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
      <KunSwitch v-model="showPlatform" label="左上角 · 游戏平台" />
      <KunSwitch v-model="showNsfwBadge" label="右上角 · NSFW 角标" />
      <KunSwitch v-model="showViewLike" label="左下角 · 浏览 / 点赞" />
      <KunSwitch v-model="showUpdateTime" label="右下角 · 更新时间" />
    </div>

    <div class="text-primary border-b pb-1 text-sm font-semibold">卡片底部</div>
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
      <KunSwitch v-model="showCompany" label="左侧 · 制作会社" />
      <KunSwitch v-model="showRating" label="右侧 · 总评分" />
      <KunSwitch v-model="showJapaneseName" label="标题下显示另一个名称" />
      <KunSwitch v-model="isOpenInNewTab" label="在新页面打开卡片" />
    </div>
  </div>
</template>
