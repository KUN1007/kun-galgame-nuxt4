import { prisma } from '~~/prisma/prisma'
import { deleteUserSchema } from '~/validations/user'
import { s3 } from '~/lib/s3/client'
import { DeleteObjectCommand } from '@aws-sdk/client-s3'

export default defineEventHandler(async (event) => {
  const input = kunParseDeleteQuery(event, deleteUserSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效')
  }
  if (userInfo.role <= 2) {
    return kunError(event, '您没有权限删除用户')
  }

  const targetUser = await prisma.user.findUnique({
    where: {
      id: input.userId
    },
    include: {
      galgame_toolset_resource: true
    }
  })
  if (!targetUser) {
    return kunError(event, '用户不存在')
  }
  if (targetUser.role > 1) {
    return kunError(event, '不能删除一个管理员')
  }

  for (const res of targetUser.galgame_toolset_resource) {
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

  await prisma.user.delete({
    where: { id: input.userId }
  })

  return 'Moemoe delete user successfully!'
})
