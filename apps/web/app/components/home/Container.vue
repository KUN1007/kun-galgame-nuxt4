<script setup lang="ts">
import { isNewsFeedTab } from '~/constants/activity'

const settings = usePersistSettingsStore()
const { feedTabs } = storeToRefs(settings)
const tabItems = computed(() =>
  feedTabs.value.map((t) => ({ value: t.id, textValue: t.name, icon: t.icon }))
)
const activeTab = useTabQuery(feedTabs.value[0]?.id ?? 'all')
const currentTab = computed(
  () =>
    feedTabs.value.find((t) => t.id === activeTab.value) ?? feedTabs.value[0]
)
const activeTypes = computed(() => (currentTab.value?.kinds ?? []).join(','))
watchEffect(() => {
  if (
    feedTabs.value.length &&
    !feedTabs.value.some((t) => t.id === activeTab.value)
  ) {
    activeTab.value = feedTabs.value[0]!.id
  }
})
</script>

<template>
  <div
    class="grid grid-cols-1 items-start gap-4 sm:grid-cols-[7rem_minmax(0,1fr)] lg:grid-cols-[7rem_minmax(0,1fr)_18rem] xl:grid-cols-[7rem_minmax(0,1fr)_20rem]"
  >
    <!-- 铁律 #11: keep this template's SINGLE real root element. A sibling
         or a leading comment at the template root is itself a root node,
         trips Nuxt's "does not have a single root node" warning, and
         silently drops the page-transition enter animation. -->

    <div class="sm:hidden">
      <KunTab
        v-model="activeTab"
        :items="tabItems"
        variant="underlined"
        color="primary"
        scrollable
      />
    </div>

    <!-- Columns are placed EXPLICITLY (col-start), never auto-placed.
         Both feeds take a top-level `await`, so switching tabs mounts one
         inside the page's already-resolved Suspense boundary, where Vue
         renders an empty comment placeholder instead of the component until
         the fetch lands. Auto-placement then pulled the aside out of the
         20rem right column into the 1fr middle one, and on a slow link the
         home page stayed that way until the feed arrived. -->
    <div class="sticky top-20 hidden self-start sm:col-start-1 sm:block">
      <KunTab
        v-model="activeTab"
        :items="tabItems"
        orientation="vertical"
        variant="underlined"
        color="primary"
        full-width
      />
    </div>

    <HomeNewsFeed
      v-if="isNewsFeedTab(currentTab)"
      class="min-w-0 sm:col-start-2"
    />
    <HomeActivityFeed
      v-else
      :tab-id="activeTab"
      :types="activeTypes"
      class="sm:col-start-2"
    />

    <aside
      class="sticky top-20 hidden h-[calc(100dvh-6rem)] flex-col self-start lg:col-start-3 lg:flex"
    >
      <div class="space-y-4">
        <HomeCarousel />
        <HomeAsideHelp />
      </div>
      <HomeFooter class="mt-auto" />
    </aside>
  </div>
</template>
