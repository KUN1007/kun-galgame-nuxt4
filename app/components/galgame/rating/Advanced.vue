<script setup lang="ts">
import {
  KUN_GALGAME_DIMENSIONS,
  KUN_GALGAME_DIM_LABELS,
  KUN_GALGAME_DIM_DESCRIPTIONS,
  type KunGalgameDim
} from '~/constants/galgame-rating'

const props = withDefaults(
  defineProps<{
    modelValue: Record<KunGalgameDim, number>
    readonly?: boolean
  }>(),
  { readonly: false }
)

const emit = defineEmits<{
  'update:modelValue': [value: Record<KunGalgameDim, number>]
}>()

const dims = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})
</script>

<template>
  <div class="flex flex-col gap-4 md:flex-row">
    <div class="flex-shrink-0">
      <GalgameRatingRadar v-model="dims" :size="300" :readonly="readonly" />
    </div>

    <div class="grid w-full grid-cols-1 gap-4 md:grid-cols-2">
      <div v-for="dim in KUN_GALGAME_DIMENSIONS" :key="dim" class="space-y-1">
        <div class="flex justify-between">
          <span class="font-medium">{{ KUN_GALGAME_DIM_LABELS[dim] }}</span>
          <span class="text-sm">{{ dims[dim] }}</span>
        </div>
        <div class="pl-3">
          <KunSlider :min="1" :max="10" v-model="dims[dim]" />
        </div>
        <p class="text-xs text-gray-500">
          {{ KUN_GALGAME_DIM_DESCRIPTIONS[dim][dims[dim]] }}
        </p>
      </div>
    </div>
  </div>
</template>
