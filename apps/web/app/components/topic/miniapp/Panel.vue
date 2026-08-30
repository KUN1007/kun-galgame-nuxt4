<script setup lang="ts">
import { ref } from 'vue'
import { TOPIC_MINI_APPS, type TopicMiniAppKey } from './registry'
import type { KunUIColor } from '@kungal/ui-core'

const emits = defineEmits<{
  create: [key: TopicMiniAppKey]
}>()

// Tailwind only emits classes it can see as literals, so the tint cannot be
// built from app.badgeColor at runtime.
const ICON_TINT: Partial<Record<KunUIColor, string>> = {
  primary: 'bg-primary-500/10 text-primary-600 dark:text-primary-400',
  secondary: 'bg-secondary-500/10 text-secondary-600 dark:text-secondary-400'
}

const openHelp = ref<TopicMiniAppKey | ''>('')

const toggleHelp = (key: TopicMiniAppKey) => {
  openHelp.value = openHelp.value === key ? '' : key
}
</script>

<template>
  <KunCard
    :is-hoverable="false"
    :is-transparent="false"
    content-class="space-y-3"
  >
    <div class="flex items-center gap-2">
      <KunIcon name="lucide:blocks" class="text-default-500" />
      <h2 class="text-sm font-semibold">话题小程序</h2>
      <span class="text-default-400 text-xs">只有你和管理员看得到这里</span>
    </div>

    <div class="grid items-start gap-3 sm:grid-cols-2">
      <div
        v-for="app in TOPIC_MINI_APPS"
        :key="app.key"
        class="border-default-200 hover:border-default-300 rounded-xl border p-3 transition-colors"
      >
        <div class="flex items-start gap-3">
          <span
            :class="
              cn(
                'flex size-10 shrink-0 items-center justify-center rounded-lg',
                ICON_TINT[app.badgeColor] ?? 'bg-default-200 text-default-600'
              )
            "
          >
            <KunIcon :name="app.icon" class="text-xl text-inherit" />
          </span>

          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-1">
              <span class="font-medium">{{ app.label }}</span>
              <KunTooltip :text="`${app.label}怎么用`">
                <KunButton
                  size="sm"
                  variant="light"
                  color="default"
                  :is-icon-only="true"
                  :aria-expanded="openHelp === app.key"
                  @click="toggleHelp(app.key)"
                >
                  <KunIcon name="lucide:circle-help" />
                </KunButton>
              </KunTooltip>
            </div>
            <p class="text-default-500 text-xs">{{ app.tagline }}</p>
          </div>

          <KunButton
            size="sm"
            variant="flat"
            :color="app.badgeColor"
            @click="emits('create', app.key)"
          >
            <KunIcon name="lucide:plus" class="text-inherit" />
            创建
          </KunButton>
        </div>

        <ol
          v-if="openHelp === app.key"
          class="text-default-500 border-default-200 mt-3 list-decimal space-y-1 border-t pt-3 pl-5 text-xs"
        >
          <li v-for="(line, index) in app.help" :key="index">{{ line }}</li>
        </ol>
      </div>
    </div>
  </KunCard>
</template>
