import { prisma } from '~~/prisma/prisma'
import { getToolsetResourceDetailSchema } from '~/validations/toolset'
import type { ToolsetResourceDetail } from '~/types/api/toolset'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getToolsetResourceDetailSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const res = await prisma.$transaction(async (prisma) => {
    const res = await prisma.galgame_toolset_resource.findUnique({
      where: { id: input.toolsetResourceId },
      include: {
        user: {
          select: {
            id: true,
            name: true,
            avatar: true
          }
        }
      }
    })

    await prisma.galgame_toolset_resource.update({
      where: { id: input.toolsetResourceId },
      data: { download: { increment: 1 } }
    })

    return res
  })
  if (!res) {
    return kunError(event, '未找到该工具资源')
  }

  return res satisfies ToolsetResourceDetail
})
