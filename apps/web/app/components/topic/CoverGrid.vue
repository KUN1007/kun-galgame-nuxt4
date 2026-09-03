<script setup lang="ts">
import { useContentLightbox } from '@kungal/ui-vue'

const props = defineProps<{
  images: string[]
  meta?: Record<string, KunImageMeta>
  zoomable?: boolean
}>()

const shown = computed(() => props.images.slice(0, 9))
const isSingle = computed(() => shown.value.length === 1)

const metaOf = (token: string): KunImageMeta | undefined => props.meta?.[token]

const aspectOf = (token: string): string | undefined => {
  const m = metaOf(token)
  return m?.width && m?.height ? `${m.width} / ${m.height}` : undefined
}

const SINGLE_MAX_HEIGHT_PX = 384

const singleWidth = computed(() => {
  const m = metaOf(shown.value[0]!)
  if (!m?.width || !m?.height) {
    return undefined
  }
  const heightCapped = Math.round((SINGLE_MAX_HEIGHT_PX * m.width) / m.height)
  return { width: `min(${m.width}px, 100%, ${heightCapped}px)` }
})

const root = ref<HTMLElement | null>(null)
const lightboxRoot = props.zoomable ? root : ref<HTMLElement | null>(null)

const {
  isLightboxOpen,
  images: lightboxImages,
  currentImageIndex
} = useContentLightbox(lightboxRoot)
</script>

<template>
  <div v-if="shown.length">
    <div ref="root">
      <KunImage
        v-if="isSingle"
        :src="imageTokenUrl(shown[0]!)"
        :thumbhash="metaOf(shown[0]!)?.thumbhash"
        :aspect-ratio="aspectOf(shown[0]!)"
        :width="metaOf(shown[0]!)?.width"
        :height="metaOf(shown[0]!)?.height"
        :style="singleWidth"
        alt="话题封面"
        loading="lazy"
        object-fit="cover"
        :class-name="cn('rounded-lg', zoomable && 'cursor-zoom-in')"
      />

      <KunScrollShadow
        v-else
        axis="horizontal"
        shadow-size="2rem"
        scrollbar="thin"
      >
        <div class="flex gap-1.5">
          <KunImage
            v-for="(token, idx) in shown"
            :key="`${idx}-${token}`"
            :src="imageTokenUrl(token)"
            :thumbhash="metaOf(token)?.thumbhash"
            :aspect-ratio="aspectOf(token)"
            alt="话题封面"
            loading="lazy"
            object-fit="contain"
            :class-name="
              cn(
                'h-40 w-auto shrink-0 rounded-lg',
                zoomable && 'cursor-zoom-in'
              )
            "
          />
        </div>
      </KunScrollShadow>
    </div>

    <KunLightbox
      v-if="zoomable"
      v-model:is-open="isLightboxOpen"
      :images="lightboxImages"
      :initial-index="currentImageIndex"
    />
  </div>
</template>
