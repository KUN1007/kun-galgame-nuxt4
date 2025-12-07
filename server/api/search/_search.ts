import { prisma } from '~~/prisma/prisma'
import { markdownToText } from '~~/server/utils/markdownToText'
import type {
  SearchResultTopic,
  SearchResultGalgame,
  SearchResultUser,
  SearchResultReplyTarget,
  SearchResultReply,
  SearchResultComment
} from '~/types/api/search'

export const searchTopic = async (
  keywords: string[],
  skip: number,
  limit: number,
  isSFW?: boolean
) => {
  const data = await prisma.topic.findMany({
    where: {
      status: { not: 1 },
      ...(isSFW ? { is_nsfw: false } : {}),
      AND: keywords.map((kw) => ({
        OR: [
          { category: { in: keywords } },
          { tag: { hasSome: keywords } },
          { title: { contains: kw, mode: 'insensitive' } },
          { content: { contains: kw, mode: 'insensitive' } }
        ]
      }))
    },
    skip,
    take: limit,
    orderBy: { created: 'desc' },
    include: {
      best_answer: true,
      user: {
        select: {
          id: true,
          name: true,
          avatar: true
        }
      },
      _count: {
        select: {
          like: true,
          reply: true,
          comment: true,
          poll: true
        }
      },
      section: {
        select: {
          topic_section: {
            select: {
              name: true
            }
          }
        }
      }
    }
  })

  const topics: SearchResultTopic[] = data.map((topic) => ({
    id: topic.id,
    title: topic.title,
    view: topic.view,
    statusUpdateTime: topic.status_update_time,
    tag: topic.tag,
    section: topic.section.map((s) => s.topic_section.name),
    user: topic.user,
    status: topic.status,
    upvoteTime: topic.upvote_time,
    likeCount: topic._count.like,
    replyCount: topic._count.reply,
    commentCount: topic._count.comment,
    hasBestAnswer: !!topic.best_answer,
    isPollTopic: !!topic._count.poll,
    isNSFWTopic: topic.is_nsfw
  }))

  return topics
}

export const searchGalgame = async (
  keywords: string[],
  skip: number,
  limit: number
) => {
  const data = await prisma.galgame.findMany({
    skip,
    take: limit,
    orderBy: { created: 'desc' },
    where: {
      status: { not: 1 },
      AND: keywords.map((kw) => ({
        OR: [
          { vndb_id: { in: keywords } },
          {
            tag: {
              some: {
                tag: {
                  name: { contains: kw, mode: 'insensitive' }
                }
              }
            }
          },
          { name_en_us: { contains: kw, mode: 'insensitive' } },
          { name_ja_jp: { contains: kw, mode: 'insensitive' } },
          { name_zh_cn: { contains: kw, mode: 'insensitive' } },
          { name_zh_tw: { contains: kw, mode: 'insensitive' } },
          { intro_en_us: { contains: kw, mode: 'insensitive' } },
          { intro_ja_jp: { contains: kw, mode: 'insensitive' } },
          { intro_zh_cn: { contains: kw, mode: 'insensitive' } },
          { intro_zh_tw: { contains: kw, mode: 'insensitive' } },
          {
            alias: {
              some: {
                name: { contains: kw, mode: 'insensitive' }
              }
            }
          }
        ]
      }))
    },
    include: {
      user: {
        select: {
          id: true,
          name: true,
          avatar: true
        }
      },
      resource: {
        select: {
          platform: true,
          language: true
        }
      },
      _count: {
        select: {
          like: true
        }
      }
    }
  })

  const galgames: SearchResultGalgame[] = data.map((galgame) => {
    const platforms = [...new Set(galgame.resource.map((r) => r.platform))]
    const languages = [...new Set(galgame.resource.map((r) => r.language))]

    return {
      id: galgame.id,
      name: {
        'en-us': galgame.name_en_us,
        'ja-jp': galgame.name_ja_jp,
        'zh-cn': galgame.name_zh_cn,
        'zh-tw': galgame.name_zh_tw
      },
      banner: galgame.banner,
      user: galgame.user,
      contentLimit: galgame.content_limit,
      view: galgame.view,
      likeCount: galgame._count.like,
      resourceUpdateTime: galgame.resource_update_time,
      platform: platforms,
      language: languages
    }
  })

  return galgames
}

export const searchUser = async (name: string, skip: number, limit: number) => {
  const data = await prisma.user.findMany({
    skip,
    take: limit,
    orderBy: { created: 'desc' },
    where: {
      name: { contains: name, mode: 'insensitive' }
    }
  })

  const users: SearchResultUser[] = data.map((user) => ({
    id: user.id,
    name: user.name,
    avatar: user.avatar,
    bio: user.bio,
    moemoepoint: user.moemoepoint,
    created: user.created
  }))

  return users
}

export const searchReply = async (
  content: string,
  skip: number,
  limit: number
) => {
  const data = await prisma.topic_reply.findMany({
    skip,
    take: limit,
    orderBy: { created: 'desc' },
    where: {
      content: { contains: content, mode: 'insensitive' }
    },
    include: {
      user: {
        select: {
          id: true,
          name: true,
          avatar: true
        }
      },
      target: {
        orderBy: {
          target_reply: { floor: 'asc' }
        },
        select: {
          content: true,
          target_reply: {
            select: {
              id: true,
              floor: true,
              content: true,
              user: {
                select: { id: true, name: true, avatar: true }
              }
            }
          }
        }
      },
      topic: {
        select: {
          title: true
        }
      }
    }
  })

  const replies: SearchResultReply[] = data.map((reply) => {
    const targets: SearchResultReplyTarget[] = reply.target.map(
      (targetRelation) => {
        const targetReply = targetRelation.target_reply
        const content = markdownToText(targetRelation.content)
        const contentPreviewText = markdownToText(targetReply.content)
        return {
          id: targetReply.id,
          user: targetReply.user,
          content: content.slice(0, 150) + (content.length > 150 ? '...' : ''),
          contentPreview:
            contentPreviewText.slice(0, 150) +
            (contentPreviewText.length > 150 ? '...' : '')
        }
      }
    )

    return {
      topicId: reply.topic_id,
      topicTitle: reply.topic.title,
      content: markdownToText(reply.content).slice(0, 233),
      user: reply.user,
      floor: reply.floor,
      targets,
      created: reply.created
    }
  })

  return replies
}

export const searchComment = async (
  content: string,
  skip: number,
  limit: number
) => {
  const data = await prisma.topic_comment.findMany({
    skip,
    take: limit,
    orderBy: { created: 'desc' },
    where: {
      content: { contains: content, mode: 'insensitive' }
    },
    include: {
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
      },
      topic: {
        select: {
          title: true
        }
      }
    }
  })

  const comments: SearchResultComment[] = data.map((comment) => ({
    topicId: comment.topic_id,
    topicTitle: comment.topic.title,
    content: comment.content.slice(0, 233),
    user: comment.user,
    targetUser: comment.target_user,
    created: comment.created
  }))

  return comments
}
