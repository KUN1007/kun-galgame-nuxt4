<script setup lang="ts">
definePageMeta({
  middleware: 'auth'
})

useKunDisableSeo('系统消息')

const pageData = reactive({
  page: 1,
  limit: 30
})

const { data } = await useKunFetch<MessageSystemMessage[]>('/message/admin', {
  query: pageData
})

onMounted(async () => {
  const hasUnreadMessage = data.value?.some((message) => !message.is_read)
  if (hasUnreadMessage) {
    await kunFetch('/message/admin/read', { method: 'PUT' })
  }
})
</script>

<template>
  <div class="flex w-full flex-col space-y-3" v-if="data">
    <header class="flex items-center gap-2">
      <KunButton size="lg" :is-icon-only="true" variant="light" href="/message">
        <KunIcon name="lucide:chevron-left" />
      </KunButton>
      <h2 class="text-lg">系统消息</h2>
    </header>

    <KunDivider />

    <KunOverlayScroll v-if="data.length" class="h-full">
      <MessageAsideSystem
        v-for="message in data"
        :key="message.id"
        :message="message"
      />
    </KunOverlayScroll>

    <KunNull v-if="!data.length" />
  </div>
</template>
