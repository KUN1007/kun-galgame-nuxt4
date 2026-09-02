<script setup lang="ts">
import { KUN_TOPIC_SECTION } from '~/constants/topic'
import type {
  DiscussionForumPosting,
  WithContext,
  Person,
  Comment,
  InteractionCounter
} from 'schema-dts'

definePageMeta({ key: (route) => route.path })

const route = useRoute()

const userId = storeToRefs(usePersistUserStore()).id.value
const { showKUNGalgameContentLimit } = storeToRefs(usePersistSettingsStore())
const isNsfwMode = computed(
  () =>
    showKUNGalgameContentLimit.value === 'nsfw' ||
    showKUNGalgameContentLimit.value === 'all'
)

const isShowTopic = ref(true)

const { isReplyRewriting } = storeToRefs(useTempReplyStore())
const { isEdit } = storeToRefs(useTempReplyStore())

const topicId = computed(() => {
  return parseInt((route.params as { id: string }).id)
})
provide<number>('topicId', topicId.value)

const { data } = await useKunFetch<TopicDetail>(`/topic/${topicId.value}`, {
  method: 'GET',
  watch: false,
  query: { topic_id: topicId.value }
})

onBeforeRouteLeave(async () => {
  let proceed = true
  if (isReplyRewriting.value) {
    const res =
      await useComponentMessageStore().alert(
        '确认离开界面吗？您的更改将不会保存。'
      )
    if (res) {
      useTempReplyStore().resetRewriteReplyData()
    } else {
      proceed = false
    }
  }
  isEdit.value = false
  return proceed
})

onBeforeMount(() => {
  isEdit.value = false
})

const getFirstImageSrc = (htmlString: string) => {
  const imgRegex = /<img[^>]+src="([^">]+)"/i
  const match = htmlString.match(imgRegex)

  return match ? match[1] : `${kungal.domain.main}/kungalgame.webp`
}

if (data.value) {
  const topic = data.value

  const markdown = topic.content_markdown
  const banner =
    imageTokenUrl(topic.cover_images?.[0] ?? '') ||
    getFirstImageSrc(topic.content_html)
  const created = new Date(topic.created).toString()
  const updated = topic.edited ? new Date(topic.edited).toString() : ''
  const description = computed(() =>
    truncateRunes(markdownToText(markdown).trim(), 233)
  )

  const jsonLd = computed<WithContext<DiscussionForumPosting>>(() => {
    const topicUrl = `${kungal.domain.main}/topic/${topic.id}`

    const authorSchema: Person = {
      '@type': 'Person',
      name: topic.user.name,
      url: `${kungal.domain.main}/user/${topic.user.id}`,
      image: topic.user.avatar
    }

    const interactionStatistics: InteractionCounter[] = [
      {
        '@type': 'InteractionCounter',
        interactionType: {
          '@type': 'CommentAction'
        },
        userInteractionCount: topic.reply_count
      },
      {
        '@type': 'InteractionCounter',
        interactionType: {
          '@type': 'LikeAction'
        },
        userInteractionCount: topic.like_count
      },
      {
        '@type': 'InteractionCounter',
        interactionType: {
          '@type': 'VoteAction'
        },
        userInteractionCount: topic.upvote_count
      }
    ]

    const ba = topic.best_answer
    const acceptedAnswerSchema: Comment | undefined = ba
      ? {
          '@type': 'Comment',
          text: truncateRunes(markdownToText(ba.content_markdown).trim(), 5000),
          datePublished: new Date(ba.created).toISOString(),
          url: `${topicUrl}#k${ba.floor}`,
          author: {
            '@type': 'Person',
            name: ba.user.name,
            url: `${kungal.domain.main}/user/${ba.user.id}`,
            image: ba.user.avatar
          }
        }
      : undefined

    return {
      '@context': 'https://schema.org',
      '@type': 'DiscussionForumPosting',
      mainEntityOfPage: topicUrl,
      headline: topic.title,
      description: description.value,
      image: banner,
      author: authorSchema,
      datePublished: new Date(topic.created).toISOString(),
      dateModified: topic.edited
        ? new Date(topic.edited).toISOString()
        : new Date(topic.created).toISOString(),
      interactionStatistic: interactionStatistics,
      commentCount: topic.reply_count,
      ...(acceptedAnswerSchema && { acceptedAnswer: acceptedAnswerSchema }),
      keywords: [
        ...topic.section.map((s) => KUN_TOPIC_SECTION[s]).filter(Boolean)
      ].join(', ')
    }
  })

  useHead({
    script: [
      {
        id: 'schema-org-qa-page',
        type: 'application/ld+json',
        innerHTML: jsonLd.value
      }
    ]
  })

  if (topic.is_nsfw) {
    const trustedVisitor = !!userId || isNsfwMode.value
    useKunDisableSeo(trustedVisitor ? topic.title : '')
    if (!trustedVisitor) {
      isShowTopic.value = false
    }
  } else {
    useKunSeoMeta({
      title: data.value.title,
      description: description.value,
      ogCard: { kind: 'topic', id: topic.id },
      ogType: 'article',
      articleAuthor: [`${kungal.domain.main}/user/${data.value.user.id}`],
      articlePublishedTime: created,
      articleModifiedTime: updated
    })
  }
} else {
  useKunDisableSeo(data.value ? '话题已被封禁' : '未找到此话题')
}
</script>

<template>
  <div>
    <template v-if="data && data.status !== 1">
      <TopicDetail v-if="isShowTopic" :topic="data" />

      <KunCard v-else :is-hoverable="false" :is-transparent="false">
        <p>这个话题含有 NSFW 内容, 您需要点击确认以显示这个话题</p>
        <KunButton @click="isShowTopic = true">确认显示</KunButton>
      </KunCard>
    </template>

    <KunNull v-if="!data" description="未找到这个话题" />

    <KunNull
      v-if="data && data.status === 1"
      description="话题被隐藏, 或您未开启网站 NSFW 模式"
    />
  </div>
</template>
