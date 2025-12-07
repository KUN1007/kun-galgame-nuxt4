import { prisma } from '~~/prisma/prisma'
import {
  updateGalgameSchema,
  updateGalgamePrMergeSchema
} from '~/validations/galgame'
import { resyncVndbData } from '../../_syncVndb'
import type { z } from 'zod'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, updateGalgamePrMergeSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const galgamePR = await prisma.galgame_pr.findUnique({
    where: { id: input.galgamePrId },
    include: {
      galgame: {
        select: {
          user_id: true,
          vndb_id: true
        }
      }
    }
  })
  if (!galgamePR) {
    return kunError(event, '未找到这个更新请求')
  }
  if (galgamePR.status !== 0) {
    return kunError(event, '这个更新请求已经被拒绝或合并')
  }
  if (uid !== galgamePR.galgame.user_id && userInfo.role < 2) {
    return kunError(event, '您没有权限合并这个更新请求')
  }

  const prJSONObject = galgamePR.new_data as z.infer<typeof updateGalgameSchema>
  const parseResult = updateGalgameSchema.safeParse(prJSONObject)
  if (!parseResult.success) {
    return kunError(event, '解析 Galgame 更新请求内容失败')
  }

  const galgameId = galgamePR.galgame_id

  return prisma.$transaction(async (prisma) => {
    if (galgamePR.galgame.vndb_id !== prJSONObject.vndbId) {
      await resyncVndbData(prisma, {
        galgameId,
        newVndbId: prJSONObject.vndbId,
        userId: uid
      })
    }

    await prisma.galgame_pr.update({
      where: { id: input.galgamePrId },
      data: {
        status: 1,
        completed_time: new Date()
      }
    })

    await createGalgameHistory(prisma, {
      galgame_id: galgameId,
      user_id: uid,
      action: 'merged',
      type: 'pr',
      content: `#${galgamePR.index}`
    })

    const { vndbId, contentLimit, aliases, ...rest } = prJSONObject
    await prisma.galgame.update({
      where: { id: galgameId },
      data: {
        vndb_id: vndbId,
        content_limit: contentLimit,
        ...rest
      }
    })
    await prisma.galgame_alias.deleteMany({
      where: { galgame_id: galgameId }
    })

    await prisma.galgame_alias.createMany({
      data: aliases.split(',').map((a) => ({
        galgame_id: galgameId,
        name: a
      }))
    })

    await prisma.galgame_contributor.createMany({
      data: [
        { user_id: galgamePR.user_id, galgame_id: galgameId },
        { user_id: uid, galgame_id: galgameId }
      ],
      skipDuplicates: true
    })

    if (uid !== galgamePR.user_id) {
      await prisma.user.update({
        where: { id: galgamePR.user_id },
        data: { moemoepoint: { increment: 1 } }
      })

      await createMessage(
        prisma,
        uid,
        galgamePR.user_id,
        'merged',
        `#${galgamePR.index}`,
        undefined,
        galgamePR.galgame_id
      )
    }

    return 'MOEMOE merged galgame pull request successfully!'
  })
})
