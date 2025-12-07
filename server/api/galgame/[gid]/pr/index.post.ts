import { prisma } from '~~/prisma/prisma'
import { updateGalgameSchema } from '~/validations/galgame'
import { formatDate } from '~/utils/time'
import { resyncVndbData } from '../../_syncVndb'

export default defineEventHandler(async (event) => {
  const input = await kunParsePostBody(event, updateGalgameSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  const uid = userInfo.uid

  const gid = getRouterParam(event, 'gid')
  if (!gid) {
    return kunError(event, '读取 Galgame ID 失败')
  }
  const galgameId = Number(gid)

  if (input.vndbId) {
    const duplicateGalgame = await prisma.galgame.findFirst({
      where: {
        vndb_id: input.vndbId,
        id: { not: galgameId }
      }
    })
    if (duplicateGalgame) {
      return kunError(event, '新 VNDB ID 的 Galgame 已经有人发布过了')
    }
  }

  const originalGalgame = await prisma.galgame.findUnique({
    where: { id: galgameId },
    include: {
      alias: {
        select: {
          name: true
        }
      }
    }
  })
  if (!originalGalgame) {
    return kunError(event, '未找到原 Galgame')
  }

  const originGalgameObject = {
    vndbId: originalGalgame.vndb_id,
    name_en_us: originalGalgame.name_en_us,
    name_ja_jp: originalGalgame.name_ja_jp,
    name_zh_cn: originalGalgame.name_zh_cn,
    name_zh_tw: originalGalgame.name_zh_tw,
    intro_en_us: originalGalgame.intro_en_us,
    intro_ja_jp: originalGalgame.intro_ja_jp,
    intro_zh_cn: originalGalgame.intro_zh_cn,
    intro_zh_tw: originalGalgame.intro_zh_tw,
    contentLimit: originalGalgame.content_limit,
    aliases: originalGalgame.alias.map((a) => a.name).toString()
  }

  return prisma.$transaction(async (prisma) => {
    if (uid === originalGalgame.user_id || userInfo.role > 1) {
      if (originalGalgame.vndb_id !== input.vndbId) {
        await resyncVndbData(prisma, {
          galgameId,
          newVndbId: input.vndbId,
          userId: uid
        })
      }

      const { vndbId, contentLimit, aliases, ...rest } = input
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

      await createGalgameHistory(prisma, {
        galgame_id: galgameId,
        user_id: uid,
        action: 'updated',
        type: 'galgame',
        content: formatDate(new Date(), { isPrecise: true, isShowYear: true })
      })

      await prisma.galgame_contributor.createMany({
        data: [{ user_id: uid, galgame_id: galgameId }],
        skipDuplicates: true
      })
    } else {
      const maxIndexPR = await prisma.galgame_pr.findFirst({
        orderBy: { id: 'desc' },
        take: 1,
        where: { galgame_id: galgameId }
      })
      const baseIndex = maxIndexPR ? maxIndexPR.index : 0
      const index = baseIndex + 1
      await prisma.galgame_pr.create({
        data: {
          galgame_id: galgameId,
          user_id: uid,
          index,
          old_data: originGalgameObject,
          new_data: input
        }
      })

      await createGalgameHistory(prisma, {
        galgame_id: galgameId,
        user_id: userInfo.uid,
        action: 'created',
        type: 'pr',
        content: ''
      })

      await createMessage(
        prisma,
        uid,
        originalGalgame.user_id,
        'requested',
        Object.values(input).join(',').slice(0, 233),
        undefined,
        galgameId
      )
    }

    return 'MOEMOE committed galgame pull request successfully!'
  })
})
