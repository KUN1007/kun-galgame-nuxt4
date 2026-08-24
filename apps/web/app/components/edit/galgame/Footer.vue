<script setup lang="ts">
import { submitGalgameSchema } from '~/validations/galgame'

const {
  name,
  content_limit,
  age_limit,
  original_language,
  introduction,
  aliases,
  release_date,
  release_date_tba
} = storeToRefs(usePersistEditGalgameStore())

const isPublishing = ref(false)

const handleSubmitGalgame = async () => {
  const banner = await getImage('kun-galgame-publish-banner')
  const data: Record<
    string,
    number | string | string[] | Blob | boolean | null
  > = {
    name_en_us: name.value['en-us'],
    name_ja_jp: name.value['ja-jp'],
    name_zh_cn: name.value['zh-cn'],
    name_zh_tw: name.value['zh-tw'],
    intro_en_us: introduction.value['en-us'],
    intro_ja_jp: introduction.value['ja-jp'],
    intro_zh_cn: introduction.value['zh-cn'],
    intro_zh_tw: introduction.value['zh-tw'],
    content_limit: content_limit.value,
    age_limit: age_limit.value,
    original_language: original_language.value,
    release_date: release_date.value,
    release_date_tba: release_date_tba.value,
    banner
  }
  const result = submitGalgameSchema.safeParse(data)
  if (!result.success) {
    const message = JSON.parse(result.error.message)[0]
    useMessage(formatKunZodIssue(message), 'warn')
    return
  }
  const res = await useComponentMessageStore().alert(
    '确定提交 Galgame 申请吗?',
    '提交后将进入审核队列, 审核通过后才会被公开展示。审核结果会通过站内消息通知您。在审核期间您可以在「我的提交」页继续编辑或撤回。'
  )
  if (!res) {
    return
  }

  if (isPublishing.value) {
    return
  } else {
    isPublishing.value = true
    useMessage(10525, 'info', 7777)
  }

  const { banner: _bannerBlob, ...jsonFields } = data
  let bannerHash = ''
  if (banner) {
    const uploaded = await uploadGalgameImage(banner, 'galgame_banner')
    if (uploaded) {
      bannerHash = uploaded.hash
    }
  }
  const created = await kunFetch<{
    gid: number
    claim_state: string
    banner_attached: boolean
  }>('/galgame/submit', {
    method: 'POST',
    body: {
      ...jsonFields,
      aliases: aliases.value,
      banner_hash: bannerHash
    }
  })
  isPublishing.value = false

  if (created?.gid) {
    if (bannerHash && !created.banner_attached) {
      useMessage('封面上传失败, 请在「我的提交」中重新添加封面', 'warn', 7777)
    }
    await deleteImage('kun-galgame-publish-banner')
    if (import.meta.client) {
      localStorage.removeItem('kun-galgame-publish-step')
    }

    useKunLoliInfo('Galgame 申请已提交, 等待审核', 5)
    await navigateTo('/edit/galgame/mine')
    usePersistEditGalgameStore().resetEditGalgameStore()
  }
}
</script>

<template>
  <div class="flex justify-end">
    <KunButton
      :disabled="isPublishing"
      :loading="isPublishing"
      size="lg"
      @click="handleSubmitGalgame"
    >
      提交 Galgame 申请
    </KunButton>
  </div>
</template>
