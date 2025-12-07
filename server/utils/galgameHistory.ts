import type { Prisma, PrismaClient } from '~~/prisma/generated/prisma/client'
import type { DefaultArgs } from '@prisma/client/runtime/library'
import type { GalgameHistoryCreateRequestData } from '~/types/api/galgame-history'

export const createGalgameHistory = async (
  prisma: Omit<
    PrismaClient<Prisma.PrismaClientOptions, never, DefaultArgs>,
    '$connect' | '$disconnect' | '$on' | '$transaction' | '$use' | '$extends'
  >,
  history: GalgameHistoryCreateRequestData
) => {
  await prisma.galgame_history.create({
    data: history
  })
}
