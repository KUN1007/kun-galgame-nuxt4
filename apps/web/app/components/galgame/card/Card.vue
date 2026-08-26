<script setup lang="ts" generic="T extends GalgameCard">
import { storeToRefs } from 'pinia'
import { GALGAME_RESOURCE_PLATFORM_ICON_MAP } from '~/constants/galgameResource'

const props = withDefaults(
  defineProps<{
    galgames: T[]
    isTransparent?: boolean
    columns?: 3 | 4
    layout?: 'landscape' | 'portrait'
  }>(),
  { columns: 4, layout: 'landscape' }
)

defineSlots<{
  meta?: (props: { galgame: T }) => unknown
}>()

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

const isPortrait = computed(() => props.layout === 'portrait')

const gridClass = computed(() => {
  if (isPortrait.value) {
    return 'grid grid-cols-2 gap-2 sm:grid-cols-3 sm:gap-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6'
  }
  return props.columns === 3
    ? 'grid grid-cols-2 gap-2 sm:grid-cols-2 sm:gap-3 lg:grid-cols-3 xl:grid-cols-3'
    : 'grid grid-cols-2 gap-2 sm:grid-cols-2 sm:gap-3 lg:grid-cols-3 xl:grid-cols-4'
})

const coverOf = (galgame: T) =>
  isPortrait.value
    ? getEffectivePortrait(galgame, { variant: 'mini' })
    : getEffectiveBanner(galgame, { variant: 'mini' })

const thumbhashOf = (galgame: T) =>
  isPortrait.value
    ? resolvePortraitThumbhash(galgame)
    : resolveBannerThumbhash(galgame)

const cardHref = (galgame: GalgameCard) =>
  galgame.id > 0 ? `/galgame/${galgame.id}` : undefined
</script>

<template>
  <div :class="gridClass">
    <KunCard
      :is-transparent="isTransparent"
      v-for="galgame in galgames"
      :key="galgame.catalog_id ?? galgame.id"
      :href="cardHref(galgame)"
      :target="isOpenInNewTab ? '_blank' : undefined"
      class-name="p-0"
    >
      <div class="relative overflow-hidden">
        <KunImage
          :src="coverOf(galgame)"
          loading="lazy"
          :alt="galgame.name"
          placeholder="/placeholder.webp"
          :thumbhash="thumbhashOf(galgame)"
          class="h-full w-full object-cover transition-transform duration-300"
          :style="{ aspectRatio: isPortrait ? '5/7' : '16/9' }"
        />

        <div
          v-if="
            showPlatform ||
            (showRating && galgame.rating_count) ||
            showNsfwBadge
          "
          class="absolute top-2 right-2 left-2 flex items-start gap-1"
        >
          <div v-if="showPlatform" class="flex flex-wrap gap-1">
            <template v-if="galgame.platform.length">
              <span
                v-for="(platform, i) in galgame.platform"
                :key="i"
                class="bg-background flex items-center justify-center rounded-full p-1.5 text-xs backdrop-blur-sm"
                :class="isPortrait ? 'size-6' : 'size-6 sm:size-8 sm:text-sm'"
              >
                <KunIcon
                  :name="GALGAME_RESOURCE_PLATFORM_ICON_MAP[platform]"
                  class="h-4 w-4"
                />
              </span>
            </template>
            <span
              v-else-if="!isPortrait"
              class="bg-background rounded-full px-3 py-1 text-xs backdrop-blur-sm sm:text-sm"
            >
              准备中
            </span>
          </div>

          <div class="ml-auto flex flex-col items-end gap-1">
            <span
              v-if="showRating && galgame.rating_count"
              class="bg-background flex items-center gap-1 rounded-full px-2 py-1 text-xs font-medium backdrop-blur-sm"
              :class="!isPortrait && 'sm:text-sm'"
            >
              <KunIcon name="lucide:star" class="text-warning" />
              {{ galgame.rating?.toFixed(1) }}
            </span>

            <KunChip
              v-if="showNsfwBadge"
              variant="solid"
              :size="isPortrait ? 'xs' : undefined"
              :color="galgame.content_limit === 'sfw' ? 'success' : 'danger'"
            >
              {{ galgame.content_limit.toLocaleUpperCase() }}
            </KunChip>
          </div>
        </div>

        <!-- SANCTIONED EXCEPTION to 铁律 #1 (no gradients): a bottom-to-top
             black scrim so the caption stays legible over an arbitrary cover.
             Listed in CLAUDE.md; do NOT remove it in a no-gradient sweep. -->
        <div
          v-if="(showViewLike || showLanguage) && galgame.is_on_forum !== false"
          class="absolute right-0 bottom-0 left-0 flex items-center gap-2 bg-gradient-to-t from-black/60 to-transparent p-2 text-xs transition-opacity duration-300"
          :class="!isPortrait && 'sm:text-sm'"
        >
          <div v-if="showViewLike" class="flex gap-3">
            <span class="flex items-center gap-1">
              <KunIcon class="text-white" name="lucide:eye" />
              <span class="text-white">{{ galgame.view }}</span>
            </span>

            <span class="flex items-center gap-1">
              <KunIcon class="text-white" name="lucide:thumbs-up" />
              <span class="text-white">{{ galgame.like_count }}</span>
            </span>
          </div>

          <div v-if="showLanguage" class="ml-auto flex gap-2">
            <span
              class="text-white"
              v-for="(lang, i) in galgame.language"
              :key="i"
            >
              {{ lang.substring(0, 2).toUpperCase() }}
            </span>
          </div>
        </div>
      </div>

      <div
        class="flex flex-auto flex-col"
        :class="isPortrait ? 'p-2' : 'p-2 sm:p-3'"
      >
        <h2
          class="hover:text-primary line-clamp-2 font-medium transition-colors"
          :class="isPortrait && 'text-sm'"
        >
          {{ galgame.name }}
        </h2>

        <p
          v-if="showJapaneseName && galgame.name_original"
          class="text-default-500 mt-1 line-clamp-1"
          :class="isPortrait ? 'text-xs' : 'text-sm'"
        >
          {{ galgame.name_original }}
        </p>

        <slot name="meta" :galgame="galgame" />

        <div
          v-if="
            showPublisher && galgame.is_on_forum !== false && galgame.user.id
          "
          class="text-default-600 mt-auto flex min-w-0 items-center gap-1 pt-3"
          :class="isPortrait ? 'text-xs' : 'text-sm'"
        >
          <KunAvatar
            :disable-floating="true"
            :user="galgame.user"
            size="xs"
            :is-navigation="false"
          />
          <span class="truncate">{{ galgame.user.name }}</span>
          <span v-if="!isPortrait">·</span>
          <KunTime
            v-if="!isPortrait"
            class="shrink-0"
            :time="galgame.resource_update_time"
          />
        </div>
      </div>
    </KunCard>
  </div>
</template>
