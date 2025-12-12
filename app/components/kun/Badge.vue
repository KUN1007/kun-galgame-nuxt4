<script setup lang="ts">
import type { KunUIVariant, KunUIColor, KunUISize } from './ui/type'

interface Props {
  className?: string
  color?: KunUIColor
  size?: KunUISize
  variant?: KunUIVariant
}

const props = withDefaults(defineProps<Props>(), {
  color: 'default',
  className: '',
  size: 'sm',
  variant: 'flat'
})

const variantClasses = computed(() => {
  switch (props.variant) {
    case 'solid':
      return ' text-white'
    case 'bordered':
      return 'border-2 bg-transparent'
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

const colorVariants: Record<KunUIVariant, Record<KunUIColor, string>> = {
  solid: {
    default: 'bg-default',
    primary: 'bg-primary',
    secondary: 'bg-secondary',
    success: 'bg-success-600 dark:bg-success-300',
    warning: 'bg-warning dark:bg-warning-300',
    danger: 'bg-danger-600 dark:bg-danger-300'
  },
  bordered: {
    default: 'bg-transparent border-default',
    primary: 'bg-transparent border-primary text-primary',
    secondary: 'bg-transparent border-secondary text-secondary',
    success: 'bg-transparent border-success text-success',
    warning: 'bg-transparent border-warning text-warning',
    danger: 'bg-transparent border-danger text-danger'
  },
  light: {
    default: 'bg-transparent hover:bg-default/20',
    primary: 'bg-transparent text-primary hover:bg-primary/20',
    secondary: 'bg-transparent text-secondary hover:bg-secondary/20',
    success: 'bg-transparent text-success hover:bg-success/20',
    warning: 'bg-transparent text-warning hover:bg-warning/20',
    danger: 'bg-transparent text-danger hover:bg-danger/20'
  },
  flat: {
    default: 'bg-default/20 text-default-700',
    primary: 'bg-primary/20 text-primary-600',
    secondary: 'bg-secondary/20 text-secondary-600',
    success: 'bg-success/20 text-success-700 dark:text-success',
    warning: 'bg-warning/20 text-warning-700 dark:text-warning',
    danger: 'bg-danger/20 text-danger-600 dark:text-danger-500'
  },
  faded: {
    default: 'border-default bg-default-100',
    primary: 'border-default bg-primary-100 text-primary',
    secondary: 'border-default bg-secondary-100 text-secondary',
    success: 'border-default bg-success-100 text-success',
    warning: 'border-default bg-warning-100 text-warning',
    danger: 'border-default bg-danger-100 text-danger'
  },
  shadow: {
    default: ' shadow-default/40 bg-default',
    primary: ' shadow-primary/40 bg-primary',
    secondary: ' shadow-secondary/40 bg-secondary',
    success: ' shadow-success/40 bg-success',
    warning: ' shadow-warning/40 bg-warning dark:text-black',
    danger: ' shadow-danger/40 bg-danger'
  },
  ghost: {
    default: 'border-default',
    primary: 'border-primary text-primary',
    secondary: 'border-secondary text-secondary',
    success: 'border-success text-success',
    warning: 'border-warning text-warning',
    danger: 'border-danger text-danger'
  }
} as const

const colorClasses = computed(() => {
  return colorVariants[props.variant]?.[props.color] || ''
})

const sizeClasses: Record<KunUISize, string> = {
  xs: 'px-2 py-0.5 text-xs',
  sm: 'px-2 py-1 text-xs',
  md: 'px-3 py-1.5 text-sm',
  lg: 'px-4 py-2 text-sm',
  xl: 'px-6 py-3 text-base'
}
</script>

<template>
  <span
    :class="
      cn(
        'inline-flex cursor-default items-center justify-center gap-1 rounded-full font-medium',
        sizeClasses[size],
        variantClasses,
        colorClasses,
        className
      )
    "
  >
    <slot />
  </span>
</template>
