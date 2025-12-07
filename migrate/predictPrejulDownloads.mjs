// predicted_preJul = round( β * time_est + (1 - β) * view_est )
// est = SCALE * (BETA * time_est + (1 - BETA) * view_est)

import { PrismaClient } from '../prisma/generated/prisma/client'
import process from 'process'

const prisma = new PrismaClient()

const JUL26_ISO = '2025-07-26T00:00:00.000Z'
const JUL26 = new Date(JUL26_ISO)

const BETA = 0.3
const SCALE = 2.0

function daysBetween(a, b) {
  return Math.max(0, (b.getTime() - a.getTime()) / (1000 * 60 * 60 * 24))
}

async function main() {
  const games = await prisma.galgame.findMany({
    select: {
      id: true,
      view: true,

      resource: {
        select: {
          id: true,
          created: true,
          download: true
        }
      }
    },
    where: {}
  })

  console.log(
    `Loaded ${games.length} games. Computing global conversion factor...`
  )

  let total_downloads_after = 0
  let total_views_for_games_with_after = 0
  const gamesWithAfterMap = new Map()

  for (const g of games) {
    let gameHasAfter = false
    for (const r of g.resource) {
      if (r.created >= JUL26) {
        total_downloads_after += r.download
        gameHasAfter = true
      }
    }
    if (gameHasAfter) {
      total_views_for_games_with_after += g.view || 0
      gamesWithAfterMap.set(g.id, true)
    }
  }

  const global_conv =
    total_views_for_games_with_after > 0
      ? total_downloads_after / total_views_for_games_with_after
      : 0.01
  console.log(
    `total_downloads_after (resources created >= Jul26): ${total_downloads_after}`
  )
  console.log(
    `total_views_for_games_with_after: ${total_views_for_games_with_after}`
  )
  console.log(
    `global conversion (downloads_per_view) used: ${global_conv.toFixed(6)}`
  )

  const updates = []

  for (const g of games) {
    const gameId = g.id
    const gameView = g.view || 0

    const totalDownloadsOfGame = g.resource.reduce(
      (s, r) => s + (r.download || 0),
      0
    )

    for (const r of g.resource) {
      if (!(r.created < JUL26)) continue

      const created = r.created
      const currentDownload = r.download || 0
      const resourceAgeDaysTotal = daysBetween(created, new Date())
      const resourceDaysPreJul = daysBetween(created, JUL26)

      const frac =
        resourceAgeDaysTotal > 0 ? resourceDaysPreJul / resourceAgeDaysTotal : 0
      const time_est = currentDownload * frac

      let view_est = 0
      if (totalDownloadsOfGame > 0) {
        const share = currentDownload / totalDownloadsOfGame
        view_est = gameView * share * global_conv
      } else {
        const preResourcesCount =
          g.resource.filter((rr) => rr.created < JUL26).length || 1
        view_est = (gameView * global_conv) / preResourcesCount
      }

      const est = SCALE * (BETA * time_est + (1 - BETA) * view_est)
      const predicted_preJul = Math.max(0, Math.round(est))

      if (predicted_preJul <= 0) continue

      updates.push({
        resourceId: r.id,
        add: predicted_preJul,
        before: currentDownload,
        after: currentDownload + predicted_preJul,
        gameId
      })
    }
  }

  try {
    await prisma.$transaction(async (tx) => {
      for (const u of updates) {
        await tx.galgame_resource.update({
          where: { id: u.resourceId },
          data: { download: { increment: u.add } }
        })
      }
    })
    console.log('Database updates applied successfully.')
  } catch (err) {
    console.error('Error during DB transaction:', err)
    console.error('No partial writes should have occurred; check logs and DB.')
  }
}

main().catch(async (e) => {
  console.error('Unhandled error:', e)
  try {
    await prisma.$disconnect()
  } finally {
    //
  }
  process.exit(1)
})
