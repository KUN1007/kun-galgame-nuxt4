<script setup lang="ts">
import { SEARCH_ENTITY_FAMILIES } from './items'

const props = defineProps<{
  keywords: string
}>()

const ENTITY_LIMIT_ALL = 8
const ENTITY_LIMIT_ONE = 48

const family = useTabQuery('all', 'family')

const familyItems = computed(() => [
  { value: 'all', textValue: '全部', icon: 'lucide:layout-grid' },
  ...SEARCH_ENTITY_FAMILIES
])

const result = ref<SearchEntityResult | null>(null)
const pending = ref(!!props.keywords)
const failed = ref(false)

let latest = 0

const load = async () => {
  const current = ++latest
  if (!props.keywords) {
    result.value = null
    pending.value = false
    return
  }
  pending.value = true
  const isAll = family.value === 'all'
  const data = await kunFetch<SearchEntityResult>('/search/entity', {
    method: 'GET',
    query: {
      keywords: props.keywords,
      family: isAll ? undefined : family.value,
      limit: isAll ? ENTITY_LIMIT_ALL : ENTITY_LIMIT_ONE
    }
  })
  if (current !== latest) {
    return
  }
  result.value = data
  failed.value = !data
  pending.value = false
}

watch([() => props.keywords, family], load, { immediate: true })

const groups = computed(() => result.value?.groups ?? [])
const isEmpty = computed(
  () =>
    !pending.value &&
    !failed.value &&
    groups.value.every((group) => !group.items.length)
)
</script>

<template>
  <div class="space-y-5">
    <KunTab
      :items="familyItems"
      :model-value="family"
      variant="pills"
      size="sm"
      :scrollable="true"
      @update:model-value="(value) => (family = value)"
    />

    <SearchSkeleton v-if="pending" shape="entity" />

    <template v-else>
      <SearchEntityGroup
        v-for="group in groups"
        :key="group.family"
        :group="group"
        :keywords="keywords"
        :show-header="family === 'all'"
        :show-cap="true"
      />

      <KunNull v-if="failed" description="资料库搜索没能完成, 请稍后重试" />
      <KunNull v-else-if="isEmpty" description="资料库里没有找到匹配的条目" />
    </template>
  </div>
</template>
