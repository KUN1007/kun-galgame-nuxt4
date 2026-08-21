import type { KunEditorAdapters } from '@kungal/editor-core'

const isOpen = ref(false)
const selectedText = ref('')
const anchor = ref<HTMLElement | null>(null)

// Only ever assigned from a click inside the editor, so this never runs on the
// server and cannot leak between requests the way module state usually does.
let resolveHref: ((href: string | null) => void) | null = null
let lastTrigger: HTMLElement | null = null

const settle = (href: string | null) => {
  isOpen.value = false
  selectedText.value = ''
  anchor.value = null
  resolveHref?.(href)
  resolveHref = null
}

// The editor calls `linkPrompt` from inside its own handler and passes no event,
// so the capture phase — already over by then — is the only place left to learn
// which button the panel should hang off.
//
// It has to be BOTH events. The selection bubble acts on `pointerdown`, because
// waiting for `click` would let the mousedown blur the editor and lose the very
// selection being linked; a click-only listener therefore recorded the button
// AFTER the prompt had already run, and the panel opened centred on the screen
// instead of under the bubble. `click` still earns its place: activating the
// toolbar button from the keyboard fires no pointer event at all.
const rememberTrigger = (event: Event) => {
  const target = event.target as HTMLElement | null
  lastTrigger = target?.closest?.('button') ?? null
}

export const useKunEditorLinkPrompt = () => {
  const prompt: NonNullable<KunEditorAdapters['linkPrompt']> = ({ text }) => {
    settle(null)
    anchor.value = lastTrigger?.isConnected ? lastTrigger : null
    selectedText.value = text
    isOpen.value = true
    return new Promise<string | null>((resolve) => {
      resolveHref = resolve
    })
  }

  return {
    isOpen,
    anchor,
    selectedText,
    prompt,
    rememberTrigger,
    submit: (href: string) => settle(href),
    cancel: () => settle(null)
  }
}
