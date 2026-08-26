<script setup lang="ts">
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

const { showKUNGalgameNoResource, showKUNGalgamePreferOriginalName } =
  storeToRefs(usePersistSettingsStore())

// Names are elected by the API, so the switch only takes effect on the next
// request — the same reason the NSFW switch reloads.
const preferOriginalName = ref(showKUNGalgamePreferOriginalName.value)
watch(preferOriginalName, (value) => {
  showKUNGalgamePreferOriginalName.value = value
  location.reload()
})
</script>

<template>
  <div class="space-y-4">
    <div class="space-y-2.5">
      <p class="text-default-700 text-sm font-semibold">名称显示</p>
      <div class="flex items-start justify-between gap-4">
        <div class="space-y-0.5">
          <p class="text-default-700">优先显示日语原名</p>
          <p class="text-default-500 text-sm">
            游戏标题、角色、声优与制作人员、会社、系列等所有名称都改为显示条目本身的原名，中文名退到副标题。开启或关闭后会自动刷新页面。角色属性等译名词表不受影响。
          </p>
        </div>
        <KunSwitch v-model="preferOriginalName" class="shrink-0" />
      </div>
    </div>

    <p class="text-default-500 text-sm">
      以下选项控制 Galgame 列表卡片的展示内容，更改后即时生效于全站所有 Galgame
      列表。
    </p>

    <div class="space-y-2.5">
      <p class="text-default-700 text-sm font-semibold">卡片四角</p>
      <div class="grid grid-cols-1 gap-2.5 sm:grid-cols-2">
        <KunSwitch v-model="showPlatform" label="左上角 · 游戏平台" />
        <KunSwitch v-model="showRating" label="右上角 · 总评分" />
        <KunSwitch v-model="showViewLike" label="左下角 · 浏览 / 点赞" />
        <KunSwitch v-model="showLanguage" label="右下角 · 可用语言" />
      </div>
    </div>

    <div class="space-y-2.5">
      <p class="text-default-700 text-sm font-semibold">其它</p>
      <div class="grid grid-cols-1 gap-2.5">
        <KunSwitch v-model="showNsfwBadge" label="显示 NSFW 角标" />
        <KunSwitch v-model="showPublisher" label="底部显示发布者与时间" />
        <KunSwitch v-model="showJapaneseName" label="标题下显示另一个名称" />
        <KunSwitch v-model="isOpenInNewTab" label="在新页面打开卡片" />
        <KunSwitch
          v-model="showKUNGalgameNoResource"
          label="排行与动态包含无资源的 Galgame"
        />
      </div>
    </div>
  </div>
</template>
