import { prisma } from '~~/prisma/prisma'
import { createToolsetResourceSchema } from '~/validations/toolset'
import {
  getUploadCache,
  removeUploadCache
} from '~~/server/utils/upload/saveUploadSalt'
import { isValidArchive } from '~/utils/validate'
import type { ToolsetResource } from '~/types/api/toolset'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const input = await kunParsePostBody(event, createToolsetResourceSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const { toolsetId, content, code, password, size, note, salt } = input

  const fileCache = await getUploadCache(salt)
  if (fileCache && !isValidArchive(fileCache.key)) {
    return kunError(event, '非法的文件扩展名')
  }

  const toolset = await prisma.galgame_toolset.findUnique({
    where: { id: toolsetId },
    select: { id: true }
  })
  if (!toolset) {
    return kunError(event, '该工具资源不存在')
  }

  const user = await prisma.user.findUnique({
    where: { id: userInfo.uid },
    select: {
      id: true,
      name: true,
      avatar: true
    }
  })
  if (!user) {
    return '未找到该用户'
  }

  const newResource = await prisma.$transaction(async (p) => {
    const res = await p.galgame_toolset_resource.create({
      data: {
        content: salt && fileCache ? fileCache.key : content,
        type: salt ? 's3' : 'user',
        code,
        password,
        size: salt && fileCache ? fileCache.filesize.toString() : size,
        note,
        toolset_id: toolsetId,
        user_id: userInfo.uid
      }
    })

    await p.galgame_toolset.update({
      where: { id: toolset.id },
      data: { resource_update_time: new Date() }
    })

    await prisma.user.update({
      where: { id: userInfo.uid },
      data: { moemoepoint: { increment: 3 } }
    })

    await p.galgame_toolset_contributor.createMany({
      data: [{ toolset_id: toolsetId, user_id: userInfo.uid }],
      skipDuplicates: true
    })

    return res
  })

  await removeUploadCache(salt)

  return newResource satisfies ToolsetResource
})
