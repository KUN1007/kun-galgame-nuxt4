import { prisma } from '~~/prisma/prisma'
import { getPreferredLanguageText } from '~/utils/getPreferredLanguageText'
import { markdownToText } from '~~/server/utils/markdownToText'
import type { ActivityItem, ActivityEventType } from '~/types/api/activity'

export type ActivityFetcher = (
  limit: number,
  skip: number,
  isSFW?: boolean
) => Promise<{
  items: ActivityItem[]
  total: number
}>

export const activityFetchers: Record<ActivityEventType, ActivityFetcher> = {
  GALGAME_CREATION: async (limit, skip, isSFW) => {
    const [items, total] = await prisma.$transaction([
      prisma.galgame.findMany({
        where: isSFW ? { content_limit: 'sfw' } : undefined,
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          name_en_us: true,
          name_ja_jp: true,
          name_zh_cn: true,
          name_zh_tw: true,
          created: true,
          user: { select: { id: true, name: true, avatar: true } }
        }
      }),
      prisma.galgame.count({
        where: isSFW ? { content_limit: 'sfw' } : undefined
      })
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `galgame-${item.id}`,
        type: 'GALGAME_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/galgame/${item.id}`,
        content: getPreferredLanguageText({
          'en-us': item.name_en_us,
          'ja-jp': item.name_ja_jp,
          'zh-cn': item.name_zh_cn,
          'zh-tw': item.name_zh_tw
        })
      })),
      total
    }
  },

  GALGAME_RATING_CREATION: async (limit, skip, isSFW) => {
    const where = isSFW ? { galgame: { content_limit: 'sfw' } } : undefined
    const [items, total] = await prisma.$transaction([
      prisma.galgame_rating.findMany({
        where,
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          short_summary: true,
          created: true,
          user: { select: { id: true, name: true, avatar: true } },
          galgame: {
            select: {
              id: true,
              name_en_us: true,
              name_ja_jp: true,
              name_zh_cn: true,
              name_zh_tw: true
            }
          }
        }
      }),
      prisma.galgame_rating.count({ where })
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `galgame-rating-${item.id}`,
        type: 'GALGAME_RATING_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/galgame-rating/${item.id}`,
        content: `${getPreferredLanguageText({
          'en-us': item.galgame.name_en_us,
          'ja-jp': item.galgame.name_ja_jp,
          'zh-cn': item.galgame.name_zh_cn,
          'zh-tw': item.galgame.name_zh_tw
        })} · ${item.short_summary.slice(0, 100)}`
      })),
      total
    }
  },

  GALGAME_COMMENT_CREATION: async (limit, skip, isSFW) => {
    const where = isSFW ? { galgame: { content_limit: 'sfw' } } : undefined
    const [items, total] = await prisma.$transaction([
      prisma.galgame_comment.findMany({
        where,
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          content: true,
          created: true,
          user: { select: { id: true, name: true, avatar: true } },
          galgame: { select: { id: true } }
        }
      }),
      prisma.galgame_comment.count({ where })
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `galgame-comment-${item.id}`,
        type: 'GALGAME_COMMENT_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/galgame/${item.galgame.id}`,
        content: item.content.slice(0, 100)
      })),
      total
    }
  },

  GALGAME_PR_CREATION: async (limit, skip) => {
    const [items, total] = await prisma.$transaction([
      prisma.galgame_pr.findMany({
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          created: true,
          new_data: true,
          user: { select: { id: true, name: true, avatar: true } },
          galgame: { select: { id: true } }
        }
      }),
      prisma.galgame_pr.count()
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `galgame-pr-${item.id}`,
        type: 'GALGAME_PR_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/galgame/${item.galgame.id}`,
        content: Object.values(item.new_data || {})
          .join('')
          .slice(0, 100)
      })),
      total
    }
  },

  GALGAME_WEBSITE_CREATION: async (limit, skip) => {
    const [items, total] = await prisma.$transaction([
      prisma.galgame_website.findMany({
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          user: { select: { id: true, name: true, avatar: true } },
          id: true,
          url: true,
          name: true,
          created: true
        }
      }),
      prisma.galgame_website.count()
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `galgame-website-${item.id}`,
        type: 'GALGAME_WEBSITE_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/website/${item.url}`,
        content: `新增 Galgame 站点: ${item.name}`
      })),
      total
    }
  },

  GALGAME_RATING_COMMENT_CREATION: async (limit, skip, isSFW) => {
    const where = isSFW
      ? { galgame_rating: { galgame: { content_limit: 'sfw' } } }
      : undefined
    const [items, total] = await prisma.$transaction([
      prisma.galgame_rating_comment.findMany({
        where,
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          content: true,
          created: true,
          user: { select: { id: true, name: true, avatar: true } },
          galgame_rating_id: true
        }
      }),
      prisma.galgame_rating_comment.count({ where })
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `galgame-rating-comment-${item.id}`,
        type: 'GALGAME_RATING_COMMENT_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/galgame-rating/${item.galgame_rating_id}`,
        content: item.content.substring(0, 100)
      })),
      total
    }
  },

  GALGAME_WEBSITE_COMMENT_CREATION: async (limit, skip) => {
    const [items, total] = await prisma.$transaction([
      prisma.galgame_website_comment.findMany({
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          content: true,
          created: true,
          user: { select: { id: true, name: true, avatar: true } },
          website: { select: { url: true } }
        }
      }),
      prisma.galgame_website_comment.count()
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `galgame-website-comment-${item.id}`,
        type: 'GALGAME_WEBSITE_COMMENT_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/website/${item.website.url}`,
        content: item.content.substring(0, 100)
      })),
      total
    }
  },

  GALGAME_RESOURCE_CREATION: async (limit, skip, isSFW) => {
    const where = isSFW ? { galgame: { content_limit: 'sfw' } } : undefined
    const [items, total] = await prisma.$transaction([
      prisma.galgame_resource.findMany({
        where,
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          type: true,
          created: true,
          user: { select: { id: true, name: true, avatar: true } },
          galgame: {
            select: {
              id: true,
              name_en_us: true,
              name_ja_jp: true,
              name_zh_cn: true,
              name_zh_tw: true
            }
          }
        }
      }),
      prisma.galgame_resource.count({ where })
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `galgame-resource-${item.id}`,
        type: 'GALGAME_RESOURCE_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/galgame/${item.galgame.id}`,
        content: `在《${getPreferredLanguageText({
          'en-us': item.galgame.name_en_us,
          'ja-jp': item.galgame.name_ja_jp,
          'zh-cn': item.galgame.name_zh_cn,
          'zh-tw': item.galgame.name_zh_tw
        })}》发布了下载资源`
      })),
      total
    }
  },

  TOOLSET_CREATION: async (limit, skip) => {
    const [items, total] = await prisma.$transaction([
      prisma.galgame_toolset.findMany({
        where: { status: { not: 1 } },
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          name: true,
          created: true,
          user: { select: { id: true, name: true, avatar: true } }
        }
      }),
      prisma.galgame_toolset.count({ where: { status: { not: 1 } } })
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `toolset-${item.id}`,
        type: 'TOOLSET_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/toolset/${item.id}`,
        content: item.name
      })),
      total
    }
  },

  TOOLSET_RESOURCE_CREATION: async (limit, skip) => {
    const [items, total] = await prisma.$transaction([
      prisma.galgame_toolset_resource.findMany({
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          created: true,
          user: { select: { id: true, name: true, avatar: true } },
          toolset: { select: { id: true, name: true } }
        }
      }),
      prisma.galgame_toolset_resource.count()
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `toolset-resource-${item.id}`,
        type: 'TOOLSET_RESOURCE_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/toolset/${item.toolset.id}`,
        content: item.toolset.name
      })),
      total
    }
  },

  TOOLSET_COMMENT_CREATION: async (limit, skip) => {
    const [items, total] = await prisma.$transaction([
      prisma.galgame_toolset_comment.findMany({
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          content: true,
          created: true,
          user: { select: { id: true, name: true, avatar: true } },
          toolset_id: true
        }
      }),
      prisma.galgame_toolset_comment.count()
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `toolset-comment-${item.id}`,
        type: 'TOOLSET_COMMENT_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/toolset/${item.toolset_id}`,
        content: item.content.substring(0, 100)
      })),
      total
    }
  },

  TOPIC_CREATION: async (limit, skip, isSFW) => {
    const [items, total] = await prisma.$transaction([
      prisma.topic.findMany({
        where: { status: { not: 1 }, ...(isSFW ? { is_nsfw: false } : {}) },
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          title: true,
          created: true,
          user: { select: { id: true, name: true, avatar: true } }
        }
      }),
      prisma.topic.count({
        where: { status: { not: 1 }, ...(isSFW ? { is_nsfw: false } : {}) }
      })
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `topic-${item.id}`,
        type: 'TOPIC_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/topic/${item.id}`,
        content: item.title
      })),
      total
    }
  },

  TOPIC_REPLY_CREATION: async (limit, skip, isSFW) => {
    const [items, total] = await prisma.$transaction([
      prisma.topic_reply.findMany({
        where: isSFW ? { topic: { is_nsfw: false } } : undefined,
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          topic_id: true,
          content: true,
          created: true,
          user: { select: { id: true, name: true, avatar: true } },
          target: { select: { content: true } }
        }
      }),
      prisma.topic_reply.count({
        where: isSFW ? { topic: { is_nsfw: false } } : undefined
      })
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `topic-reply-${item.id}`,
        type: 'TOPIC_REPLY_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/topic/${item.topic_id}`,
        content: `${markdownToText(item.content)}${item.target
          .map((t) => markdownToText(t.content))
          .join('')}`.slice(0, 100)
      })),
      total
    }
  },

  TOPIC_COMMENT_CREATION: async (limit, skip, isSFW) => {
    const where = isSFW ? { topic: { is_nsfw: false } } : undefined
    const [items, total] = await prisma.$transaction([
      prisma.topic_comment.findMany({
        where,
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          content: true,
          created: true,
          user: { select: { id: true, name: true, avatar: true } },
          topic_id: true
        }
      }),
      prisma.topic_comment.count({ where })
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `topic-comment-${item.id}`,
        type: 'TOPIC_COMMENT_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/topic/${item.topic_id}`,
        content: item.content.substring(0, 100)
      })),
      total
    }
  },

  TODO_CREATION: async (limit, skip) => {
    const [items, total] = await prisma.$transaction([
      prisma.todo.findMany({
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          created: true,
          content_en_us: true,
          content_ja_jp: true,
          content_zh_cn: true,
          content_zh_tw: true,
          user: { select: { id: true, name: true, avatar: true } }
        }
      }),
      prisma.todo.count()
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `todo-${item.id}`,
        type: 'TODO_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/update/todo`,
        content: getPreferredLanguageText({
          'en-us': item.content_en_us,
          'ja-jp': item.content_ja_jp,
          'zh-cn': item.content_zh_cn,
          'zh-tw': item.content_zh_tw
        }).substring(0, 100)
      })),
      total
    }
  },

  UPDATE_LOG_CREATION: async (limit, skip) => {
    const [items, total] = await prisma.$transaction([
      prisma.update_log.findMany({
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          user: { select: { id: true, name: true, avatar: true } },
          id: true,
          version: true,
          created: true,
          content_zh_cn: true
        }
      }),
      prisma.update_log.count()
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `update-log-${item.id}`,
        type: 'UPDATE_LOG_CREATION',
        timestamp: item.created,
        actor: item.user,
        link: `/update/history`,
        content: item.content_zh_cn
      })),
      total
    }
  },

  MESSAGE_SOLUTION: async (limit, skip) => {
    const [items, total] = await prisma.$transaction([
      prisma.message.findMany({
        where: { type: 'solution' },
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          type: true,
          content: true,
          link: true,
          created: true,
          sender: { select: { id: true, name: true, avatar: true } }
        }
      }),
      prisma.message.count({ where: { type: 'solution' } })
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `message-${item.id}`,
        type: 'MESSAGE_SOLUTION',
        timestamp: item.created,
        actor: item.sender,
        link: item.link,
        content: item.content.slice(0, 100)
      })),
      total
    }
  },

  MESSAGE_UPVOTE: async (limit, skip) => {
    const [items, total] = await prisma.$transaction([
      prisma.message.findMany({
        where: { type: 'upvoted' },
        orderBy: { created: 'desc' },
        take: limit,
        skip,
        select: {
          id: true,
          type: true,
          content: true,
          link: true,
          created: true,
          sender: { select: { id: true, name: true, avatar: true } }
        }
      }),
      prisma.message.count({ where: { type: 'upvoted' } })
    ])
    return {
      items: items.map((item) => ({
        uniqueId: `message-${item.id}`,
        type: 'MESSAGE_UPVOTE',
        timestamp: item.created,
        actor: item.sender,
        link: item.link,
        content: markdownToText(item.content).slice(0, 100)
      })),
      total
    }
  }
}
