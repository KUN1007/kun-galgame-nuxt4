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

const meta = computed(() => SEARCH_CATEGORY_MAP[props.type])
const isComplete = computed(
  () => total.value > 0 && results.value.length >= total.value
)

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
  page.value = 1
  results.value = []
  total.value = 0
  if (!props.keywords) {
    pending.value = false
    return
  }
  pending.value = true
  const data = await fetchPage(1)
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

const loadMore = async () => {
  const current = latest
  pending.value = true
  const next = page.value + 1
  const data = await fetchPage(next)
  if (current !== latest) {
    return
  }
  if (data) {
    page.value = next
    results.value = results.value.concat(data.items)
    total.value = data.total
  }
  pending.value = false
}

watch(() => props.keywords, load, { immediate: true })
</script>

<template>
  <div class="space-y-6">
    <p class="text-default-500 text-sm">
      <template v-if="pending && !results.length">正在搜索…</template>
      <template v-else-if="results.length">
        共 <span class="text-default-700 tabular-nums">{{ total }}</span>
        {{ meta.countUnit }}
      </template>
    </p>

    <SearchSkeleton
      v-if="pending && !results.length"
      :shape="type === 'galgame' || type === 'toolset' ? 'card' : 'row'"
    />

    <SearchResult
      v-if="results.length"
      :results="results"
      :type="type"
      :keywords="keywords"
    />

    <KunNull v-if="!pending && failed" description="搜索没能完成, 请稍后重试" />

    <KunNull
      v-else-if="!pending && !results.length && keywords"
      description="杂鱼杂鱼杂鱼~什么也没有搜索到"
    />

    <KunDivider v-if="results.length">
      <KunButton
        v-if="!isComplete"
        variant="flat"
        :loading="pending"
        :disabled="pending"
        @click="loadMore"
      >
        加载更多
      </KunButton>
      <span v-else class="text-default-500 text-sm">
        被榨干了呜呜呜呜呜, 一滴也不剩了
      </span>
    </KunDivider>
  </div>
</template>
