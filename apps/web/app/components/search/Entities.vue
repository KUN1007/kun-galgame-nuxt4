<script setup lang="ts">
import { SEARCH_ENTITY_FAMILIES } from './items'

const props = defineProps<{
  keywords: string
}>()

const ENTITY_LIMIT_ALL = 8
const ENTITY_LIMIT_ONE = 24

const family = useTabQuery('all', 'family')

const familyItems = computed(() => [
  { value: 'all', textValue: '全部', icon: 'lucide:layout-grid' },
  ...SEARCH_ENTITY_FAMILIES
])

const result = ref<SearchEntityResult | null>(null)
const pending = ref(!!props.keywords)
const failed = ref(false)
const page = ref(1)
const top = useTemplateRef<HTMLElement>('top')

const isAll = computed(() => family.value === 'all')

let latest = 0

const load = async () => {
  const current = ++latest
  if (!props.keywords) {
    result.value = null
    pending.value = false
    return
  }
  pending.value = true
  const data = await kunFetch<SearchEntityResult>('/search/entity', {
    method: 'GET',
    query: {
      keywords: props.keywords,
      family: isAll.value ? undefined : family.value,
      page: page.value,
      limit: isAll.value ? ENTITY_LIMIT_ALL : ENTITY_LIMIT_ONE
    }
  })
  if (current !== latest) {
    return
  }
  result.value = data
  failed.value = !data
  pending.value = false
}

watch(page, async () => {
  await load()
  top.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
})

watch(
  [() => props.keywords, family],
  () => {
    page.value = 1
    result.value = null
    load()
  },
  { immediate: true }
)

const groups = computed(() => result.value?.groups ?? [])
const isEmpty = computed(
  () =>
    !pending.value &&
    !failed.value &&
    groups.value.every((group) => !group.items.length)
)

// One family per request, so the single-family tab is the only one with a page
// count to divide: 全部 is a preview of every family at once.
const totalPage = computed(() =>
  isAll.value ? 0 : Math.ceil((groups.value[0]?.total ?? 0) / ENTITY_LIMIT_ONE)
)
</script>

<template>
  <div ref="top" class="scroll-mt-40 space-y-5">
    <KunTab
      :items="familyItems"
      :model-value="family"
      variant="pills"
      size="sm"
      :scrollable="true"
      @update:model-value="(value) => (family = value)"
    />

    <SearchSkeleton v-if="pending && !groups.length" shape="entity" />

    <template v-else>
      <KunLoading :loading="pending">
        <div class="space-y-5">
          <SearchEntityGroup
            v-for="group in groups"
            :key="group.family"
            :group="group"
            :keywords="keywords"
            :show-header="isAll"
            :show-cap="isAll"
            @open="(value) => (family = value)"
          />
        </div>
      </KunLoading>

      <KunNull v-if="failed" description="资料库搜索没能完成, 请稍后重试" />
      <KunNull v-else-if="isEmpty" description="资料库里没有找到匹配的条目" />

      <KunPagination
        v-if="totalPage > 1"
        v-model:current-page="page"
        :total-page="totalPage"
        :is-loading="pending"
      />
    </template>
  </div>
</template>
