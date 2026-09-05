import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { EditStorePersist } from '~/store/types/edit/topic'

export const usePersistEditTopicStore = defineStore(
  'KUNGalgameEditTopic',
  () => {
    const mode = ref<EditStorePersist['mode']>('preview')
    const title = ref<EditStorePersist['title']>('')
    const content = ref<EditStorePersist['content']>('')
    const category = ref<EditStorePersist['category']>('')
    const section = ref<EditStorePersist['section']>([])
    const isNSFW = ref<EditStorePersist['isNSFW']>(false)
    const coverImages = ref<EditStorePersist['coverImages']>([])
    const accessScope = ref<EditStorePersist['accessScope']>('public')
    const accessRoles = ref<EditStorePersist['accessRoles']>([])
    const accessUserIds = ref<EditStorePersist['accessUserIds']>([])

    const resetTopicData = () => {
      title.value = ''
      content.value = ''
      category.value = ''
      section.value = []
      isNSFW.value = false
      coverImages.value = []
      accessScope.value = 'public'
      accessRoles.value = []
      accessUserIds.value = []
    }

    return {
      mode,
      title,
      content,
      category,
      section,
      isNSFW,
      coverImages,
      accessScope,
      accessRoles,
      accessUserIds,
      resetTopicData
    }
  },
  {
    persist: {
      storage: piniaPluginPersistedstate.localStorage()
    }
  }
)
