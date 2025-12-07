import { prisma } from '~~/prisma/prisma'
import { getGalgameLinkSchema } from '~/validations/galgame'
import type { GalgameLink } from '~/types/api/galgame-link'

export default defineEventHandler(async (event) => {
  const input = kunParseGetQuery(event, getGalgameLinkSchema)
  if (typeof input === 'string') {
    return kunError(event, input)
  }

  const data = await prisma.galgame_link.findMany({
    where: { galgame_id: input.galgameId },
    include: {
      user: {
        select: {
          id: true,
          name: true,
          avatar: true
        }
      }
    }
  })

  const links: GalgameLink[] = data.map((link) => ({
    id: link.id,
    galgameId: link.galgame_id,
    user: link.user,
    name: link.name,
    link: link.link
  }))

  return links
})
