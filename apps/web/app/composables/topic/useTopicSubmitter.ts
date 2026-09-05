import { createTopicSchema } from '~/validations/topic'
import { useTopicEditorStore } from './useTopicEditorStore'
import {
  TOPIC_SECTION_CONSUME_MOEMOEPOINTS,
  MOEMOEPOINT_COST_FOR_CONSUME_SECTION
} from '~/config/moemoepoint'

export const useTopicSubmitter = () => {
  const {
    category,
    section,
    title,
    content,
    isNSFW,
    coverImages,
    accessScope,
    accessRoles,
    accessUserIds
  } = useTopicEditorStore()
  const tempStore = useTempEditStore()
  const persistStore = usePersistEditTopicStore()
  const { moemoepoint } = usePersistUserStore()

  const rules = reactive({
    isReadRule: false,
    isAgreeCategory: false,
    isValidTitle: false,
    isKnownConsequence: false
  })
  const isSubmitting = ref(false)
  const isRewriteMode = computed(() => tempStore.isTopicRewriting)

  const submit = async () => {
    if (isSubmitting.value) {
      return
    }

    const isReadAllRules = Object.values(rules).every((value) => value)
    if (moemoepoint < 50 && !isReadAllRules) {
      useMessage('请勾选同意所有发布须知后再发布话题', 'warn')
      return
    }

    // PUT /topic/:tid answers 400 "无效的话题访问范围" when access_scope is
    // missing, so every submit carries it, a default-public topic included.
    const data = {
      title: title.value,
      content: content.value,
      category: category.value,
      section: section.value,
      is_nsfw: isNSFW.value,
      cover_images: coverImages.value,
      access_scope: accessScope.value,
      access_roles: accessScope.value === 'role' ? accessRoles.value : [],
      access_user_ids: accessScope.value === 'users' ? accessUserIds.value : []
    }

    const submitData = isRewriteMode.value
      ? { ...data, topic_id: tempStore.id }
      : data

    const result = createTopicSchema.safeParse(submitData)
    if (!result.success) {
      const error = JSON.parse(result.error.message)[0]
      useMessage(formatKunZodIssue(error), 'warn')
      return
    }

    const hasConsumeSection = TOPIC_SECTION_CONSUME_MOEMOEPOINTS.some((item) =>
      submitData.section.includes(item as 'g-seeking')
    )
    if (
      hasConsumeSection &&
      moemoepoint < MOEMOEPOINT_COST_FOR_CONSUME_SECTION
    ) {
      useMessage(
        `您没有足够的萌萌点来发布求助或者寻求资源的话题, 您可以通过发布 Galgame, 签到, 接受别人的赞赏, 等等来获取萌萌点`,
        'warn'
      )
      return
    }

    isSubmitting.value = true
    if (isRewriteMode.value) {
      const topicId = tempStore.id
      await kunFetch<string>(`/topic/${topicId}`, {
        method: 'PUT',
        body: submitData
      })
      useKunLoliInfo('重新编辑成功', 5)
      tempStore.resetRewriteTopicData()
      await navigateTo(`/topic/${topicId}`)
    } else {
      const tid = await kunFetch<number>('/topic', {
        method: 'POST',
        body: submitData
      })
      if (tid) {
        useKunLoliInfo('发布成功', 5)
        persistStore.resetTopicData()
        await navigateTo(`/topic/${tid}`)
      }
    }
    isSubmitting.value = false
  }

  return {
    rules,
    submit,
    isSubmitting,
    isRewriteMode
  }
}
