import { TAG_MAP } from '~/lib/tagMap'
import type { PrismaClient } from '~~/prisma/generated/prisma/client'

const API_BASE_URL = 'https://api.vndb.org/kana'

const TAG_CATEGORY_MAP: Record<string, string> = {
  cont: 'content',
  ero: 'sexual',
  tech: 'technical'
}

const PRODUCER_TYPE_MAP: Record<string, string> = {
  co: 'company',
  in: 'individual',
  ng: 'amateur'
}

interface VndbExtLink {
  url: string
  label: string
}

interface VndbProducer {
  id: string
  name: string
  original: string | null
  aliases: string[]
  lang: string
  type: 'co' | 'in' | 'ng'
  extlinks: VndbExtLink[]
}

interface VndbTag {
  id: string
  name: string
  aliases: string[]
  description: string
  category: 'cont' | 'ero' | 'tech'
  spoiler: number
}

interface VnResult {
  id: string
  developers: VndbProducer[]
  tags: VndbTag[]
}

interface ReleaseResult {
  vns: { id: string }[]
  engine: string | null
}

const fetchVnData = async (vndbId: string): Promise<VnResult | null> => {
  const response = await fetch(`${API_BASE_URL}/vn`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      filters: ['id', '=', vndbId],
      fields:
        'id, developers{id,name,original,aliases,lang,type,extlinks{url,label}}, tags{id,name,aliases,category,spoiler,description}'
    })
  })
  if (!response.ok) {
    console.error(
      `[VNDB Sync] Failed to fetch VN data for ${vndbId}:`,
      await response.text()
    )
    return null
  }
  const data = await response.json()
  return data.results[0] || null
}

const fetchReleaseData = async (vndbId: string): Promise<ReleaseResult[]> => {
  const response = await fetch(`${API_BASE_URL}/release`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      filters: ['vn', '=', ['id', '=', vndbId]],
      fields: 'engine, vns{id}',
      results: 100
    })
  })
  if (!response.ok) {
    console.error(
      `[VNDB Sync] Failed to fetch Release data for ${vndbId}:`,
      await response.text()
    )
    return []
  }
  const data = await response.json()
  return data.results
}

type PrismaTransactionClient = Omit<
  PrismaClient,
  '$connect' | '$disconnect' | '$on' | '$transaction' | '$use' | '$extends'
>

export const syncVndbData = async (
  prisma: PrismaTransactionClient,
  {
    galgameId,
    vndbId,
    userId
  }: { galgameId: number; vndbId: string; userId: number }
) => {
  try {
    const [vnResult, releaseResults] = await Promise.all([
      fetchVnData(vndbId),
      fetchReleaseData(vndbId)
    ])

    if (!vnResult) {
      throw new Error('VNDB sync failed, vnResult not found')
    }

    const engineNames = [
      ...new Set(releaseResults.map((r) => r.engine).filter(Boolean))
    ]
    const engineRelations = []
    for (const name of engineNames) {
      if (name) {
        const engine = await prisma.galgame_engine.upsert({
          where: { name },
          update: {},
          create: { name, description: '' }
        })
        engineRelations.push({ galgame_id: galgameId, engine_id: engine.id })
      }
    }
    if (engineRelations.length > 0) {
      await prisma.galgame_engine_relation.createMany({
        data: engineRelations,
        skipDuplicates: true
      })
    }

    const officialRelations = []
    const otherLinks = []
    for (const dev of vnResult.developers) {
      const officialWebsite =
        dev.extlinks?.find((l) => l.label === 'Official website')?.url || ''
      if (dev.extlinks) {
        for (const link of dev.extlinks) {
          if (link.label !== 'Official website') {
            otherLinks.push({ name: link.label, link: link.url })
          }
        }
      }

      const official = await prisma.galgame_official.upsert({
        where: { name: dev.original || dev.name },
        update: { link: officialWebsite },
        create: {
          name: dev.original || dev.name,
          category: PRODUCER_TYPE_MAP[dev.type] || 'company',
          lang: dev.lang,
          link: officialWebsite
        }
      })

      officialRelations.push({
        galgame_id: galgameId,
        official_id: official.id
      })

      const aliases = (
        dev.original ? [dev.name, ...dev.aliases] : dev.aliases
      ).filter(Boolean)
      if (aliases.length > 0) {
        await prisma.galgame_official_alias.createMany({
          data: aliases.map((aliasName) => ({
            galgame_official_id: official.id,
            name: aliasName
          })),
          skipDuplicates: true
        })
      }
    }

    if (officialRelations.length > 0) {
      await prisma.galgame_official_relation.createMany({
        data: officialRelations,
        skipDuplicates: true
      })
    }

    if (otherLinks.length > 0) {
      await prisma.galgame_link.createMany({
        data: otherLinks.map((link) => ({
          ...link,
          galgame_id: galgameId,
          user_id: userId
        }))
      })
    }

    let hasSexualTag = false
    const tagRelations = []

    for (const tag of vnResult.tags) {
      if (!TAG_CATEGORY_MAP[tag.category]) continue

      if (tag.category === 'ero') {
        hasSexualTag = true
      }

      const mappedValue = TAG_MAP[tag.name] || tag.name
      const parts = mappedValue.split('/')
      const newName = parts[0]
      const additionalAliases = parts.slice(1)

      const dbTag = await prisma.galgame_tag.upsert({
        where: { name: newName },
        update: {},
        create: {
          name: newName,
          category: TAG_CATEGORY_MAP[tag.category],
          description: tag.description || ''
        }
      })

      tagRelations.push({
        galgame_id: galgameId,
        tag_id: dbTag.id,
        spoiler_level: tag.spoiler
      })

      const aliases = [
        ...new Set([tag.name, ...tag.aliases, ...additionalAliases])
      ].filter(Boolean)
      if (aliases.length > 0) {
        await prisma.galgame_tag_alias.createMany({
          data: aliases.map((aliasName) => ({
            galgame_tag_id: dbTag.id,
            name: aliasName
          })),
          skipDuplicates: true
        })
      }
    }
    if (tagRelations.length > 0) {
      await prisma.galgame_tag_relation.createMany({
        data: tagRelations,
        skipDuplicates: true
      })
    }

    if (hasSexualTag) {
      await prisma.galgame.update({
        where: { id: galgameId },
        data: { age_limit: 'r18' }
      })
    }
  } catch (error) {
    throw new Error('VNDB sync failed, transaction will be rolled back.')
  }
}

export const resyncVndbData = async (
  prisma: PrismaTransactionClient,
  {
    galgameId,
    newVndbId,
    userId
  }: { galgameId: number; newVndbId: string; userId: number }
) => {
  await prisma.galgame.update({
    where: { id: galgameId },
    data: { age_limit: 'all' }
  })

  await prisma.galgame_tag_relation.deleteMany({
    where: { galgame_id: galgameId }
  })
  await prisma.galgame_engine_relation.deleteMany({
    where: { galgame_id: galgameId }
  })
  await prisma.galgame_official_relation.deleteMany({
    where: { galgame_id: galgameId }
  })

  await prisma.galgame_link.deleteMany({
    where: {
      galgame_id: galgameId
    }
  })
  await prisma.galgame_link.create({
    data: {
      galgame_id: galgameId,
      user_id: userId,
      name: 'VNDB',
      link: `https://vndb.org/${newVndbId}`
    }
  })

  await syncVndbData(prisma, {
    galgameId: galgameId,
    vndbId: newVndbId,
    userId: userId
  })
}
