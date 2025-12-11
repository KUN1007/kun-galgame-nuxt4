<script setup lang="ts">
import { useRipple } from '../ripple/useRipple'
import type { KunUIVariant, KunUIColor } from '../ui/type'
import type { KunButtonProps } from './type'

const props = withDefaults(defineProps<KunButtonProps>(), {
  variant: 'solid',
  color: 'primary',
  size: 'md',
  rounded: 'lg',
  type: 'button',
  disabled: false,
  loading: false,
  fullWidth: false,
  isIconOnly: false,
  icon: false,
  iconPosition: 'left',
  className: '',
  href: '',
  target: '_self',
  ariaLabel: ''
})

const emits = defineEmits<{
  click: [event: MouseEvent]
}>()

const slots = useSlots()

const computedAriaLabel = computed(() => {
  if (props.ariaLabel) {
    return props.ariaLabel
  }

  if (props.isIconOnly) {
    // if (import.meta.dev) {
    //   console.warn(
    //     `[KunButton] An icon-only button should have an explicit 'ariaLabel' prop for accessibility.`
    //   )
    // }
    return 'button'
  }

  if (slots.default) {
    const slotText = extractTextFromVNodes(slots.default()).trim()
    return slotText || ''
  }

  return ''
})

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'xs':
      return 'text-xs px-2 py-1'
    case 'sm':
      return 'text-sm px-3 py-1.5'
    case 'md':
      return 'text-sm px-4 py-2'
    case 'lg':
      return 'text-base px-5 py-2.5'
    case 'xl':
      return 'text-lg px-6 py-3'
    default:
      return 'text-sm px-4 py-2'
  }
})

const variantClasses = computed(() => {
  switch (props.variant) {
    case 'solid':
      return ' text-white'
    case 'bordered':
      return 'border-[1.5px] bg-transparent'
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
    warning: 'bg-warning',
    danger: 'bg-danger'
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
    success: 'bg-transparent text-success hover:bg-success-600/20',
    warning: 'bg-transparent text-warning hover:bg-warning/20',
    danger: 'bg-transparent text-danger hover:bg-danger/20'
  },
  flat: {
    default: 'bg-default/20 text-default-700',
    primary: 'bg-primary/20 text-primary-600',
    secondary: 'bg-secondary/20 text-secondary-600',
    success: 'bg-success-600/20 text-success-700 dark:text-success',
    warning: 'bg-warning/20 text-warning-700 dark:text-warning',
    danger: 'bg-danger/20 text-danger-600 dark:text-danger-500'
  },
  faded: {
    default: 'border-default bg-default-100',
    primary: 'border-default bg-primary-100 text-primary',
    secondary: 'border-default bg-secondary-100 text-secondary',
    success: 'border-default bg-success-600-100 text-success',
    warning: 'border-default bg-warning-100 text-warning',
    danger: 'border-default bg-danger-100 text-danger'
  },
  shadow: {
    default: ' shadow-default/40 bg-default',
    primary: ' shadow-primary/40 bg-primary',
    secondary: ' shadow-secondary/40 bg-secondary',
    success: ' shadow-success/40 bg-success-600',
    warning: ' shadow-warning/40 bg-warning',
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

const roundedClasses = computed(() => {
  switch (props.rounded) {
    case 'none':
      return 'rounded-none'
    case 'sm':
      return 'rounded-sm'
    case 'md':
      return 'rounded-md'
    case 'lg':
      return 'rounded-lg'
    case 'full':
      return 'rounded-full'
    default:
      return 'rounded-lg'
  }
})

const isIconOnlyClasses = computed(() => {
  if (!props.isIconOnly) {
    return ''
  }
  switch (props.size) {
    case 'xs':
      return 'p-1'
    case 'sm':
      return 'p-1.5'
    case 'md':
      return 'p-2'
    case 'lg':
      return 'p-2.5'
    case 'xl':
      return 'p-3'
    default:
      return 'p-2'
  }
})

const { ripples, onClick } = useRipple()

const handleKunButtonClick = (event: MouseEvent) => {
  onClick(event)
  emits('click', event)
}
</script>

<template>
  <component
    :is="props.href ? defineNuxtLink({}) : 'button'"
    :class="
      cn(
        'relative inline-flex cursor-pointer items-center justify-center gap-1 overflow-hidden rounded-md font-medium transition-all hover:opacity-80 active:scale-[0.97] disabled:opacity-50',
        sizeClasses,
        variantClasses,
        colorClasses,
        roundedClasses,
        fullWidth ? 'w-full' : '',
        isIconOnlyClasses,
        disabled && 'cursor-not-allowed hover:bg-none',
        className
      )
    "
    :to="props.href"
    :target="props.target"
    :disabled="disabled || loading"
    :role="props.href ? 'link' : 'button'"
    :type="type"
    :aria-label="computedAriaLabel"
    @click="handleKunButtonClick"
  >
    <KunIcon
      class="text-sm"
      v-if="loading"
      name="svg-spinners:90-ring-with-bg"
    />
    <span v-if="icon && iconPosition === 'left'" class="mr-2">
      <slot name="icon" />
    </span>
    <slot />
    <span v-if="icon && iconPosition === 'right'" class="ml-2">
      <slot name="icon" />
    </span>

    <KunRipple :ripples="ripples" />
  </component>
</template>
