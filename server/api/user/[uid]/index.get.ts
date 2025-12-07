import { prisma } from '~~/prisma/prisma'
import { getUserInfoSchema } from '~/validations/user'
import { subDays } from 'date-fns'
import type { UserInfo } from '~/types/api/user'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getUserInfoSchema)
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
      galgame: {
        where: {
          created: {
            gte: subDays(new Date(), 1)
          }
        }
      },
      topic: {
        where: {
          created: {
            gte: subDays(new Date(), 1)
          }
        }
      },
      _count: {
        select: {
          reply_created: true,
          comment_created: true,
          topic: true,
          created_topic_poll: true,
          galgame: true,
          galgame_comment: true,
          galgame_pr: true,
          galgame_link: true,
          galgame_rating: true,
          galgame_contributor: true,
          galgame_resource: true,
          galgame_toolset: true,
          galgame_toolset_resource: true
        }
      }
    }
  })
  // not use a userPromise, because some users were not fount
  if (!user) {
    return kunError(event, '未找到该用户')
  }
  if (user.status === 1) {
    return 'banned'
  }

  // Likes
  const receivedLikesPromise = Promise.all([
    prisma.topic_like.count({ where: { topic: { user_id: uid } } }),
    prisma.topic_reply_like.count({ where: { reply: { user_id: uid } } }),
    prisma.topic_comment_like.count({ where: { comment: { user_id: uid } } }),
    prisma.galgame_like.count({ where: { galgame: { user_id: uid } } }),
    prisma.galgame_comment_like.count({
      where: { galgame_comment: { user_id: uid } }
    }),
    prisma.galgame_resource_like.count({
      where: { galgame_resource: { user_id: uid } }
    })
  ]).then((counts) => counts.reduce((acc, count) => acc + count, 0))

  // Dislikes
  const receivedDislikesPromise = Promise.all([
    prisma.topic_dislike.count({ where: { topic: { user_id: uid } } }),
    prisma.topic_reply_dislike.count({ where: { reply: { user_id: uid } } })
  ]).then((counts) => counts.reduce((acc, count) => acc + count, 0))

  // Upvotes
  const receivedUpvotesPromise = prisma.topic_upvote.count({
    where: { topic: { user_id: uid } }
  })

  // Favorites
  const receivedFavoritesPromise = Promise.all([
    prisma.topic_favorite.count({ where: { topic: { user_id: uid } } }),
    prisma.galgame_favorite.count({ where: { galgame: { user_id: uid } } })
  ]).then((counts) => counts.reduce((acc, count) => acc + count, 0))

  const [totalLikes, totalDislikes, totalUpvotes, totalFavorites] =
    await Promise.all([
      receivedLikesPromise,
      receivedDislikesPromise,
      receivedUpvotesPromise,
      receivedFavoritesPromise
    ])

  const responseData: UserInfo = {
    id: user.id,
    name: user.name,
    avatar: user.avatar,
    role: user.role,
    status: user.status,
    moemoepoint: user.moemoepoint,
    bio: user.bio,
    created: user.created,

    upvote: totalUpvotes,
    like: totalLikes,
    dislike: totalDislikes,
    favorite: totalFavorites,

    replyCreated: user._count.reply_created,
    commentCreated: user._count.comment_created,
    topic: user._count.topic,
    topicPoll: user._count.created_topic_poll,

    galgame: user._count.galgame,
    galgameComment: user._count.galgame_comment,
    galgamePr: user._count.galgame_pr,
    galgameLink: user._count.galgame_link,
    galgameRating: user._count.galgame_rating,
    contributeGalgame: user._count.galgame_contributor,

    galgameResource: user._count.galgame_resource,
    galgameToolset: user._count.galgame_toolset,
    galgameToolsetResource: user._count.galgame_toolset_resource,

    dailyTopicCount: user.topic.length,
    dailyGalgameCount: user.galgame.length
  }

  return responseData
})
