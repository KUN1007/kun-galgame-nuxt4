<script setup lang="ts">
import {
  KUN_GALGAME_TOOLSET_TYPE_MAP,
  KUN_GALGAME_TOOLSET_TYPE_ICON_MAP,
  KUN_GALGAME_TOOLSET_PLATFORM_MAP,
  KUN_GALGAME_TOOLSET_PLATFORM_ICON_MAP,
  KUN_GALGAME_TOOLSET_LANGUAGE_MAP,
  KUN_GALGAME_TOOLSET_VERSION_MAP
} from '~/constants/toolset'

defineProps<{
  items: ToolsetCard[]
  keywords?: string
}>()

const label = (map: Record<string, string>, key: string) => map[key] || key
</script>

<template>
  <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
    <KunCard
      v-for="t in items"
      :key="t.id"
      :href="`/toolset/${t.id}`"
      :is-transparent="false"
      :is-hoverable="true"
      padding="none"
      content-class="group"
    >
      <div class="flex w-full flex-col gap-2.5 p-3">
        <div class="flex items-start gap-2.5">
          <span
            class="bg-primary/10 text-primary flex size-9 shrink-0 items-center justify-center rounded-lg"
          >
            <KunIcon
              :name="
                KUN_GALGAME_TOOLSET_TYPE_ICON_MAP[t.type] || 'lucide:wrench'
              "
              class="size-5"
            />
          </span>

          <div class="min-w-0 flex-1">
            <h3
              class="group-hover:text-primary truncate font-medium transition-colors"
            >
              <SearchHighlight :text="t.name" :keywords="keywords" />
            </h3>
            <p class="text-default-500 truncate text-xs">
              {{ label(KUN_GALGAME_TOOLSET_TYPE_MAP, t.type) }}
              <span class="mx-1">·</span>
              {{ label(KUN_GALGAME_TOOLSET_VERSION_MAP, t.version) }}
            </p>
          </div>

          <span
            v-if="t.practicality_avg != null"
            class="text-secondary flex shrink-0 items-center gap-1 text-sm tabular-nums"
          >
            <KunIcon name="lucide:star" class="size-4" />
            {{ t.practicality_avg.toFixed(1) }}
          </span>
        </div>

        <div class="flex flex-wrap items-center gap-1.5">
          <KunChip size="xs" variant="flat" color="primary">
            <KunIcon
              :name="
                KUN_GALGAME_TOOLSET_PLATFORM_ICON_MAP[t.platform] ||
                'lucide:ellipsis'
              "
              class="size-3.5"
            />
            {{ label(KUN_GALGAME_TOOLSET_PLATFORM_MAP, t.platform) }}
          </KunChip>
          <KunChip size="xs" variant="flat">
            {{ label(KUN_GALGAME_TOOLSET_LANGUAGE_MAP, t.language) }}
          </KunChip>
        </div>

        <div
          class="border-default-200 text-default-500 flex items-center gap-2 border-t pt-2 text-xs"
        >
          <KunAvatar size="xs" :user="t.user" :is-navigation="false" />
          <span class="min-w-0 truncate">{{ t.user.name }}</span>

          <span class="ml-auto flex shrink-0 items-center gap-3 tabular-nums">
            <span class="flex items-center gap-1">
              <KunIcon name="lucide:eye" class="size-3.5" />
              {{ formatNumber(t.view) }}
            </span>
            <span class="flex items-center gap-1">
              <KunIcon name="lucide:download" class="size-3.5" />
              {{ formatNumber(t.download) }}
            </span>
            <span v-if="t.comment_count" class="flex items-center gap-1">
              <KunIcon name="lucide:message-square" class="size-3.5" />
              {{ t.comment_count }}
            </span>
          </span>
        </div>
      </div>
    </KunCard>
  </div>
</template>
