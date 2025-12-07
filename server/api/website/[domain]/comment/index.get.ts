import { prisma } from '~~/prisma/prisma'
import { getCommentsSchema } from '~/validations/website'
import type { WebsiteComment } from '~/types/api/website'

const nestComments = (flatComments: WebsiteComment[]): WebsiteComment[] => {
  const commentMap: { [key: number]: WebsiteComment } = {}

  flatComments.forEach((comment) => {
    comment.reply = []
    commentMap[comment.id] = { ...comment, targetUser: null }
  })

  const nestedComments: WebsiteComment[] = []

  flatComments.forEach((comment) => {
    if (comment.parentId) {
      const parentComment = commentMap[comment.parentId]
      if (parentComment) {
        parentComment.reply.push(comment)
        comment.targetUser = commentMap[comment.parentId].user
      }
    } else {
      nestedComments.push(comment)
    }
  })

  return nestedComments
}

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getCommentsSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const data = await prisma.galgame_website_comment.findMany({
    where: { website_id: input.websiteId },
    orderBy: { created: 'desc' },
    include: {
      user: { select: { id: true, name: true, avatar: true } }
    }
  })

  const flatComments: WebsiteComment[] = data.map((comment) => ({
    id: comment.id,
    content: comment.content,
    parentId: comment.parent_id,
    userId: comment.user_id,
    websiteId: comment.website_id,
    created: comment.created,
    edited: comment.edited,
    reply: [],
    user: comment.user
  }))

  const nestedComments = nestComments(flatComments)

  return nestedComments
})
