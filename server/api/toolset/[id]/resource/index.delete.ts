import { prisma } from '~~/prisma/prisma'
import { deleteToolsetResourceSchema } from '~/validations/toolset'
import { s3 } from '~/lib/s3/client'
import { DeleteObjectCommand } from '@aws-sdk/client-s3'

export default defineEventHandler(async (event) => {
  const input = kunParseDeleteQuery(event, deleteToolsetResourceSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const toolsetResource = await prisma.galgame_toolset_resource.findUnique({
    where: { id: input.toolsetResourceId },
    select: { user_id: true, type: true, content: true }
  })
  if (!toolsetResource) {
    return kunError(event, '工具资源不存在或已删除')
  }
  if (toolsetResource.user_id !== userInfo.uid && userInfo.role < 2) {
    return kunError(event, '您无权限删除该工具资源')
  }

  if (toolsetResource.type === 's3') {
    await s3.send(
      new DeleteObjectCommand({
        Bucket: process.env.KUN_VISUAL_NOVEL_S3_STORAGE_BUCKET_NAME!,
        Key: toolsetResource.content
      })
    )
  }

  return prisma.$transaction(async (prisma) => {
    await prisma.user.update({
      where: { id: toolsetResource.user_id },
      data: { moemoepoint: { increment: -3 } }
    })

    await prisma.galgame_toolset_resource.delete({
      where: { id: input.toolsetResourceId }
    })

    return 'MOEMOE delete toolset successfully!'
  })
})
