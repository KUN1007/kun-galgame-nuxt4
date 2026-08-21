<script setup lang="ts">
import type { KUN_ADMIN_PAGE_ROUTE_TYPE } from '~/constants/admin'

useKunDisableSeo('管理系统')

const route = useRoute()
const pageType = computed(() => {
  const routeType = route.path.split('/').pop()
  return routeType as KUN_ADMIN_PAGE_ROUTE_TYPE
})

const { items } = useAdminNav()

const adminNavItems = computed(() =>
  items.value.map((item) => ({
    value: item.to ?? item.router!,
    textValue: item.label,
    icon: item.icon
  }))
)
</script>

<template>
  <div class="flex gap-3">
    <div class="hidden w-48 shrink-0 sm:block">
      <KunTab
        :model-value="pageType"
        :items="adminNavItems"
        orientation="vertical"
        variant="underlined"
        color="primary"
        size="lg"
        full-width
        @update:model-value="
          (value) =>
            navigateTo(value.startsWith('/') ? value : `/admin/${value}`)
        "
      />
    </div>

    <NuxtPage />
  </div>
</template>
