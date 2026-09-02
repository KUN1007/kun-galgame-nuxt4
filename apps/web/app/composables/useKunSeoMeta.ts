import type {
  ActiveHeadEntry,
  UseHeadOptions,
  UseSeoMetaInput
} from '@unhead/vue'
import type { NuxtApp } from '#app/nuxt'

interface NuxtUseHeadOptions extends UseHeadOptions {
  nuxt?: NuxtApp
}

type KunSeoMetaInput = Omit<
  UseSeoMetaInput,
  | 'ogUrl'
  | 'ogTitle'
  | 'ogDescription'
  | 'twitterCard'
  | 'twitterTitle'
  | 'twitterDescription'
  | 'twitterImage'
  | 'twitterImageAlt'
> & {
  ogCard?: KunOgCardRef
}

export const useKunSeoMeta = (
  input: KunSeoMetaInput,
  options?: NuxtUseHeadOptions
  // eslint-disable-next-line @typescript-eslint/no-invalid-void-type
): ActiveHeadEntry<UseSeoMetaInput> | void => {
  const { ogCard, ogImage: explicitImage, ...rest } = input

  const title = `${input.title?.toString()} - ${kungal.title}`
  const description = input.description?.toString()
  const route = useRoute()

  const cardsEnabled = useRuntimeConfig().public.ogCardEnabled
  const site = kunOgSiteImage(cardsEnabled, kungal.domain.main)

  const pageUrl = `${kungal.domain.main}${route.path}`
  const cardUrl =
    ogCard && cardsEnabled
      ? `${kungal.domain.main}${kunOgCardPath(ogCard.kind, ogCard.id)}`
      : ''
  const image = cardUrl || explicitImage || site.url
  // A caller's own image is of unknown size, so the site-wide width/height must be cleared
  // rather than left to describe a different picture.
  const size = cardUrl ? KUN_OG_CARD_SIZE : explicitImage ? null : site

  useSeoMeta(
    {
      title,
      description,
      keywords: kungal.keywords.toString(),
      ogUrl: pageUrl,
      ogType: input.ogType || 'website',
      ogTitle: title,
      ogDescription: description,
      ogImage: image,
      ogImageAlt: title,
      ogImageWidth: size ? size.width : false,
      ogImageHeight: size ? size.height : false,
      twitterCard: 'summary_large_image',
      twitterTitle: title,
      twitterDescription: description,
      twitterImage: image,
      twitterImageAlt: title,
      ...rest
    },
    options
  )

  useHead({
    link: [{ rel: 'canonical', href: pageUrl }]
  })
}
