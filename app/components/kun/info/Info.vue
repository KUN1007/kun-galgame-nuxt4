<script setup lang="ts">
import type { KunUIVariant, KunUIColor } from '../ui/type'

type KunInfoColor = KunUIColor | 'info'

interface Props {
  title: string
  description?: string
  className?: string
  color?: KunInfoColor
  variant?: KunUIVariant
  icon?: string
}

const props = withDefaults(defineProps<Props>(), {
  description: '',
  color: 'default',
  className: '',
  variant: 'flat',
  icon: ''
})

const variantClasses = computed(() => {
  switch (props.variant) {
    case 'solid':
      return ' text-white'
    case 'bordered':
      return 'border-[1.5px]'
    case 'light':
      return 'bg-opacity-20 border-transparent'
    case 'flat':
      return 'bg-opacity-20 border-transparent shadow-none'
    case 'faded':
      return 'bg-opacity-10 border-transparent'
    case 'shadow':
      return ' text-white'
    case 'ghost':
      return 'bg-transparent border-transparent shadow-none hover:bg-opacity-10'
    default:
      return ''
  }
})

const colorVariants: Record<KunUIVariant, Record<KunInfoColor, string>> = {
  solid: {
    default: 'bg-default',
    primary: 'bg-primary',
    secondary: 'bg-secondary',
    success: 'bg-success-600',
    warning: 'bg-warning',
    danger: 'bg-danger-600',
    info: 'bg-info-600'
  },
  bordered: {
    default: 'bg-transparent bg-default/15 border-default',
    primary: 'bg-transparent bg-primary/15 border-primary text-primary',
    secondary: 'bg-transparent bg-secondary/15 border-secondary text-secondary',
    success: 'bg-transparent bg-success/15 border-success-600 text-success',
    warning: 'bg-transparent bg-warning/15 border-warning-600 text-warning',
    danger: 'bg-transparent bg-danger/15 border-danger text-danger',
    info: 'bg-transparent bg-info/15 border-info text-info'
  },
  light: {
    default: 'bg-transparent hover:bg-default/40',
    primary: 'bg-transparent text-primary-800 hover:bg-primary/20',
    secondary: 'bg-transparent text-secondary-800 hover:bg-secondary/20',
    success: 'bg-transparent text-success-800 hover:bg-success/20',
    warning: 'bg-transparent text-warning-800 hover:bg-warning/20',
    danger: 'bg-transparent text-danger-800 hover:bg-danger/20',
    info: 'bg-transparent text-info-800 hover:bg-info/20'
  },
  flat: {
    default: 'bg-default/15 text-default-800',
    primary: 'bg-primary/15 text-primary-800',
    secondary: 'bg-secondary/15 text-secondary-800',
    success: 'bg-success/15 text-success-800 dark:text-success',
    warning: 'bg-warning/15 text-warning-800 dark:text-warning',
    danger: 'bg-danger/15 text-danger-800 dark:text-danger-500',
    info: 'bg-info/15 text-info-800 dark:text-info-500'
  },
  faded: {
    default: 'border-default bg-default-100',
    primary: 'border-default bg-primary-100 text-primary',
    secondary: 'border-default bg-secondary-100 text-secondary',
    success: 'border-default bg-success-100 text-success',
    warning: 'border-default bg-warning-100 text-warning',
    danger: 'border-default bg-danger-100 text-danger',
    info: 'border-default bg-info-100 text-info'
  },
  shadow: {
    default: ' shadow-default/50 bg-default',
    primary: ' shadow-primary/40 bg-primary',
    secondary: ' shadow-secondary/40 bg-secondary',
    success: ' shadow-success/40 bg-success',
    warning: ' shadow-warning/40 bg-warning dark:text-black',
    danger: ' shadow-danger/40 bg-danger',
    info: ' shadow-info/40 bg-info'
  },
  ghost: {
    default: 'border-default',
    primary: 'border-primary text-primary',
    secondary: 'border-secondary text-secondary',
    success: 'border-success text-success',
    warning: 'border-warning text-warning',
    danger: 'border-danger text-danger',
    info: 'border-info text-info'
  }
} as const

const colorClasses = computed(() => {
  return colorVariants[props.variant]?.[props.color] || ''
})

const titleColor = computed(() => {
  switch (props.color) {
    case 'primary':
      return 'text-primary-900'
    case 'secondary':
      return 'text-secondary-900'
    case 'danger':
      return 'text-danger-900'
    case 'info':
      return 'text-info-900'
    case 'success':
      return 'text-success-900'
    case 'warning':
      return 'text-warning-900'
    default:
      return 'text-default-900'
  }
})
</script>

<template>
  <div
    :class="
      cn('space-y-2 rounded-lg p-3', variantClasses, colorClasses, className)
    "
  >
    <h3 :class="cn('flex items-center gap-2 font-medium', titleColor)">
      <KunIcon
        v-if="icon"
        :name="icon"
        :class-name="cn('h-5 w-5 flex-shrink-0')"
      />
      <span>{{ title }}</span>
    </h3>

    <div v-if="description || $slots.default" class="text-sm opacity-90">
      <p v-if="description">{{ description }}</p>
      <slot />
    </div>
  </div>
</template>
