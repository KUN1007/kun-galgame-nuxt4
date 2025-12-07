import {
  type ProviderKey,
  PROVIDER_PATTERNS
} from '~/constants/galgameResource'

export const detectProviderFromUrl = (url?: string): ProviderKey => {
  if (!url) {
    return 'other'
  }
  const s = url.toLowerCase()
  for (const key of Object.keys(PROVIDER_PATTERNS) as ProviderKey[]) {
    if (key === 'other') {
      continue
    }

    for (const p of PROVIDER_PATTERNS[key]) {
      if (s.includes(p)) {
        return key
      }
    }
  }
  return 'other'
}

export const detectProvidersFromUrls = (
  urls: (string | undefined)[]
): ProviderKey[] => {
  const set = new Set<ProviderKey>()
  for (const u of urls || []) {
    const p = detectProviderFromUrl(u)
    set.add(p)
  }
  return Array.from(set)
}
