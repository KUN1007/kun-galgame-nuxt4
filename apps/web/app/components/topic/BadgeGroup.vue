<script setup lang="ts">
import { KUN_TOPIC_SECTION } from '~/constants/topic'
import { topicMiniAppsOf } from './miniapp/registry'
import type { KunUIColor, KunUISize } from '@kungal/ui-core'

const props = withDefaults(
  defineProps<{
    section: string[]
    upvoteTime?: Date | string | null
    hasBestAnswer?: boolean
    miniApps?: string[]
    isNSFWTopic?: boolean
    isNavToSection?: boolean
  }>(),
  {
    upvoteTime: null,
    isNavToSection: false
  }
)

const miniApps = computed(() => topicMiniAppsOf(props.miniApps))

const iconMap: Record<string, string> = {
  g: 'lucide:gamepad-2',
  t: 'lucide:drafting-compass',
  o: 'lucide:circle-ellipsis'
}

const sectionColors: Record<string, KunUIColor> = {
  g: 'primary',
  t: 'success',
  o: 'secondary'
}

const isRecentlyUpvoted = useState(`kun-recent-upvote-${useId()}`, () =>
  hourDiff(props.upvoteTime || 0, 24)
)
onMounted(() => {
  isRecentlyUpvoted.value = hourDiff(props.upvoteTime || 0, 24)
})

const handleClickSection = async (section: string) => {
  if (props.isNavToSection) {
    await navigateTo(`/section/${section}`)
  }
}
</script>

<template>
  <div class="flex flex-wrap items-center gap-2">
    <KunChip
      variant="solid"
      color="warning"
      v-if="upvoteTime && isRecentlyUpvoted"
    >
      <KunIcon name="lucide:sparkles" class="size-4 text-inherit" />
      <span class="text-inherit">该话题被推</span>
    </KunChip>

    <span v-if="hasBestAnswer" class="flex gap-1">
      <KunChip variant="solid" color="success">
        <KunIcon name="lucide:bookmark-check" class="size-4 text-inherit" />
        有解答
      </KunChip>
    </span>

    <span v-for="app in miniApps" :key="app.key" class="flex gap-1">
      <KunChip variant="solid" :color="app.badgeColor">
        <KunIcon :name="app.icon" class="size-4 text-inherit" />
        {{ app.badge }}
      </KunChip>
    </span>

    <span v-if="isNSFWTopic" class="flex gap-1">
      <KunChip variant="solid" color="primary" class-name="bg-orange-600">
        <KunIcon name="uil:18-plus" class="size-4 text-inherit" />
        NSFW 话题
      </KunChip>
    </span>

    <span class="flex gap-1">
      <KunChip
        v-for="(sec, index) in props.section"
        :key="index"
        :color="sectionColors[sec.toLowerCase()[0]!]"
        @click="handleClickSection(sec.toLowerCase())"
        :class-name="cn(props.isNavToSection ? 'cursor-pointer' : '')"
      >
        <KunIcon
          :name="iconMap[sec.toLowerCase()[0]!]"
          class="size-4 text-inherit"
        />
        {{ KUN_TOPIC_SECTION[sec] }}
      </KunChip>
    </span>
  </div>
</template>
