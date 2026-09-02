<script setup lang="ts">
import {
  getGalgameCharacterLangName,
  getGalgameCharacterIntroCredit
} from '~/constants/galgameCharacter'

const route = useRoute()
const characterId = computed(() => Number((route.params as { id: string }).id))

if (!Number.isInteger(characterId.value) || characterId.value <= 0) {
  throw createError({
    statusCode: 404,
    statusMessage: '未找到该角色',
    fatal: true
  })
}

const PAGE_SIZE = 50

const { data } = await useKunFetch<GalgameCharacterDetail>(
  `/galgame-character/${characterId.value}`,
  { method: 'GET', query: { limit: PAGE_SIZE }, watch: false }
)

const moved = !!data.value?.moved_to
if (data.value?.moved_to) {
  await navigateTo(`/galgame/character/${data.value.moved_to}`, {
    redirectCode: 301,
    replace: true
  })
}

if (!data.value) {
  throw createError({
    statusCode: 404,
    statusMessage: '未找到该角色',
    fatal: true
  })
}

const works = ref<GalgameCharacterWork[]>(moved ? [] : [...data.value.works])
const nextOffset = ref<number | null>(moved ? null : data.value.next_offset)
const loadingMore = ref(false)

const loadMore = async () => {
  if (nextOffset.value === null || loadingMore.value) {
    return
  }
  loadingMore.value = true
  const res = await kunFetch<GalgameCharacterDetail>(
    `/galgame-character/${characterId.value}`,
    { method: 'GET', query: { limit: PAGE_SIZE, offset: nextOffset.value } }
  )
  loadingMore.value = false
  if (!res) {
    return
  }
  works.value.push(...res.works)
  nextOffset.value = res.next_offset
}

const voiceLabel = (voice: GalgameDetailCharacterVoice) =>
  voice.lang && voice.lang.toLowerCase() !== 'ja'
    ? `${voice.name}（${getGalgameCharacterLangName(voice.lang)}）`
    : voice.name

const voiceTitle = (voices: GalgameDetailCharacterVoice[]) =>
  voices.map((v) => v.latin || v.name).join(' / ')

const subtitle = computed(() => {
  const parts = [data.value?.name_original, data.value?.latin].filter(
    (part): part is string => !!part && part !== data.value?.name
  )
  return parts.join(' · ')
})

const figureFrame = computed(() => artFrame(data.value?.figure_meta))
const bustFrame = computed(() => artFrame(data.value?.image_meta))

const leadIntro = computed(() =>
  data.value?.intros.find((i) => i.intro === data.value?.intro)
)
const introCredit = computed(() =>
  getGalgameCharacterIntroCredit(leadIntro.value)
)
const otherIntros = computed(() =>
  (data.value?.intros ?? []).filter((i) => i.intro !== data.value?.intro)
)

const isTraitSpoilerRevealed = ref(false)
const traits = computed(() => {
  const all = data.value?.traits ?? []
  return isTraitSpoilerRevealed.value ? all : all.filter((t) => t.spoiler === 0)
})
const hiddenTraitCount = computed(
  () => (data.value?.traits ?? []).filter((t) => t.spoiler > 0).length
)
const traitGroups = computed(() => {
  const groups: { name: string; traits: GalgameCharacterTrait[] }[] = []
  for (const trait of traits.value) {
    const name = trait.group || '其他'
    const last = groups.at(-1)
    if (last && last.name === name) {
      last.traits.push(trait)
    } else {
      groups.push({ name, traits: [trait] })
    }
  }
  return groups
})

if (!moved) {
  useKunSeoMeta({
    title: `${data.value.name} 登场的 Galgame`,
    description:
      data.value.intro ||
      `角色 ${data.value.name} 在本站收录的 Galgame 中的登场作品与配音演员一览。`,
    ogCard: { kind: 'character', id: data.value.id }
  })
}
</script>

<template>
  <div v-if="data && !data.moved_to" class="space-y-6">
    <KunHeader :name="data.name" :description="subtitle">
      <template v-if="data.figure || data.image" #headerEndContent>
        <KunLightboxGallery>
          <div class="flex shrink-0 items-start gap-2">
            <KunLightboxGalleryItem
              v-if="data.figure"
              :src="data.figure"
              :alt="data.name"
              :wrap="false"
              v-slot="{ open }"
            >
              <button
                type="button"
                class="bg-default-100 w-fit cursor-zoom-in overflow-hidden rounded-lg"
                :aria-label="`查看 ${data.name} 的立绘`"
                @click="open"
              >
                <KunImage
                  :src="data.figure"
                  :alt="data.name"
                  loading="eager"
                  :aspect-ratio="figureFrame.aspectRatio"
                  :object-fit="figureFrame.objectFit"
                  :thumbhash="figureFrame.thumbhash"
                  class-name="w-40 sm:w-48"
                />
              </button>
            </KunLightboxGalleryItem>

            <KunLightboxGalleryItem
              v-if="data.image"
              :src="data.image"
              :alt="data.name"
              :wrap="false"
              v-slot="{ open }"
            >
              <button
                type="button"
                class="bg-default-100 w-fit cursor-zoom-in overflow-hidden rounded-lg"
                :aria-label="`查看 ${data.name} 的头像`"
                @click="open"
              >
                <KunImage
                  :src="data.image"
                  :alt="data.name"
                  loading="eager"
                  :aspect-ratio="bustFrame.aspectRatio"
                  :object-fit="bustFrame.objectFit"
                  :thumbhash="bustFrame.thumbhash"
                  :class-name="data.figure ? 'w-24 sm:w-28' : 'w-32 sm:w-40'"
                />
              </button>
            </KunLightboxGalleryItem>
          </div>
        </KunLightboxGallery>
      </template>

      <template #endContent>
        <div class="space-y-3">
          <div v-if="data.intro" class="space-y-1">
            <p class="text-default-600 whitespace-pre-line">{{ data.intro }}</p>
            <p v-if="introCredit" class="text-default-400 text-xs">
              {{ introCredit }}
            </p>
          </div>

          <KunAccordion v-if="otherIntros.length">
            <KunAccordionItem
              v-for="intro in otherIntros"
              :key="intro.lang"
              :value="intro.lang"
              :title="`${getGalgameCharacterLangName(intro.lang)} 简介`"
            >
              <p class="text-default-600 whitespace-pre-line">
                {{ intro.intro }}
              </p>
              <p v-if="intro.source" class="text-default-400 mt-1 text-xs">
                {{ getGalgameCharacterIntroCredit(intro) }}
              </p>
            </KunAccordionItem>
          </KunAccordion>

          <div v-if="traitGroups.length" class="space-y-2">
            <div
              v-for="group in traitGroups"
              :key="group.name"
              class="space-y-1"
            >
              <p class="text-default-400 text-xs">{{ group.name }}</p>
              <div class="flex flex-wrap gap-1.5">
                <KunChip
                  v-for="trait in group.traits"
                  :key="trait.id"
                  size="xs"
                  :color="trait.spoiler > 0 ? 'warning' : 'default'"
                >
                  {{ trait.name }}<template v-if="trait.lie">（伪）</template>
                </KunChip>
              </div>
            </div>
          </div>

          <KunButton
            v-if="hiddenTraitCount && !isTraitSpoilerRevealed"
            variant="flat"
            color="warning"
            size="sm"
            @click="isTraitSpoilerRevealed = true"
          >
            <KunIcon name="lucide:eye" />
            显示 {{ hiddenTraitCount }} 条剧透特征
          </KunButton>

          <div
            v-if="data.links.length"
            class="flex flex-wrap items-center gap-3"
          >
            <template v-for="link in data.links" :key="link.source">
              <KunLink
                v-if="link.url"
                :to="link.url"
                target="_blank"
                rel="noopener noreferrer"
                size="sm"
                color="default"
                class-name="text-default-500 hover:text-default-700"
              >
                {{ link.name }}
                <KunIcon name="lucide:external-link" class="inline size-3" />
              </KunLink>
              <span v-else class="text-default-400 text-sm">{{
                link.name
              }}</span>
            </template>
          </div>

          <p class="text-default-500 text-sm">
            资料来自 NextMoe 目录的角色图谱。默认仅显示 SFW 的 Galgame, 查看
            NSFW Galgame 请在设置面板打开 NSFW 开关。如果有数据错误请
            <KunLink to="/doc/contact"> 联系我们 </KunLink>。
          </p>
        </div>
      </template>
    </KunHeader>

    <GalgameCard v-if="works.length" :galgames="works" :is-transparent="false">
      <template #meta="{ galgame }">
        <p
          v-if="galgame.voices?.length"
          class="text-default-500 mt-1 line-clamp-2 text-xs"
          :title="voiceTitle(galgame.voices)"
        >
          CV {{ galgame.voices.map(voiceLabel).join(' / ') }}
        </p>
      </template>
    </GalgameCard>

    <KunNull v-else description="暂无该角色登场的 Galgame" />

    <div v-if="nextOffset !== null" class="flex justify-center">
      <KunButton
        variant="flat"
        color="primary"
        :is-loading="loadingMore"
        @click="loadMore"
      >
        加载更多作品
      </KunButton>
    </div>
  </div>
</template>
