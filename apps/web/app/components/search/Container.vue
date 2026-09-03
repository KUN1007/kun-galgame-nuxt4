<script setup lang="ts">
import { navItems } from './items'

const { keywords } = storeToRefs(useTempSearchStore())

// `keywords` lives in a temp store the whole page shares, but the URL is what
// the command palette, a shared link and the back button all hand over, so the
// two are kept in step here.
const route = useRoute()
const router = useRouter()

const queryKeywords = computed(() => {
  const value = route.query.keywords
  return (Array.isArray(value) ? value[0] : value) ?? ''
})

watch(queryKeywords, (value) => (keywords.value = value), { immediate: true })

watch(keywords, (value) => {
  if (value === queryKeywords.value) {
    return
  }
  const query = { ...route.query }
  if (value) {
    query.keywords = value
  } else {
    delete query.keywords
  }
  router.replace({ query })
})

interface SearchPage {
  items: SearchResult[]
  total: number
}

const results = ref<SearchResult[]>([])
const total = ref(0)
const isLoading = ref(false)
const activeType = useTabQuery('topic', 'type')
const pageData = reactive({
  page: 1,
  limit: 12
})

const isLoadComplete = computed(
  () => total.value > 0 && results.value.length >= total.value
)

const searchQuery = async (searchType?: string): Promise<SearchPage> => {
  // Nothing awaits this watcher on the server, so an SSR fetch is a request
  // whose answer is thrown away and then made again on hydration.
  if (import.meta.server) {
    return { items: [], total: 0 }
  }
  isLoading.value = true
  const type = searchType || activeType.value
  if (type === 'toolset') {
    const result = await kunFetch<SearchPage>('/toolset', {
      method: 'GET',
      query: {
        query: keywords.value,
        ...pageData
      }
    })
    isLoading.value = false
    return result ?? { items: [], total: 0 }
  }
  const result = await kunFetch<SearchPage>('/search', {
    method: 'GET',
    query: {
      keywords: keywords.value,
      type,
      ...pageData
    }
  })
  isLoading.value = false
  return result ?? { items: [], total: 0 }
}

const handleSetType = async (value: SearchType) => {
  activeType.value = value
  pageData.page = 1
  results.value = []
  total.value = 0

  if (keywords.value) {
    const page = await searchQuery(value)
    results.value = page.items
    total.value = page.total
  } else {
    isLoading.value = false
  }
}

watch(
  () => keywords.value,
  async () => {
    pageData.page = 1

    if (keywords.value) {
      const page = await searchQuery()
      results.value = page.items
      total.value = page.total
    } else {
      results.value = []
      total.value = 0
      isLoading.value = false
    }
  },
  { immediate: true }
)

const handleLoadMore = async () => {
  pageData.page++
  const page = await searchQuery()
  results.value = results.value.concat(page.items)
  total.value = page.total
}
</script>

<template>
  <div class="min-h-[calc(100dvh-6rem)] space-y-6">
    <KunHeader
      name="搜索"
      description="您可以在本页面搜索本论坛的所有话题, Galgame, 用户, 回复, 评论。"
    >
      <template #endContent>
        <div class="text-default-500">
          当前的搜索会一并搜索 NSFW 内容, 如果您要按照 Galgame 厂商 / 会社 /
          标签搜索, 或者需要 <KunLink to="/galgame/tag">多标签搜索</KunLink> ,
          请前往
          <KunLink to="/galgame/official"> Galgame 会社资料库 </KunLink>
          或者
          <KunLink to="/galgame/tag"> Galgame 标签资料库 </KunLink>
          的页面进行搜索。
        </div>
      </template>
    </KunHeader>
    <KunTab
      :items="navItems"
      :model-value="activeType"
      @update:model-value="(value) => handleSetType(value as SearchType)"
      size="sm"
    />

    <SearchBox />

    <SearchHistory v-if="!keywords" />

    <SearchResult
      :results="results"
      :type="activeType as SearchType"
      v-if="results.length"
    />

    <KunDivider v-if="results.length >= 12">
      <slot />
      <KunButton
        variant="flat"
        :loading="isLoading"
        :disabled="isLoading || isLoadComplete"
        @click="handleLoadMore"
      >
        加载更多
      </KunButton>
      <span v-if="isLoadComplete">被榨干了呜呜呜呜呜, 一滴也不剩了</span>
    </KunDivider>

    <KunNull
      v-if="!results.length && keywords && !isLoading"
      description="杂鱼杂鱼杂鱼~什么也没有搜索到"
    />

    <KunLoading v-if="isLoading" />
  </div>
</template>
