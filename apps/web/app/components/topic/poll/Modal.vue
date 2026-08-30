<script setup lang="ts">
import { usePoll } from '~/composables/topic/usePoll'
import { createPollSchema, updatePollSchema } from '~/validations/topic-poll'
import { TOPIC_POLL_VISIBILITY_OPTIONS } from '~/constants/topic'
import { deadlineFromPicker, deadlineToPicker } from '../miniapp/deadline'
import type { PollFormData } from './types'

const props = defineProps<{
  modelValue: boolean
  topicId: number
  initialData?: TopicPoll
}>()

const emits = defineEmits<{
  'update:modelValue': [value: boolean]
  refresh: []
}>()

const isModalOpen = computed({
  get: () => props.modelValue,
  set: (value) => emits('update:modelValue', value)
})

const isEditing = computed(() => !!props.initialData)
const { createPoll, updatePoll } = usePoll(props.topicId)
const isLoading = ref(false)

const getInitialFormData = (): PollFormData => {
  if (isEditing.value && props.initialData) {
    return {
      topic_id: 0,
      poll_id: props.initialData.id,
      title: props.initialData.title,
      description: props.initialData.description,
      options: props.initialData.option.map((o) => ({
        id: o.id,
        text: o.text,
        _status: 'existing'
      })),
      type: props.initialData.type as 'single',
      min_choice: props.initialData.min_choice,
      max_choice: props.initialData.max_choice,
      deadline: deadlineToPicker(props.initialData.deadline),
      result_visibility: props.initialData.result_visibility as 'always',
      is_anonymous: props.initialData.is_anonymous,
      can_change_vote: props.initialData.can_change_vote,
      status: props.initialData.status as 'open'
    }
  }

  return {
    topic_id: props.topicId,
    poll_id: 0,
    title: '',
    description: '',
    options: [{ text: '' }, { text: '' }],
    type: 'single',
    min_choice: 1,
    max_choice: 1,
    deadline: undefined,
    result_visibility: 'always',
    is_anonymous: false,
    can_change_vote: false,
    status: 'open'
  }
}

const formData = reactive<PollFormData>(getInitialFormData())

watch(
  () => props.modelValue,
  (isOpen) => {
    if (isOpen) {
      Object.assign(formData, getInitialFormData())
    }
  }
)

watch(
  () => formData.type,
  (newType) => {
    if (newType === 'single') {
      formData.min_choice = 1
      formData.max_choice = 1
    } else {
      formData.min_choice = 1
    }
  }
)

const addOption = () => {
  if (formData.options.length >= 20) {
    useMessage('最多添加20个选项', 'warn')
    return
  }
  formData.options.push({ text: '', _status: 'new' })
}

const removeOption = (index: number) => {
  if (formData.options.filter((o) => o._status !== 'deleted').length <= 2) {
    useMessage('至少需要2个选项', 'warn')
    return
  }
  const option = formData.options[index]
  if (isEditing.value && option?.id) {
    option._status = 'deleted'
  } else {
    formData.options.splice(index, 1)
  }
}

const handleSubmit = async () => {
  const schema = isEditing.value ? updatePollSchema : createPollSchema
  const payload: PollFormData = {
    ...formData,
    deadline: deadlineFromPicker(formData.deadline)
  }
  const result = schema.safeParse(
    isEditing.value ? { ...payload, options: {} } : payload
  )

  if (!result.success) {
    const message = JSON.parse(result.error.message)[0]
    useMessage(formatKunZodIssue(message), 'warn')
    return
  }

  isLoading.value = true
  if (isEditing.value && props.initialData) {
    await updatePoll(props.initialData.id, props.initialData.option, payload)
  } else {
    await createPoll(payload)
  }

  emits('refresh')
  isLoading.value = false
  isModalOpen.value = false
}
</script>

<template>
  <KunModal
    :is-dismissable="false"
    v-model="isModalOpen"
    inner-class-name="max-w-3xl"
  >
    <form @submit.prevent="handleSubmit">
      <h2 class="mb-3 text-2xl font-bold">
        {{ isEditing ? '编辑投票' : '创建投票' }}
      </h2>

      <p class="text-default-500 mb-6 text-sm">
        目前阶段, 话题下方投票最多 30 个, 每个投票最多 20 个选项
      </p>

      <div class="flex flex-col gap-4">
        <KunInput v-model="formData.title" label="投票标题" required />
        <KunTextarea
          v-model="formData.description"
          label="补充描述 (可选)"
          auto-grow
          :rows="2"
        />

        <div>
          <label class="mb-1 block text-sm font-medium">选项设置</label>
          <div class="flex flex-col gap-2">
            <template v-for="(option, index) in formData.options" :key="index">
              <div
                v-if="option._status !== 'deleted'"
                class="flex items-center gap-2"
              >
                <KunInput
                  v-model="option.text"
                  :placeholder="`选项 ${index + 1}`"
                  class-name="flex-grow"
                />
                <KunButton
                  variant="light"
                  color="danger"
                  :is-icon-only="true"
                  @click="removeOption(index)"
                >
                  <KunIcon name="lucide:trash-2" />
                </KunButton>
              </div>
            </template>
          </div>
          <KunButton
            variant="light"
            size="sm"
            class-name="mt-2"
            @click="addOption"
          >
            <KunIcon name="lucide:plus" class="mr-1" />
            增加选项
          </KunButton>
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="mb-1 block text-sm font-medium">投票类型</label>
            <div class="flex gap-4">
              <KunCheckBox
                :model-value="formData.type === 'single'"
                label="单选"
                color="primary"
                @update:model-value="formData.type = 'single'"
              />
              <KunCheckBox
                :model-value="formData.type === 'multiple'"
                label="多选"
                color="primary"
                @update:model-value="formData.type = 'multiple'"
              />
            </div>
          </div>

          <div
            v-if="formData.type === 'multiple'"
            class="grid grid-cols-2 gap-2"
          >
            <KunInput
              v-model.number="formData.min_choice"
              label="至少选"
              type="number"
              :min="1"
            />
            <KunInput
              v-model.number="formData.max_choice"
              label="至多选"
              type="number"
              :min="formData.min_choice"
            />
          </div>
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <KunDatePicker v-model="formData.deadline" label="截止日期 (可选)" />
          <KunSelect
            v-model="formData.result_visibility"
            label="结果可见性"
            :options="TOPIC_POLL_VISIBILITY_OPTIONS"
          />
        </div>

        <div class="flex flex-col gap-2">
          <KunSwitch v-model="formData.is_anonymous" label="匿名投票" />
          <KunSwitch v-model="formData.can_change_vote" label="允许修改投票" />
        </div>
      </div>

      <div class="mt-8 flex justify-end gap-3">
        <KunButton variant="light" @click="isModalOpen = false">取消</KunButton>
        <KunButton type="submit" color="primary" :loading="isLoading">
          {{ isEditing ? '保存更改' : '发布投票' }}
        </KunButton>
      </div>
    </form>
  </KunModal>
</template>
