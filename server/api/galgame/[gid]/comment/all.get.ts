import { prisma } from '~~/prisma/prisma'
import { getGalgameCommentSchema } from '~/validations/galgame'
import type { GalgameComment } from '~/types/api/galgame-comment'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameCommentSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const userInfo = await getCookieTokenInfo(event)
  const { galgameId, page, limit, sortOrder } = input

  const skip = (page - 1) * limit

  const [data, totalCount] = await Promise.all([
    prisma.galgame_comment.findMany({
      take: limit,
      skip,
      where: { galgame_id: galgameId },
      orderBy: { created: sortOrder },
      include: {
        like: {
          where: {
            user_id: userInfo?.uid
          }
        },
        _count: {
          select: { like: true }
        },
        user: {
          select: {
            id: true,
            name: true,
            avatar: true
          }
        },
        target_user: {
          select: {
            id: true,
            name: true,
            avatar: true
          }
        }
      }
    }),
    prisma.galgame_comment.count({
      where: { galgame_id: galgameId }
    })
  ])

  const commentData: GalgameComment[] = data.map((comment) => ({
    id: comment.id,
    galgameId: comment.galgame_id,
    content: comment.content,
    likeCount: comment._count.like,
    isLiked: comment.like.length > 0,
    user: comment.user,
    targetUser: comment.target_user,
    created: new Date(comment.created).getTime()
  }))

  return { commentData, totalCount }
})
