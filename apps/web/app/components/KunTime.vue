<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    time: number | string | Date | null | undefined
    type?: 'relative' | 'date' | 'datetime' | 'auto'
    showYear?: boolean
  }>(),
  { type: 'relative', showYear: false }
)

const isWithinDay = (value: number | string | Date, now: number): boolean => {
  const date = new Date(value)
  return !Number.isNaN(date.getTime()) && now - date.getTime() < 86_400_000
}

const now = ref(Date.now())

const text = computed(() => {
  // formatTimeDifference reads Date.now() itself, so `now` has to be read on
  // every branch: read it only inside the 'auto' test and the 60s timer stops
  // re-running this, which is how relative times froze once already.
  const at = now.value
  const value = props.time ?? ''
  if (
    props.type === 'relative' ||
    (props.type === 'auto' && value !== '' && isWithinDay(value, at))
  ) {
    return formatTimeDifference(value)
  }
  return formatDate(value, {
    isShowYear: props.showYear,
    isPrecise: props.type === 'datetime' || props.type === 'auto'
  })
})

let timer: ReturnType<typeof setInterval> | undefined
onMounted(() => {
  if (props.type === 'relative' || props.type === 'auto') {
    timer = setInterval(() => {
      now.value = Date.now()
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
