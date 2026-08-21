<script setup lang="ts">
import { useKunFloating } from '@kungal/ui-vue'
import { onClickOutside, useEventListener } from '@vueuse/core'

const { isOpen, anchor, selectedText, rememberTrigger, submit, cancel } =
  useKunEditorLinkPrompt()

const panel = ref<HTMLElement | null>(null)
const url = ref('')

useEventListener(document, ['pointerdown', 'click'], rememberTrigger, {
  capture: true
})

const { floatingStyles } = useKunFloating(anchor, panel, {
  placement: 'bottom-start',
  open: isOpen,
  offset: 8,
  constrain: true
})

// `useKunFloating` parks an unanchored panel at the viewport's top-left corner,
// which would be worse than useless; centre it instead.
const panelStyles = computed(() =>
  anchor.value
    ? floatingStyles.value
    : {
        position: 'fixed' as const,
        top: '30%',
        left: '50%',
        transform: 'translateX(-50%)'
      }
)

// `/topic/1` and `#anchor` are real links here, so only a scheme-less bare
// domain gets an https:// put in front of it.
const normalize = (raw: string) => {
  const value = raw.trim()
  if (!value || /^([a-z][a-z0-9+.-]*:|\/|#)/i.test(value)) {
    return value
  }
  return `https://${value}`
}

const href = computed(() => normalize(url.value))

watch(isOpen, (open) => {
  if (!open) {
    return
  }
  const text = selectedText.value.trim()
  url.value = /^([a-z][a-z0-9+.-]*:|www\.)/i.test(text) ? text : ''
})

const apply = () => {
  if (!href.value) {
    return
  }
  submit(href.value)
  url.value = ''
}

onClickOutside(panel, () => cancel(), { ignore: [anchor] })
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0 scale-95"
      leave-active-class="transition duration-100 ease-in"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="isOpen"
        ref="panel"
        data-kun-overlay
        role="dialog"
        aria-label="插入链接"
        :style="panelStyles"
        class="bg-content1 z-kun-popover rounded-kun-lg shadow-kun-md w-[min(22rem,calc(100vw-2rem))] p-2"
        @keydown.esc.stop="cancel"
      >
        <form class="flex items-center gap-1" @submit.prevent="apply">
          <!-- Not type="url": it fails constraint validation on anything without
               a scheme, so `www.kungal.com/topic/1` silently refused to submit on
               Enter (the browser's own "Please enter a URL.", in English) and
               `normalize` never got to add the https:// itself. -->
          <KunInput
            v-model="url"
            inputmode="url"
            size="sm"
            placeholder="https://www.kungal.com"
            :autofocus="true"
            class-name="flex-1"
          />
          <KunButton
            type="submit"
            size="sm"
            variant="flat"
            color="primary"
            :disabled="!href"
          >
            插入
          </KunButton>
        </form>
      </div>
    </Transition>
  </Teleport>
</template>
