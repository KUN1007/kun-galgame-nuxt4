interface SpoilerAnimation {
  overlayEl: HTMLDivElement
  resizeObserver: ResizeObserver
}

export const useSpoilerContent = (containerRef: Ref<HTMLElement | null>) => {
  const spoilerAnimations = new Map<HTMLElement, SpoilerAnimation>()

  const createFrostOverlay = (spoilerEl: HTMLElement) => {
    const overlay = document.createElement('div')
    overlay.className = 'spoiler-frost'
    spoilerEl.insertBefore(overlay, spoilerEl.firstChild)
    const resizeObserver = new ResizeObserver(() => {})
    resizeObserver.observe(spoilerEl)
    return { overlayEl: overlay, resizeObserver }
  }

  const setupSpoilerAnimation = (spoilerEl: HTMLElement) => {
    if (spoilerAnimations.has(spoilerEl)) return

    Array.from(spoilerEl.childNodes).forEach((node) => {
      if (node.nodeType === Node.TEXT_NODE && node.textContent?.trim()) {
        const span = document.createElement('span')
        span.textContent = node.textContent
        spoilerEl.replaceChild(span, node)
      }
    })

    const ani = createFrostOverlay(spoilerEl)

    const computed = getComputedStyle(spoilerEl)
    if (computed.position === 'static' || !computed.position) {
      spoilerEl.classList.add('spoiler-relative-fix')
    }

    spoilerAnimations.set(spoilerEl, ani)
  }

  const destroySpoilerAnimation = (spoilerEl: HTMLElement) => {
    const animation = spoilerAnimations.get(spoilerEl)
    if (!animation) return

    animation.resizeObserver.disconnect()
    animation.overlayEl.classList.add('fade-out')

    animation.overlayEl.addEventListener(
      'transitionend',
      () => {
        animation.overlayEl.remove()
        spoilerAnimations.delete(spoilerEl)
      },
      { once: true }
    )
  }

  const handleSpoilerClick = (spoilerElement: HTMLElement) => {
    spoilerElement.classList.remove('kun-spoiler-hidden')
    spoilerElement.classList.add('!text-foreground')
    destroySpoilerAnimation(spoilerElement)
  }

  const handleCopyClick = (copyButton: HTMLElement) => {
    const container = copyButton.closest('.kun-code-container')
    const pre = container?.querySelector('pre')
    if (!pre) return

    navigator.clipboard
      .writeText(pre.innerText)
      .then(() => {
        copyButton.classList.add('copied')
        setTimeout(() => copyButton.classList.remove('copied'), 3000)
      })
      .catch((err) => {
        console.error('复制失败:', err)
      })
  }

  const handleContainerClick = (event: MouseEvent) => {
    const target = event.target as HTMLElement
    if (!target) return

    const spoilerElement = target.closest('.kun-spoiler.kun-spoiler-hidden')
    if (spoilerElement) {
      event.preventDefault()
      event.stopPropagation()
      handleSpoilerClick(spoilerElement as HTMLElement)
      return
    }

    const copyButton = target.closest('.copy')
    if (copyButton) {
      event.preventDefault()
      event.stopPropagation()
      handleCopyClick(copyButton as HTMLElement)
    }
  }

  const setup = () => {
    const container = containerRef.value
    if (!container) return

    container
      .querySelectorAll<HTMLElement>('.kun-spoiler.kun-spoiler-hidden')
      .forEach(setupSpoilerAnimation)
    container.addEventListener('click', handleContainerClick)
  }

  const cleanup = () => {
    const container = containerRef.value
    if (!container) return

    spoilerAnimations.forEach((_, spoilerEl) =>
      destroySpoilerAnimation(spoilerEl)
    )
    spoilerAnimations.clear()
    container.removeEventListener('click', handleContainerClick)
  }

  watch(
    containerRef,
    (newEl, oldEl) => {
      if (oldEl) {
        cleanup()
      }
      if (newEl) {
        nextTick(setup)
      }
    },
    { flush: 'post' }
  )
}
