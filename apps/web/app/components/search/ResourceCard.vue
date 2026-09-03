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

const props = defineProps<{
  resource: SearchResultResource
  keywords?: string
}>()

// Uploaders write the note in markdown, so the raw string is full of ###, **
// and bare links — three lines of syntax before the sentence that matched.
const note = computed(() => markdownToText(props.resource.note))
</script>

<template>
  <KunLink
    color="default"
    underline="none"
    class-name="flex-col items-start w-full gap-1.5"
    :to="`/galgame-resource/${resource.id}`"
  >
    <div class="flex w-full items-baseline gap-2">
      <KunIcon
        :name="
          GALGAME_RESOURCE_PLATFORM_ICON_MAP[resource.platform] ||
          'lucide:ellipsis'
        "
        class="text-primary size-3.5 shrink-0 self-center"
      />
      <h3 class="hover:text-primary min-w-0 flex-1 truncate font-medium">
        <SearchHighlight :text="resource.galgame_name" :keywords="keywords" />
      </h3>
      <span class="text-default-400 shrink-0 text-xs">
        <KunTime :time="resource.created" />
      </span>
    </div>

    <p v-if="note" class="text-default-600 line-clamp-2 w-full text-xs">
      <SearchHighlight :text="note" :keywords="keywords" />
    </p>

    <div
      class="text-default-500 flex w-full flex-wrap items-center gap-x-3 gap-y-1 text-xs"
    >
      <span class="flex items-center gap-1">
        <KunIcon
          :name="GALGAME_RESOURCE_TYPE_ICON_MAP[resource.type]"
          class="size-3.5"
        />
        {{ KUN_GALGAME_RESOURCE_TYPE_MAP[resource.type] }}
      </span>
      <span>{{ KUN_GALGAME_RESOURCE_LANGUAGE_MAP[resource.language] }}</span>
      <span>{{ KUN_GALGAME_RESOURCE_PLATFORM_MAP[resource.platform] }}</span>
      <span v-if="resource.size" class="flex items-center gap-1">
        <KunIcon name="lucide:database" class="size-3.5" />
        {{ resource.size }}
      </span>

      <span class="ml-auto flex shrink-0 items-center gap-3 tabular-nums">
        <span class="flex items-center gap-1">
          <KunIcon name="lucide:download" class="size-3.5" />
          {{ resource.download }}
        </span>
        <span class="flex items-center gap-1">
          <KunIcon name="lucide:eye" class="size-3.5" />
          {{ resource.view }}
        </span>
      </span>
    </div>
  </KunLink>
</template>
