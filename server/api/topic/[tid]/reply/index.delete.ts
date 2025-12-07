import { prisma } from '~~/prisma/prisma'
import { deleteReplySchema } from '~/validations/topic'
import type { Prisma, PrismaClient } from '~~/prisma/generated/prisma/client'

const deleteReplyRecursive = async (
  replyIdToDelete: number,
  prisma: Omit<
    PrismaClient<Prisma.PrismaClientOptions, never, DefaultArgs>,
    '$connect' | '$disconnect' | '$on' | '$transaction' | '$use' | '$extends'
  >
) => {
  const allIdsToDelete = new Set<number>()

  let queue = [replyIdToDelete]

  while (queue.length > 0) {
    const directChildren = await prisma.topic_reply.findMany({
      where: {
        target: {
          some: {
            target_reply_id: {
              in: queue
            }
          }
        }
      },
      select: { id: true }
    })

    const newIds = directChildren.map((r) => r.id)

    queue = []

    for (const id of newIds) {
      if (!allIdsToDelete.has(id)) {
        allIdsToDelete.add(id)
        queue.push(id)
      }
    }
  }

  allIdsToDelete.add(replyIdToDelete)

  await prisma.topic_reply.deleteMany({
    where: {
      id: { in: Array.from(allIdsToDelete) }
    }
  })
}

export default defineEventHandler(async (event) => {
  const input = kunParseDeleteQuery(event, deleteReplySchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const reply = await prisma.topic_reply.findUnique({
    where: { id: input.replyId },
    include: {
      _count: {
        select: {
          comment: true,
          like: true,
          target: true,
          target_by: true
        }
      }
    }
  })
  if (!reply) {
    return kunError(event, '未找到这个回复')
  }

  const userInfo = await getCookieTokenInfo(event)
  if (!userInfo) {
    return kunError(event, '用户登录失效', 205)
  }
  if (reply.user_id !== userInfo.uid && userInfo.role < 2) {
    return kunError(event, '您无权删除这个回复')
  }

  const moemoepointToDecreaseIfUserDelete =
    3 *
    (reply._count.comment +
      reply._count.like +
      reply._count.target +
      reply._count.target_by +
      1)

  const moemoepointToDecrease =
    userInfo.role < 2 ? moemoepointToDecreaseIfUserDelete : 3

  const user = await prisma.user.findUnique({
    where: { id: userInfo.uid },
    select: { moemoepoint: true }
  })
  if (!user) {
    return kunError(event, '未找到用户')
  }
  if (
    user.moemoepoint < moemoepointToDecreaseIfUserDelete &&
    userInfo.role < 2
  ) {
    return kunError(
      event,
      `您的萌萌点不足, 删除这个回复需要 ${moemoepointToDecreaseIfUserDelete} 萌萌点`
    )
  }

  return prisma.$transaction(async (prisma) => {
    await deleteReplyRecursive(reply.id, prisma)

    await prisma.user.update({
      where: { id: reply.user_id },
      data: { moemoepoint: { increment: -moemoepointToDecrease } }
    })

    return 'MOEMOE delete reply successfully!'
  })
})
