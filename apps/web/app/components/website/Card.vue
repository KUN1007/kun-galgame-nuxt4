<script setup lang="ts">
import { KUN_WEBSITE_STATUS_CHIP } from '~/constants/galgameWebsite'

const props = defineProps<{
  website: WebsiteCard
}>()

const statusChip = computed(() => KUN_WEBSITE_STATUS_CHIP[props.website.status])

const priceColor = computed(() => {
  const price = props.website.price
  if (price > 200) {
    return 'text-warning-500'
  }
  if (price > 100) {
    return 'text-success-600'
  }
  if (price > 0) {
    return 'text-default-700'
  }
  return 'text-danger-600'
})
</script>

<template>
  <KunCard
    :is-transparent="false"
    :href="`/website/${website.domain}`"
    class-name="group"
    content-class="space-y-3"
  >
    <div class="flex items-start space-x-4">
      <div class="flex-shrink-0">
        <KunImage
          :src="website.icon_url"
          :alt="website.name"
          :class="
            cn(
              'h-12 w-12 rounded-2xl object-cover',
              website.status === 'closed' && 'grayscale'
            )
          "
        />
      </div>
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2">
          <h3
            class="group-hover:text-primary-500 text-default-900 truncate text-lg font-semibold transition-colors"
          >
            {{ website.name }}
          </h3>
          <KunChip
            v-if="statusChip"
            :color="statusChip.color"
            class-name="shrink-0"
          >
            {{ statusChip.label }}
          </KunChip>
        </div>
        <p class="text-default-500 truncate font-mono text-sm">
          {{ website.domain }}
        </p>
      </div>
    </div>

    <p class="text-default-600 line-clamp-2 text-sm leading-relaxed">
      {{ website.description }}
    </p>

    <div class="text-default-500 flex items-center justify-between text-sm">
      <span>网站价值精算值</span>
      <span :class="cn('font-bold', priceColor)">
        {{ website.price }}
      </span>
    </div>
  </KunCard>
</template>
