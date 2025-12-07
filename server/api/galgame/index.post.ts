import { subDays } from 'date-fns'
import env from '~~/server/env/dotenv'
import { uploadGalgameBanner } from './utils/uploadGalgameBanner'
import { prisma } from '~~/prisma/prisma'
import { createGalgameSchema } from '~/validations/galgame'
import { syncVndbData } from './_syncVndb'
import type { H3Event } from 'h3'

const readGalgameData = async (event: H3Event) => {
  const input = await kunParseFormData(event, createGalgameSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const galgame = await prisma.galgame.count({
    where: { vndb_id: input.vndbId }
  })
  if (galgame) {
    return kunError(event, '未找到这个 Galgame')
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }

  const user = await prisma.user.findUnique({
    where: { id: userInfo.uid },
    include: {
      galgame: {
        where: {
          created: {
            gte: subDays(new Date(), 1)
          }
        }
      }
    }
  })
  if (!user) {
    return kunError(event, '未找到该用户')
  }

  if (user.moemoepoint / 10 < user.galgame.length) {
    return kunError(event, '您今日发布的 Galgame 已经达到上限')
  }

  return { result: input, uid: userInfo.uid }
}

export default defineEventHandler(async (event) => {
  const res = await readGalgameData(event)
  if (!res) {
    return
  }
  const { result, uid } = res

  return prisma.$transaction(
    async (prisma) => {
      const newGalgame = await prisma.galgame.create({
        data: {
          name_en_us: result.name_en_us,
          name_ja_jp: result.name_ja_jp,
          name_zh_cn: result.name_zh_cn,
          name_zh_tw: result.name_zh_tw,
          intro_en_us: result.intro_en_us,
          intro_ja_jp: result.intro_ja_jp,
          intro_zh_cn: result.intro_zh_cn,
          intro_zh_tw: result.intro_zh_tw,
          vndb_id: result.vndbId,
          content_limit: result.contentLimit,
          user_id: uid
        }
      })

      await syncVndbData(prisma, {
        galgameId: newGalgame.id,
        vndbId: result.vndbId,
        userId: uid
      })

      await prisma.galgame_alias.createMany({
        data: result.aliases.split(',').map((a) => ({
          galgame_id: newGalgame.id,
          name: a
        }))
      })

      await prisma.galgame_contributor.create({
        data: {
          galgame_id: newGalgame.id,
          user_id: uid
        }
      })

      await prisma.user.update({
        where: { id: uid },
        data: { moemoepoint: { increment: 3 } }
      })

      await uploadGalgameBanner(
        Buffer.from(result.banner as Buffer),
        newGalgame.id
      )

      const imageLink = `${env.KUN_VISUAL_NOVEL_IMAGE_BED_URL}/galgame/${newGalgame.id}/banner/banner.webp`
      await prisma.galgame.update({
        where: { id: newGalgame.id },
        data: { banner: imageLink }
      })

      await prisma.galgame_link.create({
        data: {
          galgame_id: newGalgame.id,
          user_id: uid,
          name: 'VNDB',
          link: `https://vndb.org/${result.vndbId}`
        }
      })

      await createGalgameHistory(prisma, {
        galgame_id: newGalgame.id,
        user_id: uid,
        action: 'created',
        type: 'galgame',
        content: ''
      })

      return newGalgame.id
    },
    { timeout: 60000 }
  )
})
