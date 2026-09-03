<script setup lang="ts">
interface RosterRow {
  character_id: number
  kind: number
  spoiler: number
}

const KIND_OPTIONS = [
  { value: '0', label: '未分类' },
  { value: '1', label: '主角' },
  { value: '2', label: '配角' },
  { value: '3', label: '出场' }
]

const SPOILER_OPTIONS = [
  { value: '0', label: '无剧透' },
  { value: '1', label: '轻微' },
  { value: '2', label: '严重' }
]

const props = defineProps<{
  modelValue: unknown
  suppressed?: unknown
  disabled?: boolean
  names?: Map<number, string>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: unknown]
  'update:suppressed': [value: unknown]
}>()

const asRows = (value: unknown): RosterRow[] =>
  Array.isArray(value)
    ? (value as RosterRow[]).map((row) => ({
        character_id: Number(row.character_id),
        kind: Number(row.kind),
        spoiler: Number(row.spoiler)
      }))
    : []

const asKeys = (value: unknown): string[] =>
  Array.isArray(value) ? value.map((key) => String(key)) : []

const identityOf = (characterId: number) => `roster:${characterId}`

const nameOf = (characterId: number) =>
  props.names?.get(characterId) ?? `#${characterId}`

const rows = computed(() => asRows(props.modelValue))
const hidden = computed(() => new Set(asKeys(props.suppressed)))

const emitRows = (next: RosterRow[]) => emit('update:modelValue', next)

const emitHidden = (next: Set<string>) =>
  emit(
    'update:suppressed',
    [...next].sort((a, b) => (a < b ? -1 : a > b ? 1 : 0))
  )

const patchRow = (characterId: number, patch: Partial<RosterRow>) => {
  emitRows(
    rows.value.map((row) =>
      row.character_id === characterId ? { ...row, ...patch } : row
    )
  )
}

const toggleHidden = (characterId: number) => {
  const key = identityOf(characterId)
  const next = new Set(hidden.value)
  if (next.has(key)) {
    next.delete(key)
  } else {
    next.add(key)
  }
  emitHidden(next)
}

const filter = ref('')
const PREVIEW = 20
const expanded = ref(false)

const visible = computed(() => {
  const q = filter.value.trim().toLowerCase()
  const matched = q
    ? rows.value.filter((row) =>
        nameOf(row.character_id).toLowerCase().includes(q)
      )
    : rows.value
  return expanded.value || q ? matched : matched.slice(0, PREVIEW)
})

const foldedCount = computed(() => {
  if (expanded.value || filter.value.trim()) {
    return 0
  }
  return Math.max(0, rows.value.length - PREVIEW)
})
</script>

<template>
  <div class="space-y-3">
    <KunInput
      v-if="rows.length > PREVIEW"
      v-model="filter"
      placeholder="按角色名筛选"
    />

    <div v-if="!rows.length" class="text-default-400 text-sm">
      这部作品还没有出演名单。出演边由上游导入，不能在这里新建。
    </div>

    <div
      v-for="row in visible"
      :key="row.character_id"
      class="flex flex-wrap items-center gap-2"
    >
      <span
        class="min-w-0 flex-1 truncate text-sm"
        :class="
          hidden.has(identityOf(row.character_id)) &&
          'text-default-400 line-through'
        "
      >
        {{ nameOf(row.character_id) }}
      </span>
      <KunSelect
        :model-value="String(row.kind)"
        :options="KIND_OPTIONS"
        class-name="w-24 shrink-0"
        :disabled="disabled || hidden.has(identityOf(row.character_id))"
        @update:model-value="
          (v: string | number | (string | number)[] | null) =>
            patchRow(row.character_id, { kind: Number(v ?? 0) })
        "
      />
      <KunSelect
        :model-value="String(row.spoiler)"
        :options="SPOILER_OPTIONS"
        class-name="w-24 shrink-0"
        :disabled="disabled || hidden.has(identityOf(row.character_id))"
        @update:model-value="
          (v: string | number | (string | number)[] | null) =>
            patchRow(row.character_id, { spoiler: Number(v ?? 0) })
        "
      />
      <KunButton
        size="sm"
        variant="flat"
        :color="
          hidden.has(identityOf(row.character_id)) ? 'primary' : 'default'
        "
        :disabled="disabled"
        @click="toggleHidden(row.character_id)"
      >
        {{ hidden.has(identityOf(row.character_id)) ? '取消隐藏' : '隐藏' }}
      </KunButton>
    </div>

    <KunButton
      v-if="foldedCount > 0"
      variant="light"
      color="default"
      size="sm"
      @click="expanded = true"
    >
      还有 {{ foldedCount }} 名角色，全部显示
    </KunButton>
  </div>
</template>
