<script setup lang="ts">
const routeName = computed(() => useRoute().name)

const { data: systemNav } = useKunFetch<ChatMessageAsideItem[]>(
  '/message/nav/system'
)
const { data: contactNav } = useKunFetch<ChatMessageAsideItem[]>(
  '/message/nav/contact'
)

const system = computed(() => systemNav.value as ChatMessageAsideItem[] | null)
const contact = computed(
  () => (contactNav.value as ChatMessageAsideItem[] | null) ?? []
)
</script>

<template>
  <aside
    :class="
      cn(
        'scrollbar-hide border-default-200/60 flex w-full shrink-0 flex-col space-y-3 overflow-y-auto pr-0 sm:w-88 sm:border-r sm:pr-3',
        routeName !== 'message' ? 'hidden sm:flex' : ''
      )
    "
  >
    <h2 class="px-2 text-2xl">消息</h2>

    <KunDivider />

    <MessageAsideSystemItem v-if="system" title="通知" :data="system[0]!" />

    <MessageAsideMutedItem />

    <MessageAsideSystemItem v-if="system" title="系统消息" :data="system[1]!">
      <template #system>
        <span v-if="!system[1]!.unread_count" class="zako">杂鱼~♡</span>
        <span v-if="system[1]!.unread_count" class="new">
          {{ `「 新消息 」` }}
        </span>
      </template>
    </MessageAsideSystemItem>

    <MessageAsideItem
      v-for="room in contact"
      :key="room.chatroom_name"
      :room="room"
    />

    <div class="block p-2 sm:hidden">
      <h2 class="text-lg">提示</h2>
      <div>本消息系统尚在开发中, 但是功能应该足够用</div>
      <div>如果您有任何问题, 请查看这个话题</div>
      <KunLink
        to="https://www.kungal.com/topic/1650"
        target="_blank"
        class="text-primary underline"
      >
        [公告] 有关论坛消息系统的说明
      </KunLink>
    </div>
  </aside>
</template>
