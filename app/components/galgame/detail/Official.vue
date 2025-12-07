<script setup lang="ts">
import { KUN_GALGAME_OFFICIAL_LANGUAGE_MAP } from '~/constants/galgameOfficial'

defineProps<{
  official: GalgameOfficialItem[]
}>()

const getOfficialCategoryInfo = (category: string) => {
  switch (category) {
    case 'company':
      return {
        text: '公司',
        class: 'bg-primary-100 text-primary-800'
      }
    case 'individual':
      return {
        text: '个人',
        class: 'bg-success-100 text-success-800'
      }
    case 'amateur':
      return {
        text: '同人',
        class: 'bg-secondary-100 text-secondary-800'
      }
    default:
      return {
        text: category,
        class: 'bg-default-100 text-default-800'
      }
  }
}
</script>

<template>
  <div>
    <dt class="text-default-500 text-sm font-medium">制作方</dt>
    <dd class="mt-1.5 space-y-3">
      <div class="space-y-2" v-for="item in official" :key="item.id">
        <KunTooltip :text="`该会社制作了 ${item.galgameCount} 个 Galgame`">
          <KunLink
            :to="`/galgame-official/${item.id}`"
            underline="none"
            class-name="text-foreground hover:text-primary text-base font-semibold"
          >
            {{ item.name }}
            <KunBadge size="xs">
              {{ `+ ${item.galgameCount}` }}
            </KunBadge>
          </KunLink>
        </KunTooltip>

        <div class="mt-1 flex items-center justify-between">
          <div class="flex items-center gap-x-2">
            <KunBadge
              size="xs"
              class-name="rounded-md"
              :color="
                item.category === 'company'
                  ? 'primary'
                  : item.category === 'individual'
                    ? 'secondary'
                    : 'success'
              "
            >
              {{ getOfficialCategoryInfo(item.category).text }}
            </KunBadge>
            <span class="text-default-500 dark:text-default-400 text-xs">
              {{ KUN_GALGAME_OFFICIAL_LANGUAGE_MAP[item.lang] || item.lang }}
            </span>
          </div>

          <KunLink
            :is-show-anchor-icon="true"
            target="_blank"
            :to="item.link"
            size="sm"
            underline="hover"
            rel="noopener noreferrer"
          >
            官方网站
          </KunLink>
        </div>

        <div
          v-if="item.alias.length"
          class="text-default-500 flex flex-wrap gap-2"
        >
          <KunBadge
            size="xs"
            color="success"
            v-for="(a, index) in item.alias"
            :key="index"
          >
            {{ a }}
          </KunBadge>
        </div>
      </div>
    </dd>
  </div>
</template>
