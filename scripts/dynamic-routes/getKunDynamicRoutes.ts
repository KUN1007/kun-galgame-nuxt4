import { prisma } from '~~/prisma/prisma'

export const getKunDynamicRoutes = async () => {
  const [
    topics,
    galgames,
    engines,
    officials,
    series,
    tags,
    websites,
    websiteCategories,
    websiteTags,
    toolsets,
    galgameRatings
  ] = await Promise.all([
    prisma.topic.findMany({
      where: { status: { not: 1 }, is_nsfw: false },
      select: { id: true, updated: true }
    }),

    prisma.galgame.findMany({
      where: { content_limit: { not: 'nsfw' } },
      select: { id: true, updated: true }
    }),

    prisma.galgame_engine.findMany({
      select: { id: true, updated: true }
    }),

    prisma.galgame_official.findMany({
      select: { id: true, updated: true }
    }),

    prisma.galgame_series.findMany({
      select: { id: true, updated: true }
    }),

    prisma.galgame_tag.findMany({
      where: { category: { not: 'sexual' } },
      select: { id: true, updated: true }
    }),

    prisma.galgame_website.findMany({
      where: { age_limit: { not: 'r18' } },
      select: { url: true, updated: true }
    }),

    prisma.galgame_website_category.findMany({
      select: { name: true, updated: true }
    }),

    prisma.galgame_website_tag.findMany({
      select: { name: true, updated: true }
    }),

    prisma.galgame_toolset.findMany({
      select: { id: true, updated: true }
    }),

    prisma.galgame_rating.findMany({
      select: { id: true, updated: true }
    })
  ])

  const topicRoutes = topics.map((topic) => ({
    path: `/topic/${topic.id}`,
    lastmod: topic.updated?.toISOString() || new Date().toISOString()
  }))

  const galgameRoutes = galgames.map((galgame) => ({
    path: `/galgame/${galgame.id}`,
    lastmod: galgame.updated?.toISOString() || new Date().toISOString()
  }))

  const galgameEngineRoutes = engines.map((engine) => ({
    path: `/galgame-engine/${engine.id}`,
    lastmod: engine.updated?.toISOString() || new Date().toISOString()
  }))

  const galgameOfficialRoutes = officials.map((official) => ({
    path: `/galgame-official/${official.id}`,
    lastmod: official.updated?.toISOString() || new Date().toISOString()
  }))

  const galgameSeriesRoutes = series.map((item) => ({
    path: `/galgame-series/${item.id}`,
    lastmod: item.updated?.toISOString() || new Date().toISOString()
  }))

  const galgameTagRoutes = tags.map((tag) => ({
    path: `/galgame-tag/${tag.id}`,
    lastmod: tag.updated?.toISOString() || new Date().toISOString()
  }))

  const websiteRoutes = websites.map((website) => ({
    path: `/website/${encodeURIComponent(website.url)}`,
    lastmod: website.updated?.toISOString() || new Date().toISOString()
  }))

  const websiteCategoryRoutes = websiteCategories.map((category) => ({
    path: `/website-category/${category.name}`,
    lastmod: category.updated?.toISOString() || new Date().toISOString()
  }))

  const websiteTagRoutes = websiteTags.map((tag) => ({
    path: `/website-tag/${tag.name}`,
    lastmod: tag.updated?.toISOString() || new Date().toISOString()
  }))

  const galgameToolsetRoutes = toolsets.map((toolset) => ({
    path: `/toolset/${toolset.id}`,
    lastmod: toolset.updated?.toISOString() || new Date().toISOString()
  }))

  const galgameRatingRoutes = galgameRatings.map((rating) => ({
    path: `/galgame-rating/${rating.id}`,
    lastmod: rating.updated?.toISOString() || new Date().toISOString()
  }))

  return [
    ...topicRoutes,
    ...galgameRoutes,
    ...galgameEngineRoutes,
    ...galgameOfficialRoutes,
    ...galgameSeriesRoutes,
    ...galgameTagRoutes,
    ...websiteRoutes,
    ...websiteCategoryRoutes,
    ...websiteTagRoutes,
    ...galgameToolsetRoutes,
    ...galgameRatingRoutes
  ]
}
