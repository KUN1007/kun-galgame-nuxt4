<script setup lang="ts">
import { SEARCH_CATEGORY_MAP } from './items'

const props = defineProps<{
  keywords: string
  type: SearchPagedType
}>()

const PAGE_SIZE = 12

interface SearchPage {
  items: SearchResult[]
  total: number
}

const results = ref<SearchResult[]>([])
const total = ref(0)
const pending = ref(!!props.keywords)
const failed = ref(false)
const page = ref(1)
const top = useTemplateRef<HTMLElement>('top')

const meta = computed(() => SEARCH_CATEGORY_MAP[props.type])
const totalPage = computed(() => Math.ceil(total.value / PAGE_SIZE))

let latest = 0

const fetchPage = (target: number) =>
  props.type === 'toolset'
    ? kunFetch<SearchPage>('/toolset', {
        method: 'GET',
        query: { query: props.keywords, page: target, limit: PAGE_SIZE }
      })
    : kunFetch<SearchPage>('/search', {
        method: 'GET',
        query: {
          keywords: props.keywords,
          type: props.type,
          page: target,
          limit: PAGE_SIZE
        }
      })

const load = async () => {
  const current = ++latest
  if (!props.keywords) {
    results.value = []
    total.value = 0
    pending.value = false
    return
  }
  pending.value = true
  const data = await fetchPage(page.value)
  if (current !== latest) {
    return
  }
  // kunFetch already popped a toast; telling the reader "nothing found" when the
  // request never came back is the one thing this must not do.
  failed.value = !data
  results.value = data?.items ?? []
  total.value = data?.total ?? 0
  pending.value = false
}

watch(page, async () => {
  await load()
  // The paginator sits below a full page of results, so a page opened from it
  // would otherwise start scrolled past its own first row.
  top.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
})

watch(
  () => props.keywords,
  () => {
    page.value = 1
    results.value = []
    load()
  },
  { immediate: true }
)
</script>

<template>
  <div ref="top" class="scroll-mt-40 space-y-6">
    <p class="text-default-500 text-sm">
      <template v-if="pending && !results.length">正在搜索…</template>
      <template v-else-if="total">
        共 <span class="text-default-700 tabular-nums">{{ total }}</span>
        {{ meta.countUnit }}
      </template>
    </p>

    <SearchSkeleton
      v-if="pending && !results.length"
      :shape="type === 'galgame' || type === 'toolset' ? 'card' : 'row'"
    />

    <KunLoading v-else-if="results.length" :loading="pending">
      <SearchResult :results="results" :type="type" :keywords="keywords" />
    </KunLoading>

    <KunNull v-else-if="failed" description="搜索没能完成, 请稍后重试" />

    <KunNull v-else-if="keywords" description="杂鱼杂鱼杂鱼~什么也没有搜索到" />

    <KunPagination
      v-if="totalPage > 1"
      v-model:current-page="page"
      :total-page="totalPage"
      :is-loading="pending"
    />
  </div>
</template>
