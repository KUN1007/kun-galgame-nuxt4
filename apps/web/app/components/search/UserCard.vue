<script setup lang="ts">
defineProps<{
  user: SearchResultUser
  keywords?: string
}>()
</script>

<template>
  <KunCard :href="`/user/${user.id}`" :is-hoverable="true" padding="md">
    <div class="flex items-center gap-2">
      <KunAvatar :disable-floating="true" :user="user" :is-navigation="false" />
      <span class="truncate font-medium">
        <SearchHighlight :text="user.name" :keywords="keywords" />
      </span>
    </div>

    <p
      v-if="user.bio"
      class="text-default-600 mt-2 line-clamp-2 text-sm break-all"
    >
      <SearchHighlight :text="user.bio" :keywords="keywords" />
    </p>

    <div
      v-if="user.moemoepoint || user.created"
      class="mt-2 flex items-center justify-between text-sm"
    >
      <div v-if="user.moemoepoint" class="text-secondary flex items-center">
        <KunIcon name="lucide:lollipop" class="size-5" />
        {{ user.moemoepoint }}
      </div>
      <span v-if="user.created" class="text-default-700">
        <KunTime :time="user.created" type="date" show-year />
      </span>
    </div>
  </KunCard>
</template>
