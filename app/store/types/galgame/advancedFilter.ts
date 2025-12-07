import type { ProviderKey } from '~/constants/galgameResource'

export interface GalgameAdvancedFilterPersist {
  includeProviders: ProviderKey[]
  excludeOnlyProviders: ProviderKey[]
}
