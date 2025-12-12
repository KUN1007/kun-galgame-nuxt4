<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    name?: string
    class?: string
    className?: string
  }>(),
  {
    name: '',
    class: '',
    className: ''
  }
)
</script>

<template>
  <Icon
    v-if="props.name"
    :name="props.name"
    :class="cn('shrink-0 text-inherit', props.class, props.className)"
  />
</template>

<style lang="scss" scoped>
:deep(path) {
  color: inherit;
}

:deep(ellipse) {
  color: inherit;
}

:deep(g) {
  color: inherit;
}

:deep(rect) {
  color: inherit;
}

:deep(circle) {
  color: inherit;
}
</style>
