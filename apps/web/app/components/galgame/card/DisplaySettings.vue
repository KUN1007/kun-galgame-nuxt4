<script setup lang="ts">
import { storeToRefs } from 'pinia'

const {
  showPlatform,
  showRating,
  showViewLike,
  showLanguage,
  showNsfwBadge,
  showPublisher,
  showJapaneseName,
  isOpenInNewTab
} = storeToRefs(usePersistGalgameCardStore())
</script>

<template>
  <div class="bg-default-50 space-y-4 rounded-lg border p-3">
    <div class="text-primary border-b pb-1 text-sm font-semibold">卡片四角</div>
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
      <KunSwitch v-model="showPlatform" label="左上角 · 游戏平台" />
      <KunSwitch v-model="showRating" label="右上角 · 总评分" />
      <KunSwitch v-model="showViewLike" label="左下角 · 浏览 / 点赞" />
      <KunSwitch v-model="showLanguage" label="右下角 · 可用语言" />
    </div>

    <div class="text-primary border-b pb-1 text-sm font-semibold">其它</div>
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
      <KunSwitch v-model="showNsfwBadge" label="显示 NSFW 角标" />
      <KunSwitch v-model="showPublisher" label="底部发布者与时间" />
      <KunSwitch v-model="showJapaneseName" label="标题下显示另一个名称" />
      <KunSwitch v-model="isOpenInNewTab" label="在新页面打开卡片" />
    </div>
  </div>
</template>
