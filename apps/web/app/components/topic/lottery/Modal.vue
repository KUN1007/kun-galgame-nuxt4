<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useLottery } from '~/composables/topic/useLottery'
import { lotterySchema } from '~/validations/topic-lottery'
import {
  KUN_LOTTERY_DELIVERY_OPTIONS,
  KUN_LOTTERY_DRAW_MODE_OPTIONS,
  KUN_LOTTERY_ENTRY_MODE_OPTIONS
} from '~/constants/topic'
import type { LotteryFormData, LotteryPrizeFormData } from './types'

const props = defineProps<{
  modelValue: boolean
  topicId: number
  initialData?: TopicLottery
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
const { createLottery, updateLottery } = useLottery(props.topicId)
const isLoading = ref(false)

const emptyPrize = (): LotteryPrizeFormData => ({
  name: '',
  description: '',
  image_hash: '',
  delivery: 'manual',
  point_amount: 0,
  slots: 1,
  codes: ''
})

const getInitialFormData = (): LotteryFormData => {
  const initial = props.initialData
  if (initial) {
    return {
      topic_id: props.topicId,
      lottery_id: initial.id,
      title: initial.title,
      description: initial.description,
      entry_mode: initial.entry_mode,
      floor_rule: initial.floor_rule,
      draw_mode: initial.draw_mode,
      draw_threshold: initial.draw_threshold,
      deadline: initial.deadline ? initial.deadline.toString() : undefined,
      min_account_age_days: initial.min_account_age_days,
      min_moemoepoint: initial.min_moemoepoint,
      show_entrants: initial.show_entrants,
      // Codes are never sent back to the browser, so an edit that keeps the
      // prizes must not resubmit them: an empty prize list tells the server to
      // leave the prizes and their escrowed codes exactly as they are.
      prizes: initial.prizes.map((p) => ({
        name: p.name,
        description: p.description,
        image_hash: p.image_hash,
        delivery: p.delivery,
        point_amount: p.point_amount,
        slots: p.slots,
        codes: ''
      }))
    }
  }

  return {
    topic_id: props.topicId,
    lottery_id: 0,
    title: '',
    description: '',
    entry_mode: 'signup',
    floor_rule: '',
    draw_mode: 'deadline',
    draw_threshold: 0,
    deadline: undefined,
    min_account_age_days: 0,
    min_moemoepoint: 0,
    show_entrants: true,
    prizes: [emptyPrize()]
  }
}

const formData = reactive<LotteryFormData>(getInitialFormData())
const rewritePrizes = ref(false)

watch(
  () => props.modelValue,
  (isOpen) => {
    if (isOpen) {
      Object.assign(formData, getInitialFormData())
      rewritePrizes.value = !props.initialData
    }
  }
)

const totalSlots = computed(() =>
  formData.prizes.reduce((sum, p) => sum + Number(p.slots || 0), 0)
)

const isFloor = computed(() => formData.entry_mode === 'floor')

watch(isFloor, (floor) => {
  if (floor && formData.draw_mode === 'threshold') {
    formData.draw_mode = 'deadline'
  }
})

const addPrize = () => {
  if (formData.prizes.length >= 10) {
    useMessage('最多添加 10 个奖项', 'warn')
    return
  }
  formData.prizes.push(emptyPrize())
}

const removePrize = (index: number) => {
  if (formData.prizes.length <= 1) {
    useMessage('至少需要一个奖项', 'warn')
    return
  }
  formData.prizes.splice(index, 1)
}

const codeCount = (prize: LotteryPrizeFormData) =>
  prize.codes.split('\n').filter((c) => c.trim()).length

const handleSubmit = async () => {
  const payload = {
    ...formData,
    prizes: rewritePrizes.value ? formData.prizes : []
  }
  if (rewritePrizes.value) {
    const result = lotterySchema.safeParse(payload)
    if (!result.success) {
      const message = JSON.parse(result.error.message)[0]
      useMessage(formatKunZodIssue(message), 'warn')
      return
    }
    for (const prize of formData.prizes) {
      if (prize.delivery === 'code' && codeCount(prize) !== prize.slots) {
        useMessage(
          `奖项「${prize.name || '未命名'}」有 ${prize.slots} 个名额, 需要正好 ${prize.slots} 个兑换码`,
          'warn'
        )
        return
      }
    }
  }

  isLoading.value = true
  try {
    if (isEditing.value && props.initialData) {
      await updateLottery(props.initialData.id, payload)
    } else {
      await createLottery(payload)
    }
    emits('refresh')
    isModalOpen.value = false
  } finally {
    isLoading.value = false
  }
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
        {{ isEditing ? '编辑抽奖' : '创建抽奖' }}
      </h2>

      <p class="text-default-500 mb-6 text-sm">
        每个话题最多 10 个抽奖, 每个抽奖最多 10 个奖项。本站只负责产生中奖名单,
        不担保实物履约, 也不代收收货地址。
      </p>

      <div class="flex flex-col gap-4">
        <KunInput v-model="formData.title" label="抽奖标题" required />
        <KunTextarea
          v-model="formData.description"
          label="抽奖说明 (可选)"
          auto-grow
          :rows="2"
        />

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <KunSelect
            v-model="formData.entry_mode"
            label="参与方式"
            :options="KUN_LOTTERY_ENTRY_MODE_OPTIONS"
          />
          <KunSelect
            v-model="formData.draw_mode"
            label="开奖方式"
            :options="KUN_LOTTERY_DRAW_MODE_OPTIONS"
          />
        </div>

        <KunInput
          v-if="isFloor"
          v-model="formData.floor_rule"
          label="中奖楼层"
          placeholder="8,18,28 或 every:10 (每 10 楼)"
          :description="`楼层数量需要与总名额一致, 当前总名额 ${totalSlots}`"
        />

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <KunDatePicker
            v-if="formData.draw_mode === 'deadline'"
            v-model="formData.deadline"
            label="开奖时间"
          />
          <KunInput
            v-if="formData.draw_mode === 'threshold'"
            v-model.number="formData.draw_threshold"
            label="满多少人开奖"
            type="number"
            :min="totalSlots"
          />
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <KunInput
            v-model.number="formData.min_account_age_days"
            label="参与门槛: 注册满 (天)"
            type="number"
            :min="0"
          />
          <KunInput
            v-model.number="formData.min_moemoepoint"
            label="参与门槛: 萌萌点不低于"
            type="number"
            :min="0"
          />
        </div>

        <KunSwitch v-model="formData.show_entrants" label="公开参与名单" />

        <div>
          <div class="mb-2 flex items-center justify-between">
            <label class="block text-sm font-medium">
              奖项设置 (共 {{ totalSlots }} 个名额)
            </label>
            <KunSwitch
              v-if="isEditing"
              v-model="rewritePrizes"
              label="修改奖项"
            />
          </div>

          <KunInfo v-if="isEditing && !rewritePrizes" title="奖项保持不变">
            <p class="text-sm">
              打开「修改奖项」会整体重写奖项列表, 已托管的兑换码会被清空,
              需要重新填写。已经有人参与后奖项不可再改。
            </p>
          </KunInfo>

          <div v-else class="flex flex-col gap-3">
            <div
              v-for="(prize, index) in formData.prizes"
              :key="index"
              class="border-default-200 space-y-3 rounded-md border p-3"
            >
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium">奖项 {{ index + 1 }}</span>
                <KunButton
                  variant="light"
                  color="danger"
                  size="sm"
                  :is-icon-only="true"
                  @click="removePrize(index)"
                >
                  <KunIcon name="lucide:trash-2" />
                </KunButton>
              </div>

              <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
                <KunInput v-model="prize.name" label="奖品名称" />
                <KunInput
                  v-model.number="prize.slots"
                  label="名额"
                  type="number"
                  :min="1"
                />
              </div>

              <KunTextarea
                v-model="prize.description"
                label="奖品描述 (可选)"
                auto-grow
                :rows="2"
              />

              <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
                <KunSelect
                  v-model="prize.delivery"
                  label="发放方式"
                  :options="KUN_LOTTERY_DELIVERY_OPTIONS"
                />
                <KunInput
                  v-if="prize.delivery === 'point'"
                  v-model.number="prize.point_amount"
                  label="萌萌点数量"
                  type="number"
                  :min="1"
                />
              </div>

              <div v-if="prize.delivery === 'code'">
                <KunTextarea
                  v-model="prize.codes"
                  label="兑换码 (每行一个)"
                  auto-grow
                  :rows="4"
                />
                <p class="text-default-500 mt-1 text-xs">
                  已填写 {{ codeCount(prize) }} 个, 需要 {{ prize.slots }} 个。
                  兑换码提交后即加密托管, 页面上永远不会再显示,
                  只有中奖者本人能揭示。
                </p>
              </div>

              <KunInfo v-if="prize.delivery === 'manual'" title="实物类奖品">
                <p class="text-sm">
                  开奖后请通过私聊与中奖者沟通。请不要要求对方在公开楼层留下地址,
                  本站也不会代为收集收货信息。
                </p>
              </KunInfo>
            </div>

            <KunButton variant="light" size="sm" @click="addPrize">
              <KunIcon name="lucide:plus" class="mr-1" />
              增加奖项
            </KunButton>
          </div>
        </div>
      </div>

      <div class="mt-8 flex justify-end gap-3">
        <KunButton variant="light" @click="isModalOpen = false">取消</KunButton>
        <KunButton type="submit" color="primary" :loading="isLoading">
          {{ isEditing ? '保存更改' : '发布抽奖' }}
        </KunButton>
      </div>
    </form>
  </KunModal>
</template>
