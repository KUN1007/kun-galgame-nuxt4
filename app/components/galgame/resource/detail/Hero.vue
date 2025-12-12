<script setup lang="ts">
import {
  KUN_GALGAME_RESOURCE_LANGUAGE_MAP,
  KUN_GALGAME_RESOURCE_PLATFORM_MAP,
  KUN_GALGAME_RESOURCE_TYPE_MAP
} from '~/constants/galgame'

const props = defineProps<{
  galgame: GalgameResourceSummary
}>()

const galgameName = computed(() => getPreferredLanguageText(props.galgame.name))

const typeLabels = computed(() => {
  if (!props.galgame.type.length) return ['暂无数据']
  return props.galgame.type.map(
    (type) => KUN_GALGAME_RESOURCE_TYPE_MAP[type] || type
  )
})

const languageLabels = computed(() => {
  if (!props.galgame.language.length) return ['暂无数据']
  return props.galgame.language.map(
    (lang) => KUN_GALGAME_RESOURCE_LANGUAGE_MAP[lang] || lang
  )
})

const platformLabels = computed(() => {
  if (!props.galgame.platform.length) return ['暂无数据']
  return props.galgame.platform.map(
    (platform) => KUN_GALGAME_RESOURCE_PLATFORM_MAP[platform] || platform
  )
})
</script>

<template>
  <KunCard :is-hoverable="false" :is-transparent="false">
    <div class="grid grid-cols-1 gap-3 md:grid-cols-4 lg:grid-cols-5">
      <div class="relative md:col-span-2">
        <KunImage
          data-kun-lazy-image="true"
          class="max-h-70 rounded-lg object-cover"
          :style="{ aspectRatio: '16/9' }"
          :src="galgame.banner"
          loading="lazy"
          :alt="getPreferredLanguageText(galgame.name)"
        />

        <KunBadge
          :color="galgame.contentLimit === 'sfw' ? 'success' : 'danger'"
          class-name="absolute top-2 left-2"
          variant="solid"
        >
          {{ props.galgame.contentLimit.toUpperCase() }}
        </KunBadge>
      </div>

      <div class="flex w-full flex-col gap-3 md:col-span-2 lg:col-span-3">
        <div>
          <h1 class="text-2xl font-bold">
            {{ galgameName }}
            <KunBadge
              class-name="ml-2 -translate-y-1"
              :color="galgame.contentLimit === 'all' ? 'success' : 'danger'"
            >
              {{ galgame.contentLimit === 'all' ? '全年龄' : 'R18' }}
            </KunBadge>
          </h1>
          <p class="text-default-500 mt-1 text-sm">
            {{
              `最近更新 ${formatTimeDifference(galgame.resourceUpdateTime)} · ${galgame.view.toLocaleString()} 次浏览`
            }}
          </p>
        </div>

        <div class="flex flex-wrap gap-3">
          <div>
            <p class="text-default-500 text-xs tracking-wide uppercase">
              支持下载的类型
            </p>
            <div class="mt-1 flex flex-wrap gap-1">
              <KunBadge v-for="type in typeLabels" :key="type" variant="flat">
                {{ type }}
              </KunBadge>
            </div>
          </div>

          <div>
            <p class="text-default-500 text-xs tracking-wide uppercase">
              支持下载的语言
            </p>
            <div class="mt-1 flex flex-wrap gap-1">
              <KunBadge
                v-for="lang in languageLabels"
                :key="lang"
                variant="flat"
              >
                {{ lang }}
              </KunBadge>
            </div>
          </div>

          <div>
            <p class="text-default-500 text-xs tracking-wide uppercase">
              支持下载的平台
            </p>
            <div class="mt-1 flex flex-wrap gap-1">
              <KunBadge
                v-for="platform in platformLabels"
                :key="platform"
                variant="flat"
              >
                {{ platform }}
              </KunBadge>
            </div>
          </div>
        </div>

        <div class="mt-auto flex flex-wrap items-center justify-end gap-2">
          <KunButton variant="flat" href="/galgame"> 浏览更多资源 </KunButton>
          <KunButton :href="`/galgame/${galgame.id}`">
            查看这个 Galgame 的更多资源
          </KunButton>
        </div>
      </div>
    </div>
  </KunCard>
</template>
