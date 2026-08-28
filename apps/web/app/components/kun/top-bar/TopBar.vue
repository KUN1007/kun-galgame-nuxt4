<script setup lang="ts">
import { useWindowScroll } from '@vueuse/core'

withDefaults(
  defineProps<{
    className?: string
  }>(),
  { className: '' }
)

const offsetClass = 'desktop-nav:left-[104px] desktop-nav:w-[calc(100%-116px)]'

const { y } = useWindowScroll()
const scrolled = computed(() => y.value > 8)
</script>

<template>
  <div
    :class="
      cn(
        'fixed top-0 z-30 mb-3 ml-0 shrink-0',
        'left-0 w-full',
        offsetClass,
        className
      )
    "
  >
    <div
      :class="
        cn(
          'mx-auto flex h-16 w-full transform-gpu items-center justify-between px-3 transition-[background-color,border-color,box-shadow,padding] duration-200',
          'bg-content1 border-kun shadow-kun-sm border-b backdrop-blur-md',
          'desktop-nav:max-w-7xl desktop-nav:rounded-b-lg desktop-nav:border',
          !scrolled &&
            'desktop-nav:border-transparent desktop-nav:bg-transparent desktop-nav:px-0 desktop-nav:shadow-none desktop-nav:backdrop-blur-none'
        )
      "
    >
      <KunTopBarNav />
      <KunTopBarAvatar />
    </div>
  </div>
</template>
