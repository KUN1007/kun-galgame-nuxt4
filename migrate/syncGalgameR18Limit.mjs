import { PrismaClient } from '../prisma/generated/prisma/client'

const prisma = new PrismaClient()

const syncGalgameR18Limit = async () => {
  try {
    await prisma.galgame.updateMany({
      data: { age_limit: 'all' }
    })

    const updateResult = await prisma.galgame.updateMany({
      where: {
        tag: {
          some: {
            tag: {
              category: 'sexual'
            }
          }
        }
      },

      data: {
        age_limit: 'r18'
      }
    })

    console.log(`✅ Script finished successfully.`)
    console.log(`Found and updated ${updateResult.count} galgames.`)
  } catch (error) {
    console.error('error:', error)
    process.exit(1)
  } finally {
    await prisma.$disconnect()
  }
}

syncGalgameR18Limit()
