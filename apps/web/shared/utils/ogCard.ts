export const KUN_OG_CARD_KINDS = [
  'topic',
  'galgame',
  'character',
  'staff',
  'official'
] as const

export type KunOgCardKind = (typeof KUN_OG_CARD_KINDS)[number]

export interface KunOgCardRef {
  kind: KunOgCardKind
  id: number
}

/**
 * The card image is served from our own origin rather than the renderer's: the HMAC key that
 * signs a card must never reach a browser, so the signed URL is built in Nitro and this path
 * only ever redirects to it.
 */
export const kunOgCardPath = (kind: KunOgCardKind, id: number): string =>
  `/og/${kind}/${id}`

/** Fixed by the renderer's templates; declared on every page that emits a card. */
export const KUN_OG_CARD_SIZE = { width: 1200, height: 630 } as const

/** public/kungalgame.webp, the fallback whenever card rendering is off. */
const KUN_OG_STATIC_SIZE = { width: 1672, height: 941 } as const

/**
 * The site-wide og:image. app.vue and useKunSeoMeta must agree on it — when they each picked
 * their own default, a page that set no card advertised the static artwork at the card's
 * 1200x630, which is neither image's real size.
 */
export const kunOgSiteImage = (enabled: boolean, origin: string) =>
  enabled
    ? { url: `${origin}/og/site`, ...KUN_OG_CARD_SIZE }
    : { url: `${origin}/kungalgame.webp`, ...KUN_OG_STATIC_SIZE }
