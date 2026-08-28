<script setup lang="ts">
import { watchDebounced } from '@vueuse/core'

const pageData = reactive({
  page: 1,
  limit: 100
})

const { data, status } = await useKunFetch<{
  tags: GalgameTagItem[]
  total: number
}>(`/galgame-tag`, {
  method: 'GET',
  query: pageData
})

const { showKUNGalgameContentLimit } = storeToRefs(usePersistSettingsStore())
const isSfwMode = computed(() => showKUNGalgameContentLimit.value !== 'nsfw')

const searchResult = ref<GalgameTaxonomySearchItem[]>([])
const searchQuery = ref('')
const isSearching = ref(false)

const searchMode = ref<'single' | 'multi'>('single')
const searchModeOptions = [
  { label: '单标签搜索', value: 'single' },
  { label: '多标签搜索', value: 'multi' }
] as const

const matchMode = ref<'exact' | 'contains'>('exact')
const matchModeOptions = [
  { label: '精确模式', value: 'exact' },
  { label: '包含模式', value: 'contains' }
] as const

const displayTags = computed(() => {
  if (searchMode.value === 'single') {
    if (searchQuery.value.trim()) return searchResult.value
    return data.value?.tags ?? []
  }
  return data.value?.tags ?? []
})

const isBrowsingList = computed(() =>
  searchMode.value === 'single'
    ? !searchQuery.value.trim()
    : !selectedTags.value.length
)

const inputFocused = ref(false)
const isDropdownOpen = computed(
  () =>
    searchMode.value === 'multi' &&
    inputFocused.value &&
    !!searchQuery.value.trim() &&
    !isSearching.value &&
    (!!searchResult.value.length || isSfwMode.value)
)

const handleSearch = async () => {
  if (!searchQuery.value.trim()) {
    searchResult.value = []
    return
  }
  isSearching.value = true
  const res = await kunFetch<GalgameTaxonomySearchItem[]>(
    `/galgame-tag/search`,
    {
      method: 'GET',
      query: { q: searchQuery.value }
    }
  )
  isSearching.value = false

  searchResult.value = res ?? []
}

watchDebounced(
  () => searchQuery.value,
  () => {
    handleSearch()
  },
  { debounce: 500, maxWait: 1000 }
)

const onInputBlur = () => {
  setTimeout(() => {
    inputFocused.value = false
  }, 120)
}

const selectedTags = ref<GalgameTaxonomySearchItem[]>([])
const resultGames = ref<GalgameCard[]>([])
const totalGameCount = ref(0)
const gamesPage = ref(1)
const gamesLimit = 24
const loadingGames = ref(false)

const fetchGames = async () => {
  if (!selectedTags.value.length) {
    resultGames.value = []
    totalGameCount.value = 0
    return
  }
  loadingGames.value = true
  const res = await kunFetch<{ galgames: GalgameCard[]; total: number }>(
    `/galgame-tag/multi`,
    {
      method: 'GET',
      query: {
        page: gamesPage.value,
        limit: gamesLimit,
        mode: matchMode.value,
        tag_ids: selectedTags.value.map((t) => t.id).join(',')
      }
    }
  )
  loadingGames.value = false
  if (res) {
    resultGames.value = res.galgames
    totalGameCount.value = res.total
  }
}

const addTag = (tag: GalgameTaxonomySearchItem) => {
  if (selectedTags.value.find((t) => t.id === tag.id)) return
  selectedTags.value.push(tag)
  gamesPage.value = 1
  fetchGames()
}

const removeTag = (id: number) => {
  selectedTags.value = selectedTags.value.filter((t) => t.id !== id)
  gamesPage.value = 1
  fetchGames()
}

watch(
  () => gamesPage.value,
  () => {
    fetchGames()
  }
)

watch(matchMode, () => {
  gamesPage.value = 1
  fetchGames()
})
</script>

<template>
  <div class="space-y-6">
    <KunHeader
      name="Galgame 标签资料库"
      description="这里展示了绝大多数 Galgame 的标签, 并附带有标签的别名, 您可以点击标签以查看所有含有这个标签的 Galgame"
    >
      <template #endContent>
        <div class="space-y-3">
          <p class="text-default-500">
            默认仅显示了 SFW 的标签, 成人标签既不会出现在列表里, 也搜不到,
            查看它们请在设置面板打开 NSFW 开关。如果有数据错误请
            <KunLink to="/doc/contact"> 联系我们 </KunLink>。
          </p>

          <div
            v-if="searchMode === 'multi' && selectedTags.length"
            class="flex flex-wrap gap-2"
          >
            <KunChip
              v-for="t in selectedTags"
              :key="t.id"
              color="primary"
              class="flex items-center gap-2"
            >
              {{ t.name }}
              <KunButton
                size="sm"
                :is-icon-only="true"
                variant="light"
                @click.stop="removeTag(t.id)"
              >
                <KunIcon name="lucide:x" />
              </KunButton>
            </KunChip>
          </div>

          <div class="flex items-center gap-2">
            <KunSelect
              v-model="searchMode"
              :options="searchModeOptions"
              aria-label="search-mode"
              class-name="w-36"
            />

            <KunSelect
              v-if="searchMode === 'multi'"
              v-model="matchMode"
              :options="matchModeOptions"
              aria-label="match-mode"
              class-name="w-36"
            />

            <div class="relative flex-1">
              <KunInput
                v-model="searchQuery"
                type="text"
                placeholder="输入将会自动搜索标签, 点击标签使用多标签筛选"
                @focus="inputFocused = true"
                @blur="onInputBlur"
              />

              <div
                v-if="isDropdownOpen"
                class="border-default-200 bg-content1 absolute top-full right-0 left-0 z-50 mt-2 rounded-lg border shadow-md"
              >
                <KunScrollShadow
                  axis="vertical"
                  shadow-size="3rem"
                  class-name="max-h-96"
                >
                  <div
                    v-for="tag in searchResult"
                    :key="tag.id"
                    class="hover:bg-default-100 flex cursor-pointer items-center justify-between px-3 py-2"
                    @mousedown.prevent="
                      () => {
                        addTag(tag)
                        searchQuery = ''
                      }
                    "
                  >
                    <div class="truncate">{{ tag.name }}</div>
                  </div>

                  <div
                    v-if="!searchResult.length"
                    class="text-default-400 px-3 py-2 text-sm"
                  >
                    没有匹配的标签, 成人标签需要打开 NSFW 开关才能搜到
                  </div>
                </KunScrollShadow>
              </div>
            </div>
          </div>

          <div
            v-if="searchMode === 'single'"
            class="text-default-600 mt-4 flex items-center gap-4 text-sm"
          >
            <span>{{ `共 ${data?.total || 0} 个标签` }}</span>
          </div>
        </div>
      </template>
    </KunHeader>

    <div
      v-if="searchMode === 'single' || !selectedTags.length"
      class="grid grid-cols-2 gap-3 sm:grid-cols-2 sm:gap-3 lg:grid-cols-3 xl:grid-cols-4"
    >
      <GalgameTagCard v-for="tag in displayTags" :key="tag.id" :tag="tag" />
    </div>

    <KunNull
      v-if="
        !isSearching &&
        !displayTags.length &&
        (searchMode === 'single' || !selectedTags.length)
      "
      :description="
        isSfwMode
          ? '没有匹配的标签。成人标签在 SFW 模式下搜不到, 请在设置面板打开 NSFW 开关'
          : undefined
      "
    />
    <KunLoading
      v-if="isSearching && (searchMode === 'single' || !selectedTags.length)"
    />

    <KunPagination
      v-if="isBrowsingList && data && data.total > pageData.limit"
      v-model:current-page="pageData.page"
      :total-page="Math.ceil(data.total / pageData.limit)"
      :is-loading="status === 'pending'"
    />

    <div v-if="searchMode === 'multi' && selectedTags.length">
      <KunLoading :loading="loadingGames">
        <GalgameCard v-if="resultGames.length" :galgames="resultGames" />
      </KunLoading>
      <KunNull
        v-if="!loadingGames && !resultGames.length && selectedTags.length"
      />
      <KunPagination
        class="mt-3"
        v-if="totalGameCount > gamesLimit"
        v-model:current-page="gamesPage"
        :total-page="Math.ceil(totalGameCount / gamesLimit)"
        :is-loading="loadingGames"
      />
    </div>
  </div>
</template>
