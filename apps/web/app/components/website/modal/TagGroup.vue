<script setup lang="ts">
import {
  createWebsiteTagGroupSchema,
  updateWebsiteTagGroupSchema
} from '~/validations/website'
import type {
  CreateWebsiteTagGroupPayload,
  UpdateWebsiteTagGroupPayload
} from './types'

type TagGroupData = CreateWebsiteTagGroupPayload & { group_id?: number }

const props = defineProps<{
  modelValue: boolean
  initialData?: TagGroupData
  loading?: boolean
}>()

const emits = defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [data: CreateWebsiteTagGroupPayload | UpdateWebsiteTagGroupPayload]
}>()

const isModalOpen = computed({
  get: () => props.modelValue,
  set: (value) => emits('update:modelValue', value)
})

const isEditing = computed(() => !!props.initialData?.group_id)

const getInitialFormData = (): TagGroupData => ({
  name: '',
  label: '',
  description: '',
  sort_order: 0,
  multi_select: false,
  ...(props.initialData || {})
})

const formData = reactive<TagGroupData>(getInitialFormData())

watch(
  () => isModalOpen.value,
  (isOpen) => {
    if (isOpen) {
      Object.assign(formData, getInitialFormData())
    }
  }
)

const handleSubmit = () => {
  const schema = isEditing.value
    ? updateWebsiteTagGroupSchema
    : createWebsiteTagGroupSchema
  const result = schema.safeParse(formData)

  if (!result.success) {
    const message = JSON.parse(result.error.message)[0]
    useMessage(formatKunZodIssue(message), 'warn')
    return
  }
  emits('submit', result.data)
}
</script>

<template>
  <KunModal
    :is-dismissable="false"
    v-model="isModalOpen"
    :aria-label="isEditing ? '编辑标签分组' : '创建标签分组'"
    inner-class-name="max-w-md"
  >
    <form @submit.prevent="handleSubmit">
      <h2 class="mb-6 text-xl font-bold">
        {{ isEditing ? '编辑标签分组' : '创建标签分组' }}
      </h2>

      <div class="space-y-4">
        <KunInput
          v-model="formData.name"
          label="分组标识 (小写英文)"
          placeholder="performance"
          required
        />
        <KunInput
          v-model="formData.label"
          label="分组显示名"
          placeholder="网站性能"
          required
        />
        <KunInput
          v-model="formData.sort_order"
          label="排序 (数字越小越靠前)"
          type="number"
        />
        <KunTextarea
          v-model="formData.description"
          label="分组描述 (可选)"
          auto-grow
          show-char-count
          :maxlength="300"
        />
        <KunSwitch
          v-model="formData.multi_select"
          label="该分组可多选 (关闭则组内标签互斥, 只能选一个)"
        />
      </div>

      <div class="mt-6 flex justify-end gap-3">
        <KunButton
          variant="light"
          color="danger"
          :disabled="loading"
          @click="isModalOpen = false"
        >
          取消
        </KunButton>
        <KunButton type="submit" color="primary" :loading="loading">
          {{ isEditing ? '保存更改' : '创建' }}
        </KunButton>
      </div>
    </form>
  </KunModal>
</template>
