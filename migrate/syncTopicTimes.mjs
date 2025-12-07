import { PrismaClient } from '../prisma/generated/prisma/client'

const prisma = new PrismaClient()

async function main() {
  console.log('Starting script to sync topic timestamps...')

  const updatedCount = await prisma.$executeRaw`
    UPDATE "topic"
    SET "status_update_time" = "edited"
    WHERE "status_update_time" < "edited"
  `

  console.log(`${updatedCount} topics were updated.`)
}

main()
  .catch((e) => {
    console.error(e)
    process.exit(1)
  })
  .finally(async () => {
    await prisma.$disconnect()
  })
