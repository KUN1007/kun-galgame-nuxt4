<script setup lang="ts">
import {
  createWebsiteCategorySchema,
  updateWebsiteCategorySchema
} from '~/validations/website'
import type {
  CreateWebsiteCategoryPayload,
  UpdateWebsiteCategoryPayload
} from './types'

type CategoryData = CreateWebsiteCategoryPayload & { category_id?: number }

const props = defineProps<{
  modelValue: boolean
  initialData?: CategoryData
  loading?: boolean
}>()

const emits = defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [data: CreateWebsiteCategoryPayload | UpdateWebsiteCategoryPayload]
}>()

const isModalOpen = computed({
  get: () => props.modelValue,
  set: (value) => emits('update:modelValue', value)
})

const isEditing = computed(() => !!props.initialData?.category_id)

const getInitialFormData = (): CategoryData => ({
  name: '',
  label: '',
  description: '',
  sort_order: 0,
  ...(props.initialData || {})
})

const formData = reactive<CategoryData>(getInitialFormData())

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
    ? updateWebsiteCategorySchema
    : createWebsiteCategorySchema
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
    :aria-label="isEditing ? '编辑分类' : '创建分类'"
    inner-class-name="max-w-md"
  >
    <form @submit.prevent="handleSubmit">
      <h2 class="mb-6 text-xl font-bold">
        {{ isEditing ? '编辑分类' : '创建分类' }}
      </h2>

      <div class="space-y-4">
        <KunInput
          v-model="formData.name"
          label="分类标识 (URL 用, 小写英文)"
          placeholder="resource"
          required
        />
        <KunInput
          v-model="formData.label"
          label="分类显示名"
          placeholder="Galgame 资源网站"
          required
        />
        <KunInput
          v-model="formData.sort_order"
          label="排序 (数字越小越靠前)"
          type="number"
        />
        <KunTextarea
          v-model="formData.description"
          label="分类描述 (可选)"
          auto-grow
          show-char-count
          :maxlength="300"
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
