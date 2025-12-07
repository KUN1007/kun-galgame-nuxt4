import { prisma } from '../prisma/prisma'
import { TAG_MAP } from '../lib/tagMap'

const VNDB_API_BASE = 'https://api.vndb.org/kana'
const BATCH_SIZE = 100
const REQUEST_DELAY = 1000

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

async function fetchAllVndbTags() {
  console.log('  - 正在从 VNDB 获取所有标签数据...')
  const vndbTagIdToNameMap = new Map()
  let page = 1
  let hasMore = true

  while (hasMore) {
    const response = await fetch(`${VNDB_API_BASE}/tag`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        fields: 'id, name',
        results: 100,
        page: page
      })
    })

    if (!response.ok) {
      throw new Error(
        `获取 VNDB 标签失败，页码: ${page}, 状态码: ${response.status}`
      )
    }

    const data = await response.json()
    data.results.forEach((tag: { id: number; name: string }) => {
      vndbTagIdToNameMap.set(tag.id, tag.name)
    })

    hasMore = data.more
    if (hasMore) {
      process.stdout.write(`\r  - 已获取 ${vndbTagIdToNameMap.size} 个标签...`)
      page++
      await sleep(500)
    }
  }
  console.log(`\n  - ✅ 从 VNDB 成功获取了 ${vndbTagIdToNameMap.size} 个标签。`)
  return vndbTagIdToNameMap
}

async function main() {
  console.log('🚀 开始更新 Galgame 标签的剧透等级...')

  console.log('Step 1: 准备映射关系...')

  const vndbTagIdToNameMap = await fetchAllVndbTags()

  console.log('  - 正在从本地数据库加载 Galgame 和 Tag 数据...')
  const localGalgames = await prisma.galgame.findMany({
    select: { id: true, vndb_id: true }
  })

  const localTags = await prisma.galgame_tag.findMany({
    select: { id: true, name: true }
  })

  if (!localGalgames.length) {
    console.log('数据库中没有找到 Galgame，脚本结束。')
    return
  }

  const vndbIdToGameIdMap = new Map(localGalgames.map((g) => [g.vndb_id, g.id]))
  const chineseNameToTagIdMap = new Map(localTags.map((t) => [t.name, t.id]))
  console.log(
    `  - 本地数据库：找到 ${localGalgames.length} 个 Galgame 和 ${localTags.length} 个 Tag。`
  )

  console.log('  - 正在构建 VNDB ID 到本地 Tag ID 的主映射...')
  const vndbIdToLocalTagIdMap = new Map()
  let mappedCount = 0
  for (const [vndbId, englishName] of vndbTagIdToNameMap.entries()) {
    const chineseName = TAG_MAP[englishName]
    if (chineseName) {
      const localTagId = chineseNameToTagIdMap.get(chineseName)
      if (localTagId) {
        vndbIdToLocalTagIdMap.set(vndbId, localTagId)
        mappedCount++
      }
    }
  }
  console.log(`  - ✅ 主映射构建完成，成功关联 ${mappedCount} 个标签。`)

  console.log('\nStep 2: 开始处理游戏并更新剧透等级...')
  let updatedRelationCount = 0

  for (let i = 0; i < localGalgames.length; i += BATCH_SIZE) {
    const batch = localGalgames.slice(i, i + BATCH_SIZE)
    const batchVndbIds = batch.map((g) => g.vndb_id)

    console.log(
      `\n🔄 处理批次 ${Math.floor(i / BATCH_SIZE) + 1} / ${Math.ceil(localGalgames.length / BATCH_SIZE)} (游戏: ${batchVndbIds.join(', ')})`
    )

    try {
      const body = {
        filters: ['or', ...batchVndbIds.map((id) => ['id', '=', id])],
        fields: 'tags{id, spoiler}',
        results: batch.length
      }

      const response = await fetch(`${VNDB_API_BASE}/vn`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      })

      if (!response.ok) {
        throw new Error(
          `VNDB API 请求失败，状态码: ${response.status} ${response.statusText}`
        )
      }

      const data = await response.json()
      const vns = data.results
      const updatePromises = []

      for (const vn of vns) {
        const localGameId = vndbIdToGameIdMap.get(vn.id)
        if (!localGameId) continue

        for (const tag of vn.tags) {
          const localTagId = vndbIdToLocalTagIdMap.get(tag.id)

          if (localGameId && localTagId && tag.spoiler > 0) {
            updatePromises.push(
              prisma.galgame_tag_relation
                .update({
                  where: {
                    galgame_id_tag_id: {
                      galgame_id: localGameId,
                      tag_id: localTagId
                    }
                  },
                  data: { spoiler_level: tag.spoiler }
                })
                .catch((e) => {
                  if (e.code !== 'P2025') {
                    console.warn(
                      `  - 更新 [gameId: ${localGameId}, tagId: ${localTagId}] 失败: ${e.message}`
                    )
                  }
                })
            )
          }
        }
      }

      if (updatePromises.length > 0) {
        await Promise.all(updatePromises)
        console.log(
          `  - ✅ 成功更新了 ${updatePromises.length} 条关联记录的剧透等级！`
        )
        updatedRelationCount += updatePromises.length
      } else {
        console.log('  - 该批次没有需要更新的剧透标签。')
      }
    } catch (error) {
      console.error(
        `  - 处理批次时发生错误: ${(error as { message: string }).message}`
      )
    }

    if (i + BATCH_SIZE < localGalgames.length) {
      await sleep(REQUEST_DELAY)
    }
  }

  console.log(
    `\n🎉 全部完成！总共更新了 ${updatedRelationCount} 条 galgame_tag_relation 记录。`
  )
}

main()
  .catch((e) => {
    console.error('脚本执行时发生未捕获的错误:', e)
    process.exit(1)
  })
  .finally(async () => {
    await prisma.$disconnect()
  })
