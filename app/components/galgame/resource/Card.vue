<script setup lang="ts">
import {
  GALGAME_RESOURCE_TYPE_ICON_MAP,
  GALGAME_RESOURCE_PLATFORM_ICON_MAP
} from '~/constants/galgameResource'
import {
  KUN_GALGAME_RESOURCE_TYPE_MAP,
  KUN_GALGAME_RESOURCE_LANGUAGE_MAP,
  KUN_GALGAME_RESOURCE_PLATFORM_MAP
} from '~/constants/galgame'

defineProps<{
  resource: GalgameResourceCard
}>()
</script>

<template>
  <KunCard
    :is-transparent="true"
    :is-hoverable="true"
    :is-pressable="true"
    :to="`/galgame-resource/${resource.id}`"
    :dark-border="true"
    content-class="space-y-2"
  >
    <div class="flex flex-wrap items-center gap-2">
      <KunBadge size="sm" variant="flat" color="primary">
        <KunIcon
          :name="GALGAME_RESOURCE_PLATFORM_ICON_MAP[resource.platform]"
          class="text-primary h-4 w-4"
        />
        {{ KUN_GALGAME_RESOURCE_PLATFORM_MAP[resource.platform] }}
      </KunBadge>

      <KunBadge color="warning">
        <KunIcon name="lucide:database" />
        {{ resource.size }}
      </KunBadge>

      <span class="text-default-500 text-sm">
        {{ formatTimeDifference(resource.created) }}
      </span>
    </div>

    <div class="space-y-2">
      <h3 class="hover:text-primary line-clamp-3 break-all transition-colors">
        {{ getPreferredLanguageText(resource.galgameName) }}
      </h3>

      <div class="flex items-center justify-between">
        <div class="text-default-700 flex gap-4 text-sm">
          <span class="flex items-center gap-1">
            <KunIcon
              class="icon"
              :name="GALGAME_RESOURCE_TYPE_ICON_MAP[resource.type]"
            />
            {{ KUN_GALGAME_RESOURCE_TYPE_MAP[resource.type] }}
          </span>

          {{ KUN_GALGAME_RESOURCE_LANGUAGE_MAP[resource.language] }}
        </div>

        <div class="text-default-500 flex items-center gap-1 text-sm">
          <KunIcon name="lucide:download" class="h-4 w-4" />
          {{ resource.download }}
        </div>
      </div>
    </div>
  </KunCard>
</template>
