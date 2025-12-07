import { prisma } from '~~/prisma/prisma'
import { getUserFloatingCardSchema } from '~/validations/user'
import type { UserFloatingCard } from '~/types/api/user'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getUserFloatingCardSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }
  const uid = input.userId

  const user = await prisma.user.findUnique({
    where: { id: uid },
    select: {
      id: true,
      name: true,
      avatar: true,
      role: true,
      status: true,
      moemoepoint: true,
      bio: true,
      created: true,
      _count: {
        select: {
          reply_created: true,
          comment_created: true,
          topic: true,
          galgame: true,
          galgame_comment: true,
          galgame_website_comment: true,
          galgame_contributor: true,
          galgame_resource: true
        }
      }
    }
  })
  if (!user) {
    return kunError(event, '未找到该用户')
  }
  if (user.status === 1) {
    return 'banned'
  }

  const responseData: UserFloatingCard = {
    id: user.id,
    name: user.name,
    avatar: user.avatar,
    moemoepoint: user.moemoepoint,
    topicCount: user._count.topic,
    topicReplyCount: user._count.reply_created,
    topicCommentCount:
      user._count.comment_created +
      user._count.galgame_comment +
      user._count.galgame_website_comment,
    galgameCount: user._count.galgame,
    galgameResourceCount: user._count.galgame_resource,
    galgameContributeCount: user._count.galgame_contributor
  }

  return responseData
})
