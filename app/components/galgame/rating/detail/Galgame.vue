<script setup lang="ts">
import { KUN_GALGAME_AGE_LIMIT_MAP } from '~/constants/galgame'

defineProps<{
  galgame: GalgameRatingGalgameInfo
}>()

const getLanguageName = (langCode: string) => {
  // TODO: support more language names
  const map: Record<string, string> = {
    'ja-jp': '日本語',
    'en-us': 'English',
    'zh-cn': '简体中文',
    'zh-tw': '繁体中文'
  }
  return map[langCode.toLowerCase()] || langCode
}
</script>

<template>
  <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
    <div
      className="relative rounded-lg w-full h-full overflow-hidden md:col-span-1 aspect-video"
    >
      <KunImage
        data-kun-lazy-image="true"
        class="size-full object-cover"
        :src="galgame.banner"
        loading="lazy"
        :alt="getPreferredLanguageText(galgame.name)"
      />
    </div>

    <div class="space-y-3">
      <KunLink :to="`/galgame/${galgame.id}`" underline="none">
        <h1
          class="text-content hover:text-primary text-lg font-bold transition-colors sm:text-2xl"
        >
          {{ `${getPreferredLanguageText(galgame.name)}` }}
        </h1>
      </KunLink>

      <div class="text-default-500 flex items-center gap-3">
        <div class="flex items-center gap-2">
          <KunIcon class-name="text-warning text-2xl" name="lucide:lollipop" />
          <span class="text-warning text-xl font-bold">
            {{
              galgame.ratingCount
                ? (galgame.rating / galgame.ratingCount).toFixed(1)
                : '0.0'
            }}
          </span>
        </div>
        <span class="bg-default-300 h-3 w-px" />
        <div class="flex items-center gap-2">
          <KunIcon name="lucide:users" />
          <span>{{ galgame.ratingCount }} 人评分</span>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <span
          class="text-default-500 dark:text-default-400 text-sm font-medium"
        >
          年龄限制
        </span>
        <KunTooltip
          position="left"
          :text="KUN_GALGAME_AGE_LIMIT_MAP[galgame.ageLimit]"
        >
          <KunBadge
            variant="flat"
            :color="galgame.ageLimit === 'all' ? 'success' : 'danger'"
          >
            {{ galgame.ageLimit === 'all' ? '全年龄' : 'R18' }}
          </KunBadge>
        </KunTooltip>

        <span class="bg-default-300 h-3 w-px" />

        <span
          class="text-default-500 dark:text-default-400 text-sm font-medium"
        >
          原始语言
        </span>
        <KunBadge color="warning" variant="flat">
          {{ getLanguageName(galgame.originalLanguage) }}
        </KunBadge>
      </div>

      <dl>
        <GalgameDetailOfficial :official="galgame.official" />
      </dl>
    </div>
  </div>
</template>
