import { prisma } from '~~/prisma/prisma'
import { updateToolsetResourceSchema } from '~/validations/toolset'
import type { ToolsetResource } from '~/types/api/toolset'

export default defineEventHandler(async (event) => {
  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const input = await kunParsePutBody(event, updateToolsetResourceSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const { toolsetResourceId, ...rest } = input

  const toolsetResource = await prisma.galgame_toolset_resource.findFirst({
    where: { id: toolsetResourceId }
  })
  if (!toolsetResource) {
    return kunError(event, '未找到该工具资源')
  }
  if (userInfo.role < 2 && userInfo.uid !== toolsetResource.user_id) {
    return kunError(event, '您没有权限更改该工具资源')
  }

  await prisma.$transaction(async (p) => {
    await p.galgame_toolset_resource.update({
      where: { id: toolsetResourceId },
      data:
        toolsetResource.type === 's3'
          ? {
              password: rest.password,
              note: rest.note,
              edited: new Date()
            }
          : { ...rest, edited: new Date() }
    })

    await p.galgame_toolset.update({
      where: { id: toolsetResource.toolset_id },
      data: { resource_update_time: new Date() }
    })

    await p.galgame_toolset_contributor.createMany({
      data: [{ toolset_id: toolsetResource.toolset_id, user_id: userInfo.uid }],
      skipDuplicates: true
    })
  })

  return toolsetResource satisfies ToolsetResource
})
