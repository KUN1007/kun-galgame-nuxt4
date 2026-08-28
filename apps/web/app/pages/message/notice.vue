<script setup lang="ts">
definePageMeta({
  middleware: 'auth'
})

useKunDisableSeo('通知消息')

const pageData = reactive({
  page: 1,
  limit: 30,
  sort_order: 'desc'
})

const { data, status, refresh } = await useKunFetch<MessageList>('/message', {
  query: pageData
})

const isSettingOpen = ref(false)

onMounted(async () => {
  const hasUnreadMessage = data.value?.messages.some(
    (message) => message.status === 'unread'
  )
  if (hasUnreadMessage) {
    await kunFetch('/message/system/read', { method: 'PUT' })
  }
})
</script>

<template>
  <div class="flex w-full flex-col space-y-3" v-if="data">
    <header class="flex items-center justify-between gap-2">
      <div class="flex items-center gap-2">
        <KunButton
          size="lg"
          :is-icon-only="true"
          variant="light"
          href="/message"
        >
          <KunIcon name="lucide:chevron-left" />
        </KunButton>
        <h2 class="text-lg">通知</h2>
      </div>

      <KunPopover position="bottom-end">
        <template #trigger>
          <KunButton variant="light" size="sm">
            <span class="flex items-center gap-1">
              <KunIcon name="lucide:settings" />设置
            </span>
          </KunButton>
        </template>
        <div class="flex w-36 flex-col gap-1 p-2">
          <KunButton
            variant="light"
            color="default"
            size="sm"
            class-name="w-full justify-start gap-2"
            @click="isSettingOpen = true"
          >
            <KunIcon name="lucide:settings" />通知设置
          </KunButton>
        </div>
      </KunPopover>
    </header>

    <KunModal v-model="isSettingOpen" inner-class-name="max-w-lg w-[92vw]">
      <MessageNotificationPreference v-if="isSettingOpen" />
    </KunModal>

    <KunDivider />

    <KunOverlayScroll v-if="data.messages.length" class="h-full">
      <MessageAsideNotice
        v-for="message in data.messages"
        :key="message.id"
        :message="message"
        :refresh="refresh"
      />
    </KunOverlayScroll>

    <KunNull v-if="!data.total" />

    <KunPagination
      v-if="data.total"
      v-model:current-page="pageData.page"
      :total-page="Math.ceil(data.total / pageData.limit)"
      :is-loading="status === 'pending'"
    />
  </div>
</template>
