import { prisma } from '~~/prisma/prisma'
import { deleteToolsetDetailSchema } from '~/validations/toolset'
import { s3 } from '~/lib/s3/client'
import { DeleteObjectCommand } from '@aws-sdk/client-s3'

export default defineEventHandler(async (event) => {
  const input = kunParseDeleteQuery(event, deleteToolsetDetailSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const toolset = await prisma.galgame_toolset.findUnique({
    where: { id: input.toolsetId },
    include: {
      user: {
        select: {
          moemoepoint: true
        }
      },
      resource: {
        select: {
          id: true,
          type: true,
          content: true
        }
      }
    }
  })
  if (!toolset) {
    return kunError(event, '工具不存在或已删除')
  }
  if (toolset.user_id !== userInfo.uid && userInfo.role < 2) {
    return kunError(event, '您无权限删除该工具')
  }

  const totalMoemoepointConsume = 3 + toolset.resource.length * 3
  if (totalMoemoepointConsume > toolset.user.moemoepoint) {
    return kunError(
      event,
      `您没有足够的萌萌点删除该工具, 删除需要 ${totalMoemoepointConsume} 萌萌点`
    )
  }

  return prisma.$transaction(
    async (prisma) => {
      for (const res of toolset.resource) {
        if (res.type === 's3') {
          await s3.send(
            new DeleteObjectCommand({
              Bucket: process.env.KUN_VISUAL_NOVEL_S3_STORAGE_BUCKET_NAME!,
              Key: res.content
            })
          )

          await prisma.galgame_toolset_resource.delete({
            where: { id: res.id }
          })
        }
      }

      await prisma.user.update({
        where: { id: toolset.user_id },
        data: { moemoepoint: { increment: -3 } }
      })

      await prisma.galgame_toolset.delete({ where: { id: input.toolsetId } })

      return 'MOEMOE delete toolset successfully!'
    },
    { timeout: 60000 }
  )
})
