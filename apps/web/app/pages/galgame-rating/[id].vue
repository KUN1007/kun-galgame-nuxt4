<script setup lang="ts">
import type {
  Review,
  WithContext,
  Person,
  Organization,
  VideoGame,
  Rating as SchemaRating,
  PropertyValue
} from 'schema-dts'

definePageMeta({ key: (route) => route.path })

const route = useRoute()
const id = computed(() => parseInt((route.params as { id: string }).id))

const { data, refresh } = await useKunFetch<GalgameRatingDetails>(
  `/galgame-rating/${id.value}`,
  {
    method: 'GET',
    watch: false,
    query: { galgame_rating_id: id.value }
  }
)

const jsonLd = computed<WithContext<Review> | null>(() => {
  if (!data.value) {
    return null
  }

  const rating = data.value
  const titleBase = rating.galgame.name
  const pageUrl = `${kungal.domain.main}${route.path}`
  const gameUrl = `${kungal.domain.main}/galgame/${rating.galgame.id}`

  const publisherSchema: Organization = {
    '@type': 'Organization',
    name: kungal.title,
    logo: {
      '@type': 'ImageObject',
      url: `${kungal.domain.main}/kungalgame.webp`
    }
  }

  const authorSchema: Person = {
    '@type': 'Person',
    name: rating.user.name,
    url: `${kungal.domain.main}/user/${rating.user.id}`
  }

  const itemReviewedSchema: VideoGame = {
    '@type': 'VideoGame',
    name: titleBase,
    url: gameUrl,
    image: getEffectiveBanner(rating.galgame),
    inLanguage: rating.galgame.original_language,
    isFamilyFriendly: rating.galgame.age_limit !== 'r18',
    ...(rating.galgame.official?.length && {
      publisher: rating.galgame.official.map((o) => ({
        '@type': 'Organization',
        name: o.name
      }))
    })
  }

  const reviewRatingSchema: SchemaRating = {
    '@type': 'Rating',
    ratingValue: rating.overall,
    bestRating: 10,
    worstRating: 1
  }

  const additionalProps: PropertyValue[] = [
    { name: '艺术风格', value: rating.art },
    { name: '故事情节', value: rating.story },
    { name: '音乐体验', value: rating.music },
    { name: '角色塑造', value: rating.character },
    { name: '路线设计', value: rating.route },
    { name: '系统交互', value: rating.system },
    { name: '声优演绎', value: rating.voice },
    { name: '重玩价值', value: rating.replay_value }
  ].map((p) => ({ '@type': 'PropertyValue', ...p }))

  return {
    '@context': 'https://schema.org',
    '@type': 'Review',
    mainEntityOfPage: pageUrl,
    headline: `${rating.user.name} 对 ${titleBase} 的评价`,
    datePublished: new Date(rating.created).toISOString(),
    dateModified: new Date(rating.updated).toISOString(),
    author: authorSchema,
    publisher: publisherSchema,
    itemReviewed: itemReviewedSchema,
    reviewRating: reviewRatingSchema,
    reviewBody: truncateRunes(markdownToText(rating.short_summary || ''), 250),
    interactionStatistic: [
      {
        '@type': 'InteractionCounter',
        interactionType: { '@type': 'LikeAction' },
        userInteractionCount: rating.like_count
      },
      {
        '@type': 'InteractionCounter',
        interactionType: { '@type': 'WatchAction' },
        userInteractionCount: rating.view
      }
    ],
    additionalProperty: additionalProps
  }
})

if (data.value) {
  if (data.value.galgame.content_limit === 'nsfw') {
    useKunDisableSeo(`${data.value.user.name} 的评价`)
  } else {
    useHead({
      script: [
        {
          id: 'schema-org-galgame-rating',
          type: 'application/ld+json',
          innerHTML: jsonLd.value
        }
      ]
    })

    const titleBase = data.value.galgame.name
    const original = data.value.galgame.name_original
    const title = original
      ? `${data.value.user.name} 对 ${titleBase} (${original}) 的评价`
      : `${data.value.user.name} 对 ${titleBase} 的评价`

    const description = truncateRunes(
      data.value.short_summary
        ? markdownToText(data.value.short_summary)
        : `${data.value.user.name} 对 ${titleBase} 的评分 ${data.value.overall}/10`,
      175
    )

    useKunSeoMeta({
      title,
      description,
      ogImage: getEffectiveBanner(data.value.galgame),
      articleAuthor: [`${kungal.domain.main}/user/${data.value.user.id}`],
      articlePublishedTime: data.value.created.toString(),
      articleModifiedTime: data.value.updated.toString()
    })
  }
} else {
  useKunDisableSeo('请求 Galgame 评分数据错误')
}
</script>

<template>
  <div>
    <GalgameRatingDetail
      v-if="data"
      :data="data"
      :rating-id="id"
      :refresh="refresh"
    />
    <KunNull v-else description="未找到该评分" />
  </div>
</template>
