<script setup lang="ts">
import { useTopicEditorStore } from '~/composables/topic/useTopicEditorStore'
import {
  KUN_TOPIC_ACCESS_ROLE_LIMIT,
  KUN_TOPIC_ACCESS_ROLE_OPTIONS,
  KUN_TOPIC_ACCESS_SCOPE_OPTIONS,
  KUN_TOPIC_ACCESS_USER_LIMIT,
  topicAccessScopeMeta
} from '~/constants/topic'

const { accessScope, accessRoles, accessUserIds } = useTopicEditorStore()

const meta = computed(() => topicAccessScopeMeta(accessScope.value))
</script>

<template>
  <section class="space-y-3">
    <h3 class="flex items-center gap-2 font-semibold">
      <KunIcon :name="meta.icon" class="h-5 w-5" />
      访问范围
    </h3>

    <KunSelect
      v-model="accessScope"
      :options="KUN_TOPIC_ACCESS_SCOPE_OPTIONS"
      aria-label="话题访问范围"
    />

    <p class="text-default-500 text-sm">{{ meta.hint }}</p>

    <div v-if="accessScope === 'role'" class="space-y-2">
      <span class="text-sm font-medium">指定可以看到本话题的角色</span>
      <KunCheckBoxGroup
        v-model="accessRoles"
        :options="KUN_TOPIC_ACCESS_ROLE_OPTIONS"
        :max="KUN_TOPIC_ACCESS_ROLE_LIMIT"
        variant="pill"
        color="primary"
        size="sm"
        orientation="horizontal"
      />
      <p v-if="!accessRoles.length" class="text-danger-500 text-sm">
        还没有选择任何角色, 现在只有您自己与管理人员能打开这个话题
      </p>
    </div>

    <EditTopicAccessUserPicker
      v-if="accessScope === 'users'"
      v-model="accessUserIds"
      :limit="KUN_TOPIC_ACCESS_USER_LIMIT"
    />
  </section>
</template>
