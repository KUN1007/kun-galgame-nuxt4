<script setup lang="ts">
import type {
  Article,
  WebSite,
  Organization,
  Review,
  Person,
  WithContext
} from 'schema-dts'

definePageMeta({ key: (route) => route.path })

const route = useRoute()

const domain = computed(() => {
  return (route.params as { domain: string }).domain
})

const { data, refresh } = await useKunFetch<WebsiteDetail>(
  `/website/${domain.value}`,
  {
    watch: false,
    query: { domain: domain.value }
  }
)

const jsonLd = computed<WithContext<Article> | null>(() => {
  if (!data.value) {
    return null
  }

  const website = data.value
  const pageUrl = `${kungal.domain.main}/website/${website.url}`

  const publisherSchema: Organization = {
    '@type': 'Organization',
    name: kungal.title,
    logo: {
      '@type': 'ImageObject',
      url: `${kungal.domain.main}/kungalgame.webp`
    }
  }

  const aboutWebsiteSchema: WebSite = {
    '@type': 'WebSite',
    name: website.name,
    url: website.url,
    description: website.description,
    inLanguage: website.language,
    isFamilyFriendly: website.age_limit !== 'r18',
    image: website.icon_url
  }

  const reviewsSchema: Review[] = website.comment.map((comment) => {
    return {
      '@type': 'Review',
      author: {
        '@type': 'Person',
        name: comment.user.name,
        url: `${kungal.domain.main}/user/${comment.user.id}`
      } as Person,
      datePublished: new Date(comment.created).toISOString(),
      reviewBody: comment.content,
      publisher: publisherSchema
    }
  })

  const articleSchema: Article = {
    '@type': 'Article',
    mainEntityOfPage: pageUrl,
    headline: `关于 ${website.name} 的介绍与评价`,
    description: website.description,
    image: website.icon_url,
    datePublished: new Date(website.created).toISOString(),
    dateModified: new Date(website.updated).toISOString(),
    author: publisherSchema,
    publisher: publisherSchema,
    about: aboutWebsiteSchema,
    keywords: [
      website.category.label,
      ...website.tags.map((t) => t.label)
    ].join(', '),
    ...(reviewsSchema.length > 0 && { review: reviewsSchema })
  }

  return {
    '@context': 'https://schema.org',
    ...articleSchema
  }
})

if (data.value) {
  if (data.value.age_limit === 'all') {
    useHead({
      script: [
        {
          id: 'schema-org-website',
          type: 'application/ld+json',
          innerHTML: jsonLd.value
        }
      ]
    })

    useKunSeoMeta({
      title: data.value.name,
      description: data.value.description,
      ogImage: data.value.icon_url,
      articlePublishedTime: data.value.created.toString(),
      articleModifiedTime: data.value.updated.toString()
    })
  } else {
    useKunDisableSeo(data.value.name)
  }
} else {
  useKunDisableSeo('未找到该网站')
}
</script>

<template>
  <div v-if="data" class="grid grid-cols-1 gap-3 lg:grid-cols-3">
    <div class="min-w-0 space-y-3 lg:col-span-2">
      <KunCard
        :is-transparent="false"
        :is-hoverable="false"
        class-name="p-6"
        content-class="space-y-6"
      >
        <div class="flex items-start space-x-6">
          <div class="flex-shrink-0">
            <KunImage
              :src="data.icon_url"
              :alt="data.name"
              class="h-20 w-20 rounded-2xl object-cover"
            />
          </div>
          <div class="space-y-3">
            <h1 class="text-default-900 text-3xl font-bold">
              {{ data.name }}
            </h1>

            <div class="text-default-500 flex items-center space-x-6 text-sm">
              <div class="flex items-center space-x-1">
                <KunIcon name="lucide:eye" />
                <span>{{ formatNumber(data.view) }} 次访问</span>
              </div>
              <div class="flex items-center space-x-1">
                <KunIcon name="lucide:clock" />
                <span>更新于 <KunTime :time="data.updated" type="date" /></span>
              </div>
            </div>
          </div>
        </div>

        <p class="text-default-600 text-lg leading-relaxed">
          {{ data.description }}
        </p>

        <WebsiteDetailTagVisualization :tags="data.tags" />

        <WebsiteOperation :website="data" @refresh="refresh" />
      </KunCard>

      <WebsiteCommentCommunityContainer :website-id="data.id" />
    </div>

    <div class="min-w-0 space-y-3">
      <WebsiteDetailInfo :data="data" />

      <KunCard :is-transparent="false" :is-hoverable="false" class-name="p-6">
        <h3 class="text-default-900 mb-4 text-lg font-semibold">相关标签</h3>
        <div class="flex flex-wrap gap-2">
          <WebsiteTag :tags="data.tags" :is-nav="true" />
        </div>
      </KunCard>

      <WebsiteDetailStats :data="data" />
    </div>
  </div>
</template>
