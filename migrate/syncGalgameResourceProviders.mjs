import { PrismaClient } from '../prisma/generated/prisma/client'
import process from 'process'

const PATTERNS = {
  baidu: ['pan.baidu.com', 'tieba.baidu.com', 'pan.baidu.', 'baidu.com'],
  aliyun: ['alipan.com', 'aliyun', 'aliyundrive', 'aliyuncs', 'aliyunpan'],
  quark: ['pan.quark.cn', 'quark.cn', 'quark'],
  pan123: [
    '123pan',
    '123684',
    '123865',
    '123912',
    '123912.com',
    '123684.com',
    '123865.com',
    '123pan.cn',
    'vip.123pan'
  ],
  tianyiyun: ['cloud.189.cn', '189.cn', 'ecloud.189.cn'],
  caiyun: ['caiyun.139.com', 'yun.139.com', '139.com'],
  xunlei: ['pan.xunlei.com', 'xunlei.com'],
  uc: ['drive.uc.cn', 'uc.cn'],
  lanzou: [
    'lanzou.com',
    'lanzous.com',
    'lanzoux.com',
    'lanzoui.com',
    'lanzouw.com',
    'lanzouj.com',
    'lanzouu.com',
    'lanzouq.com'
  ],
  other: []
}

function detectProvidersFromUrls(urls = []) {
  const found = []
  const pushOnce = (p) => {
    if (!found.includes(p)) found.push(p)
  }

  for (const raw of urls) {
    if (!raw || typeof raw !== 'string') continue
    const s = raw.toLowerCase()
    let matched = false
    for (const [provider, keys] of Object.entries(PATTERNS)) {
      for (const key of keys) {
        if (s.includes(key)) {
          pushOnce(provider)
          matched = true
          break
        }
      }
      if (matched) break
    }
    if (!matched) pushOnce('other')
  }
  return found
}

const prisma = new PrismaClient()

async function main() {
  const argv = process.argv.slice(2)
  const dry = argv.includes('--dry')
  const batchArgIndex = argv.findIndex((a) => a === '--batchSize')
  const batchSize =
    batchArgIndex >= 0 && argv[batchArgIndex + 1]
      ? parseInt(argv[batchArgIndex + 1], 10)
      : 200

  console.log(
    `Starting sync of galgame_resource.provider (dry=${dry}) batchSize=${batchSize}`
  )

  let skip = 0
  let totalUpdated = 0
  while (true) {
    const resources = await prisma.galgame_resource.findMany({
      skip,
      take: batchSize,
      include: { link: true },
      orderBy: { id: 'asc' }
    })
    if (!resources || resources.length === 0) break

    for (const res of resources) {
      const urls = (res.link || []).map((l) => l.url).filter(Boolean)
      const providers = detectProvidersFromUrls(urls)

      const existing = Array.isArray(res.provider)
        ? res.provider.map(String)
        : []
      const same =
        existing.length === providers.length &&
        existing.every((v, i) => v === providers[i])
      if (!same) {
        console.log(
          `resource id=${res.id} will update provider:`,
          existing,
          '=>',
          providers
        )
        if (!dry) {
          await prisma.galgame_resource.update({
            where: { id: res.id },
            data: { provider: providers }
          })
        }
        totalUpdated++
      }
    }

    skip += resources.length
    console.log(`Processed ${skip} resources...`)
    if (resources.length < batchSize) break
  }

  console.log(`Sync complete. totalUpdated=${totalUpdated}`)
}

main()
  .catch((e) => {
    console.error('Fatal error', e)
    process.exit(1)
  })
  .finally(async () => {
    await prisma.$disconnect()
  })
