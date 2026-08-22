<script setup lang="ts">
defineProps<{ group: KunNewsGroup }>()
</script>

<template>
  <div class="space-y-1">
    <div class="flex flex-wrap items-center gap-x-3 gap-y-1">
      <KunIcon name="lucide:newspaper" class="text-primary size-4 shrink-0" />
      <span class="text-default-700 font-medium">{{ group.date }}</span>
      <span class="text-default-400 text-xs">
        {{ formatTimeDifference(group.items[0]!.published_at) }}
      </span>
      <KunUserChip
        v-if="group.source?.publisher"
        :user="group.source.publisher"
        size="sm"
        is-navigation
        class-name="ml-auto"
      />
      <KunLink
        v-else-if="group.source"
        :href="group.source.homepage_url"
        target="_blank"
        rel="noopener"
        color="default"
        size="sm"
        underline="hover"
        class-name="ml-auto"
      >
        {{ group.source.name }}
      </KunLink>
    </div>
    <p v-if="group.source?.attribution" class="text-default-400 text-xs">
      {{ group.source.attribution }}
    </p>
  </div>
</template>
