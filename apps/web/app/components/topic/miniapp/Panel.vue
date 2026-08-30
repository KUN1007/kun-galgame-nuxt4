<script setup lang="ts">
import { ref } from 'vue'
import { TOPIC_MINI_APPS, type TopicMiniAppKey } from './registry'

const emits = defineEmits<{
  create: [key: TopicMiniAppKey]
}>()

const openHelp = ref<string>('')

const toggleHelp = (key: string) => {
  openHelp.value = openHelp.value === key ? '' : key
}
</script>

<template>
  <KunCard
    :is-hoverable="false"
    :is-transparent="false"
    content-class="space-y-3"
  >
    <h2>话题小程序</h2>

    <div class="flex flex-wrap items-center gap-3">
      <template v-for="app in TOPIC_MINI_APPS" :key="app.key">
        <div class="flex items-center">
          <KunButton
            size="sm"
            variant="bordered"
            class="flex-col px-4"
            @click="emits('create', app.key)"
          >
            <KunIcon :name="app.icon" class="text-2xl" />
            {{ app.label }}
          </KunButton>

          <KunTooltip :text="`${app.label}怎么用`">
            <KunButton
              size="sm"
              variant="light"
              color="default"
              :is-icon-only="true"
              @click="toggleHelp(app.key)"
            >
              <KunIcon name="lucide:circle-help" />
            </KunButton>
          </KunTooltip>
        </div>
      </template>
    </div>

    <template v-for="app in TOPIC_MINI_APPS" :key="`help-${app.key}`">
      <KunInfo v-if="openHelp === app.key" :title="`${app.label}使用帮助`">
        <div class="space-y-1 text-sm">
          <p v-for="(line, index) in app.help" :key="index">
            {{ index + 1 }}. {{ line }}
          </p>
        </div>
      </KunInfo>
    </template>
  </KunCard>
</template>
