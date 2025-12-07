import { prisma } from '~~/prisma/prisma'
import {
  KUN_ADMIN_OVERVIEW_STATS_MODEL_ITEM,
  KUN_ADMIN_OVERVIEW_STATS_MODEL_MAP
} from '~/constants/admin'
import type { PrismaPromise } from '~~/prisma/generated/prisma/internal/prismaNamespace'

type ModelCount = {
  count: () => PrismaPromise<number>
}

export default defineEventHandler(async (event) => {
  const countPromises = KUN_ADMIN_OVERVIEW_STATS_MODEL_ITEM.map((modelName) => {
    const model = prisma[modelName] as ModelCount
    return model.count()
  })

  const counts = await Promise.all(countPromises)

  const result = KUN_ADMIN_OVERVIEW_STATS_MODEL_ITEM.map((modelName, index) => {
    const modelInfo = KUN_ADMIN_OVERVIEW_STATS_MODEL_MAP[modelName]
    return {
      name: modelName,
      label: modelInfo.label,
      color: modelInfo.color,
      count: counts[index]
    }
  })

  return result
})
