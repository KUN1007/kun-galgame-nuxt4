<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    time: number | string | Date | null | undefined
    type?: 'relative' | 'date' | 'datetime' | 'auto'
    showYear?: boolean
  }>(),
  { type: 'relative', showYear: false }
)

const isWithinDay = (value: number | string | Date): boolean => {
  const date = new Date(value)
  return (
    !Number.isNaN(date.getTime()) && Date.now() - date.getTime() < 86_400_000
  )
}

const render = (): string => {
  const value = props.time ?? ''
  if (
    props.type === 'relative' ||
    (props.type === 'auto' && value !== '' && isWithinDay(value))
  ) {
    return formatTimeDifference(value)
  }
  return formatDate(value, {
    isShowYear: props.showYear,
    isPrecise: props.type === 'datetime' || props.type === 'auto'
  })
}

const text = ref(render())

watch(
  () => [props.time, props.type, props.showYear],
  () => {
    text.value = render()
  }
)

let timer: ReturnType<typeof setInterval> | undefined
onMounted(() => {
  text.value = render()
  if (props.type === 'relative' || props.type === 'auto') {
    timer = setInterval(() => {
      text.value = render()
    }, 60_000)
  }
})
onBeforeUnmount(() => {
  if (timer) {
    clearInterval(timer)
  }
})

const machineDateTime = computed(() => {
  if (props.time === null || props.time === undefined || props.time === '') {
    return undefined
  }
  const date = new Date(props.time)
  return Number.isNaN(date.getTime()) ? undefined : date.toISOString()
})
</script>

<template>
  <time
    class="text-default-500"
    :datetime="machineDateTime"
    data-allow-mismatch
    >{{ text }}</time
  >
</template>
