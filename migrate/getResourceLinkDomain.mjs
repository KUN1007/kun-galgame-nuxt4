import { PrismaClient } from '../prisma/generated/prisma/client'

const prisma = new PrismaClient()

const getResourceLinkDomain = async () => {
  const links = await prisma.galgame_resource_link.findMany({
    select: {
      url: true
    }
  })

  const hostnames = new Set()

  for (const link of links) {
    if (!link.url) {
      continue
    }

    try {
      const parsedUrl = new URL(link.url)

      hostnames.add(parsedUrl.hostname)
    } catch (error) {
      console.log(error)
    }
  }

  ;[...hostnames].forEach((hostname) => {
    console.log(hostname)
  })
}

getResourceLinkDomain()
  .catch(() => {
    process.exit(1)
  })
  .finally(async () => {
    await prisma.$disconnect()
  })
