import type {
  KunEditorAdapters,
  MentionUser,
  StickerPack
} from '@kungal/editor-core'
import { stickerArray } from '~/constants/sticker'

export const useKunEditorAdapters = (opts?: {
  image?: boolean
}): KunEditorAdapters => {
  const allowImage = opts?.image !== false

  const uploadImage = async (file: File) => {
    const form = new FormData()
    form.append('image', file)
    const url = await kunFetch<string>('/image/topic', {
      method: 'POST',
      body: form,
      watch: false
    })
    if (!url) {
      throw new Error('图片上传失败')
    }
    return url
  }

  const searchMentionUsers = async (query: string): Promise<MentionUser[]> =>
    (await kunFetch<MentionUser[]>('/user/search', {
      query: { q: query, limit: 8 }
    })) ?? []

  const stickerSource = (): StickerPack[] => [
    {
      name: 'KUNgal',
      stickers: stickerArray.map((src) => ({ src, name: src }))
    }
  ]

  const notify: KunEditorAdapters['notify'] = (message, level) => {
    useMessage(message, level)
  }

  return allowImage
    ? { uploadImage, searchMentionUsers, stickerSource, notify }
    : { searchMentionUsers, notify }
}
