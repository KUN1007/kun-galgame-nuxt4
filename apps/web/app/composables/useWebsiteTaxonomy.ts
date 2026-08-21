export const useWebsiteCategories = () =>
  useKunFetch<WebsiteCategoryListItem[]>('/website-category', {
    key: 'website-category'
  })

export const useWebsiteTagGroups = () =>
  useKunFetch<WebsiteTagGroup[]>('/website-tag-group', {
    key: 'website-tag-group'
  })
