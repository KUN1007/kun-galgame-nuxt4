<script setup lang="ts">
import { computed } from 'vue'
import { TOPIC_MINI_APPS, type TopicMiniAppKey } from './registry'
import type { KunUIColor } from '@kungal/ui-core'

const props = defineProps<{
  appKey: TopicMiniAppKey
  title: string
  meta?: (string | false | null | undefined)[]
  status: string
  statusColor: KunUIColor
}>()

// Tailwind only emits classes it can see as literals, so the tint cannot be
// built from app.badgeColor at runtime.
const ICON_TINT: Partial<Record<KunUIColor, string>> = {
  primary: 'bg-primary-500/10 text-primary-600 dark:text-primary-400',
  secondary: 'bg-secondary-500/10 text-secondary-600 dark:text-secondary-400'
}

const app = computed(
  () => TOPIC_MINI_APPS.find((item) => item.key === props.appKey)!
)
const tint = computed(
  () => ICON_TINT[app.value.badgeColor] ?? 'bg-default-200 text-default-600'
)
const metaLine = computed(() => (props.meta ?? []).filter(Boolean) as string[])
</script>

<template>
  <div class="flex items-start justify-between gap-3">
    <div class="flex min-w-0 items-start gap-3">
      <span
        :class="
          cn(
            'flex size-9 shrink-0 items-center justify-center rounded-lg',
            tint
          )
        "
      >
        <KunIcon :name="app.icon" class="text-lg text-inherit" />
      </span>

      <div class="min-w-0">
        <h3 class="text-base leading-snug font-semibold">{{ title }}</h3>
        <p
          v-if="metaLine.length"
          class="text-default-500 mt-1 flex flex-wrap items-center gap-x-2 text-xs"
        >
          <template v-for="(item, index) in metaLine" :key="index">
            <span v-if="index" class="text-default-300" aria-hidden="true">
              ·
            </span>
            <span>{{ item }}</span>
          </template>
        </p>
      </div>
    </div>

    <KunChip variant="flat" size="sm" :color="statusColor">
      {{ status }}
    </KunChip>
  </div>
</template>
