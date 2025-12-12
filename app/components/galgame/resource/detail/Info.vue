<script setup lang="ts">
const { id, role } = usePersistUserStore()
const { resources, rewriteResourceId } = storeToRefs(
  useTempGalgameResourceStore()
)
const props = defineProps<{
  details: GalgameResourceDetails
  refresh: () => void
}>()

const isFetching = ref(false)
const isResourceExpired = computed(() => props.details.status === 1)

const handleDeleteResource = async (
  galgameId: number,
  galgameResourceId: number
) => {
  const res = await useComponentMessageStore().alert(
    '您确定删除 Galgame 资源链接吗？',
    '这将会扣除您发布 Galgame 资源获得的 5 萌萌点，并且扣除其它人对资源链接的点赞影响（萌萌点和点赞数减一），此操作不可撤销。'
  )
  if (!res) {
    return
  }

  const result = await $fetch(`/api/galgame/${galgameId}/resource`, {
    method: 'DELETE',
    query: { galgameResourceId },
    watch: false,
    ...kungalgameResponseHandler
  })

  if (result) {
    props.refresh()
  }
}

const handleReportExpire = async (details: GalgameResourceDetails) => {
  if (!id) {
    useMessage(10546, 'warn')
    return
  }

  const res = await useComponentMessageStore().alert(
    '您确定报告资源链接失效吗？',
    '这将会通知资源发布者链接失效, 并将该链接标记为失效。若 17 天内资源发布者没有更换有效链接，该资源链接将会被删除。若恶意报告失效, 将会被处罚。'
  )
  if (!res) {
    return
  }

  isFetching.value = true
  const result = await $fetch(
    `/api/galgame/${details.galgameId}/resource/expired`,
    {
      method: 'PUT',
      body: { galgameResourceId: details.id },
      watch: false,
      ...kungalgameResponseHandler
    }
  )
  isFetching.value = false

  if (result) {
    useMessage(10547, 'success')
    props.refresh()
  }
}

const handleRewriteResource = (details: GalgameResourceDetails) => {
  resources.value[0] = { ...details }
  rewriteResourceId.value = details.id
}
</script>

<template>
  <div class="flex h-full flex-col gap-3" v-if="details">
    <div class="flex items-center gap-2">
      <KunAvatar :user="details.user" />
      <span>{{ details.user.name }}</span>
      <span class="text-default-500 text-sm">
        {{ `发布于 ${formatTimeDifference(details.created)}` }}
      </span>
    </div>

    <KunInfo
      variant="bordered"
      v-if="details.note"
      color="info"
      title="下载备注信息"
    >
      <pre class="break-all whitespace-pre-line">
        {{ details.note }}
      </pre>
    </KunInfo>

    <KunInfo
      :color="isResourceExpired ? 'warning' : 'success'"
      variant="bordered"
      title="下载链接 - 点击下面的链接以下载"
      class-name="relative"
    >
      <KunLink
        v-for="(kun, index) in details.link"
        :key="index"
        :to="kun"
        target="_blank"
        rel="noopener noreferrer"
        :is-show-anchor-icon="true"
      >
        {{ kun }}
      </KunLink>

      <div class="mt-3 flex items-center justify-end gap-2">
        <KunCopy
          variant="solid"
          :color="isResourceExpired ? 'warning' : 'success'"
          v-if="details.code"
          :name="`提取码 ${details.code}`"
          :text="details.code"
        />
        <KunCopy
          variant="solid"
          :color="isResourceExpired ? 'warning' : 'success'"
          v-if="details.password"
          :name="`解压码 ${details.password}`"
          :text="details.password"
        />
      </div>

      <KunBadge
        class="absolute -top-3 -right-3"
        :color="isResourceExpired ? 'danger' : 'success'"
        variant="solid"
      >
        {{
          isResourceExpired
            ? '该资源链接被其它用户标记为失效'
            : '该资源链接可用'
        }}
      </KunBadge>
    </KunInfo>

    <KunInfo title="补票提示信息" color="danger">
      <p>
        须知 Galgame 厂商制作游戏不易, 很多厂商如今都在炒冷饭,
        可见经济并不宽裕。如果条件允许, 请尽可能前往
        <KunLink size="sm" :to="`/galgame/${details.galgameId}`">
          Galgame 详情
        </KunLink>
        中的 Galgame 制作商部分 进行正版 Galgame 补票, 感谢您对 Galgame
        业界做出的贡献
      </p>
    </KunInfo>

    <KunInfo title="鲲的小请求">
      <p>
        在您下载这部 Galgame 并游玩之后, 可否请您在本网站的
        <KunLink size="sm" :to="`/galgame/${details.galgameId}`">
          Galgame 评分页面
        </KunLink>
        为这部 Galgame 提交一个评分, 这将有助于我们把优秀的 Galgame
        推荐给更多人, 谢谢您的支持
      </p>
    </KunInfo>

    <div
      class="mt-auto flex items-center justify-end gap-1"
      v-if="details.user.id === id || role > 1"
    >
      <KunButton
        :is-icon-only="true"
        variant="flat"
        @click="handleRewriteResource(details)"
      >
        编辑资源
        <KunIcon name="lucide:pencil" />
      </KunButton>
      <KunButton
        :is-icon-only="true"
        color="danger"
        variant="flat"
        @click="handleDeleteResource(details.galgameId, details.id)"
      >
        删除资源
        <KunIcon name="lucide:trash-2" />
      </KunButton>

      <div v-if="id !== details.user.id && !details.status">
        <KunButton
          variant="flat"
          color="danger"
          @click="handleReportExpire(details)"
          :loading="isFetching"
        >
          报告链接过期
        </KunButton>
      </div>

      <KunButton variant="flat" :href="`/galgame/${details.galgameId}`">
        反馈资源问题
      </KunButton>
    </div>
  </div>
</template>
