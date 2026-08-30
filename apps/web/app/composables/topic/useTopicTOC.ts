import { ref, computed } from 'vue'
import type { InjectionKey } from 'vue'

export interface TOCItem {
  id: string
  text: string
  level: number
  type: 'heading' | 'reply'
  targeted?: boolean
}

export interface TopicTocSource {
  getContentHtml: () => string
  getReplies: () => { floor: number; content_markdown: string }[]
  getTargetFloor?: () => number
}

export const TOPIC_TOC_SOURCE: InjectionKey<TopicTocSource> =
  Symbol('topicTocSource')

const TOP_BAR_OFFSET = 88

const htmlToText = (html: string) =>
  html
    .replace(/<[^>]+>/g, '')
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .trim()

const HEADING_RE = /<h([1-3])\b[^>]*\bid="([^"]*)"[^>]*>([\s\S]*?)<\/h\1>/gi

export const useTopicTOC = (source: TopicTocSource) => {
  const headings = computed<TOCItem[]>(() => {
    const items: TOCItem[] = []

    for (const m of source.getContentHtml().matchAll(HEADING_RE)) {
      items.push({
        id: m[2]!,
        level: Number(m[1]),
        text: htmlToText(m[3]!),
        type: 'heading'
      })
    }

    for (const reply of source.getReplies()) {
      const slug = truncateRunes(markdownToText(reply.content_markdown), 20)
      items.push({
        id: `${reply.floor}.${slug}`,
        level: 2,
        text: slug ? `${reply.floor}. ${slug}` : `${reply.floor}`,
        type: 'reply',
        targeted:
          !!source.getTargetFloor && source.getTargetFloor() === reply.floor
      })
    }

    return items
  })

  const activeIds = ref<string[]>([])
  let headingEls: HTMLElement[] = []
  let replyEls: HTMLElement[] = []
  let masterEl: HTMLElement | null = null
  let ticking = false

  const computeActive = () => {
    ticking = false
    const top = TOP_BAR_OFFSET
    const bottom = window.innerHeight
    const ids: string[] = []

    const scan = (els: HTMLElement[], groupEnd: () => number) => {
      for (let i = 0; i < els.length; i++) {
        const el = els[i]!
        const next = els[i + 1]
        const bandTop = el.getBoundingClientRect().top
        const bandBottom = next ? next.getBoundingClientRect().top : groupEnd()
        if (bandTop < bottom && bandBottom > top) {
          ids.push(el.id)
        }
      }
    }

    scan(headingEls, () =>
      masterEl ? masterEl.getBoundingClientRect().bottom : bottom
    )
    scan(replyEls, () =>
      replyEls.length
        ? replyEls[replyEls.length - 1]!.getBoundingClientRect().bottom
        : bottom
    )

    const changed =
      ids.length !== activeIds.value.length ||
      ids.some((id, i) => id !== activeIds.value[i])
    if (changed) {
      activeIds.value = ids
    }
  }

  const onScroll = () => {
    if (ticking) {
      return
    }
    ticking = true
    requestAnimationFrame(computeActive)
  }

  const refreshTOC = () => {
    headingEls = Array.from(
      document.querySelectorAll<HTMLElement>(
        '.kun-master h1, .kun-master h2, .kun-master h3'
      )
    )
    replyEls = Array.from(document.querySelectorAll<HTMLElement>('.kun-reply'))
    masterEl = document.querySelector<HTMLElement>('.kun-master')
    computeActive()
  }

  onMounted(() => {
    refreshTOC()
    window.addEventListener('scroll', onScroll, { passive: true })
    window.addEventListener('resize', onScroll, { passive: true })
  })

  onBeforeUnmount(() => {
    window.removeEventListener('scroll', onScroll)
    window.removeEventListener('resize', onScroll)
  })

  watch(headings, () => nextTick(refreshTOC), { flush: 'post' })

  return {
    headings,
    activeIds,
    refreshTOC
  }
}
