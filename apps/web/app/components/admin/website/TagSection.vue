<script setup lang="ts">
import type {
  CreateWebsiteTagPayload,
  UpdateWebsiteTagPayload,
  CreateWebsiteTagGroupPayload,
  UpdateWebsiteTagGroupPayload
} from '~/components/website/modal/types'

const { data: tags, refresh: refreshTags } = useKunFetch<WebsiteTag[]>(
  '/website-tag',
  { key: 'website-tag' }
)
const { data: groups, refresh: refreshGroups } = useWebsiteTagGroups()

const canCreate = useCan('website.create')
const canDelete = useCan('website.delete')

const UNGROUPED_ID = -1

const sections = computed(() => {
  const byGroup = new Map<number, WebsiteTag[]>()
  for (const tag of tags.value ?? []) {
    const key = tag.group_id ?? UNGROUPED_ID
    if (!byGroup.has(key)) {
      byGroup.set(key, [])
    }
    byGroup.get(key)!.push(tag)
  }
  for (const list of byGroup.values()) {
    list.sort((a, b) => b.level - a.level)
  }

  const result = (groups.value ?? []).map((group) => ({
    group,
    tags: byGroup.get(group.id) ?? []
  }))

  const ungrouped = byGroup.get(UNGROUPED_ID) ?? []
  if (ungrouped.length) {
    result.push({ group: null as unknown as WebsiteTagGroup, tags: ungrouped })
  }
  return result
})

const isTagModalOpen = ref(false)
const isGroupModalOpen = ref(false)
const isSubmitting = ref(false)
const editingTag = ref<
  (CreateWebsiteTagPayload & { tag_id?: number }) | undefined
>(undefined)
const editingGroup = ref<
  (CreateWebsiteTagGroupPayload & { group_id?: number }) | undefined
>(undefined)

const openCreateTag = (groupId: number | null) => {
  editingTag.value = {
    name: '',
    label: '',
    level: 0,
    description: '',
    group_id: groupId
  }
  isTagModalOpen.value = true
}

const openEditTag = (tag: WebsiteTag) => {
  editingTag.value = {
    name: tag.name,
    label: tag.label,
    level: tag.level,
    description: tag.description,
    group_id: tag.group_id,
    tag_id: tag.id
  }
  isTagModalOpen.value = true
}

const handleTagSubmit = async (
  payload: CreateWebsiteTagPayload | UpdateWebsiteTagPayload
) => {
  isSubmitting.value = true
  const result = await kunFetch('/website-tag', {
    method: 'tag_id' in payload ? 'PUT' : 'POST',
    body: payload
  })
  isSubmitting.value = false

  if (result) {
    useMessage('tag_id' in payload ? '标签已更新' : '标签已创建', 'success')
    isTagModalOpen.value = false
    await refreshTags()
  }
}

const handleTagDelete = async (tag: WebsiteTag) => {
  const confirmed = await useComponentMessageStore().alert(
    `确定删除标签「${tag.label}」吗？`,
    '所有网站身上的这个标签会一并移除, 该操作不可撤销'
  )
  if (!confirmed) {
    return
  }

  const result = await kunFetch('/website-tag', {
    method: 'DELETE',
    query: { tag_id: tag.id }
  })
  if (result) {
    useMessage('标签已删除', 'success')
    await refreshTags()
  }
}

const openCreateGroup = () => {
  editingGroup.value = undefined
  isGroupModalOpen.value = true
}

const openEditGroup = (group: WebsiteTagGroup) => {
  editingGroup.value = {
    name: group.name,
    label: group.label,
    description: group.description,
    sort_order: group.sort_order,
    multi_select: group.multi_select,
    group_id: group.id
  }
  isGroupModalOpen.value = true
}

const handleGroupSubmit = async (
  payload: CreateWebsiteTagGroupPayload | UpdateWebsiteTagGroupPayload
) => {
  isSubmitting.value = true
  const result = await kunFetch('/website-tag-group', {
    method: 'group_id' in payload ? 'PUT' : 'POST',
    body: payload
  })
  isSubmitting.value = false

  if (result) {
    useMessage('group_id' in payload ? '分组已更新' : '分组已创建', 'success')
    isGroupModalOpen.value = false
    await refreshGroups()
  }
}

const handleGroupDelete = async (group: WebsiteTagGroup) => {
  const confirmed = await useComponentMessageStore().alert(
    `确定删除分组「${group.label}」吗？`,
    '组内的标签不会被删除, 它们会落到「未分组」里'
  )
  if (!confirmed) {
    return
  }

  const result = await kunFetch('/website-tag-group', {
    method: 'DELETE',
    query: { group_id: group.id }
  })
  if (result) {
    useMessage('分组已删除', 'success')
    await Promise.all([refreshGroups(), refreshTags()])
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <p class="text-default-500 text-sm">
        标签决定网站的价值精算值,
        分组决定它们在创建网站表单里的排布。互斥分组只能选一个标签,
        多选分组可以任意勾选。
      </p>
      <div class="flex flex-wrap gap-3">
        <KunButton v-if="canCreate" variant="flat" @click="openCreateGroup">
          <KunIcon name="lucide:folder-plus" />
          创建分组
        </KunButton>
        <KunButton
          v-if="canCreate"
          color="primary"
          @click="openCreateTag(null)"
        >
          <KunIcon name="lucide:plus" />
          创建标签
        </KunButton>
      </div>
    </div>

    <div v-for="section in sections" :key="section.group?.id ?? UNGROUPED_ID">
      <div class="mb-2 flex flex-wrap items-center gap-2">
        <h2 class="text-default-900 text-xl font-bold">
          {{ section.group?.label ?? '未分组' }}
        </h2>
        <span v-if="section.group" class="text-default-400 font-mono text-xs">
          {{ section.group.name }}
        </span>
        <KunChip>{{ section.tags.length }} 个标签</KunChip>
        <KunChip v-if="section.group?.multi_select" color="primary">
          多选
        </KunChip>
        <KunChip v-else-if="section.group" color="secondary">互斥</KunChip>

        <template v-if="section.group">
          <KunButton
            size="sm"
            variant="light"
            :is-icon-only="true"
            @click="openEditGroup(section.group)"
          >
            <KunIcon name="lucide:pencil" />
          </KunButton>
          <KunButton
            v-if="canDelete"
            size="sm"
            variant="light"
            color="danger"
            :is-icon-only="true"
            @click="handleGroupDelete(section.group)"
          >
            <KunIcon name="lucide:trash-2" />
          </KunButton>
          <KunButton
            v-if="canCreate"
            size="sm"
            variant="light"
            @click="openCreateTag(section.group.id)"
          >
            <KunIcon name="lucide:plus" />
            加标签
          </KunButton>
        </template>
      </div>

      <div class="grid grid-cols-1 gap-2 md:grid-cols-2">
        <KunCard
          v-for="tag in section.tags"
          :key="tag.id"
          :is-hoverable="false"
          :is-transparent="false"
          padding="sm"
        >
          <div class="flex items-center gap-3">
            <span
              :class="
                cn(
                  'w-12 shrink-0 text-center font-mono text-sm font-bold',
                  tag.level > 0
                    ? 'text-success-600'
                    : tag.level < 0
                      ? 'text-danger-600'
                      : 'text-default-400'
                )
              "
            >
              {{ tag.level > 0 ? '+' : '' }}{{ tag.level }}
            </span>
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <KunLink :to="`/website-tag/${tag.name}`" underline="none">
                  <span class="font-medium">{{ tag.label || tag.name }}</span>
                </KunLink>
                <span class="text-default-400 font-mono text-xs">
                  {{ tag.name }}
                </span>
              </div>
              <p v-if="tag.description" class="text-default-500 text-xs">
                {{ tag.description }}
              </p>
            </div>
            <KunButton
              size="sm"
              variant="light"
              :is-icon-only="true"
              @click="openEditTag(tag)"
            >
              <KunIcon name="lucide:pencil" />
            </KunButton>
            <KunButton
              v-if="canDelete"
              size="sm"
              variant="light"
              color="danger"
              :is-icon-only="true"
              @click="handleTagDelete(tag)"
            >
              <KunIcon name="lucide:trash-2" />
            </KunButton>
          </div>
        </KunCard>
      </div>

      <div v-if="!section.tags.length" class="text-default-400 py-3 text-sm">
        这个分组下还没有标签
      </div>
    </div>

    <KunNull v-if="!sections.length" description="还没有任何网站标签" />

    <WebsiteModalTag
      v-model="isTagModalOpen"
      :initial-data="editingTag"
      :loading="isSubmitting"
      @submit="handleTagSubmit"
    />

    <WebsiteModalTagGroup
      v-model="isGroupModalOpen"
      :initial-data="editingGroup"
      :loading="isSubmitting"
      @submit="handleGroupSubmit"
    />
  </div>
</template>
