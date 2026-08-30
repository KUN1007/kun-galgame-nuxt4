<script setup lang="ts">
import { computed, ref } from 'vue'
import { useLottery } from '~/composables/topic/useLottery'
import {
  KUN_LOTTERY_DELIVERY,
  KUN_LOTTERY_DRAW_MODE,
  KUN_LOTTERY_ENTRY_MODE,
  KUN_LOTTERY_STATUS
} from '~/constants/topic'

const props = defineProps<{
  lottery: TopicLottery
  isTopicAdmin: boolean
}>()

const emits = defineEmits<{
  edit: [lottery: TopicLottery]
  refresh: []
}>()

const { id: currentUserId } = usePersistUserStore()
const {
  enter,
  withdraw,
  drawNow,
  cancel,
  deleteLottery,
  getEntrants,
  claimCode
} = useLottery(props.lottery.topic_id)

const isLoading = ref(false)
const isEntrantsOpen = ref(false)
const entrants = ref<TopicLotteryEntrant[]>([])
const revealedCode = ref('')
const isFairnessOpen = ref(false)

const nowMs = useState(`kun-lottery-now-${useId()}`, () => Date.now())
onMounted(() => {
  nowMs.value = Date.now()
})

const isAuthor = computed(() => props.lottery.user.id === currentUserId)
const canManage = computed(() => isAuthor.value || props.isTopicAdmin)
const isOpen = computed(() => props.lottery.status === 'open')
const isDrawn = computed(() => props.lottery.status === 'drawn')
const isFloor = computed(() => props.lottery.entry_mode === 'floor')

const statusColor = computed(() => {
  switch (props.lottery.status) {
    case 'open':
      return 'success'
    case 'drawing':
      return 'warning'
    case 'drawn':
      return 'primary'
    default:
      return 'default'
  }
})

const countdown = computed(() => {
  if (!props.lottery.deadline || !isOpen.value) {
    return ''
  }
  const remain = new Date(props.lottery.deadline).getTime() - nowMs.value
  if (remain <= 0) {
    return '即将开奖'
  }
  const days = Math.floor(remain / 86400000)
  const hours = Math.floor((remain % 86400000) / 3600000)
  const minutes = Math.floor((remain % 3600000) / 60000)
  if (days > 0) {
    return `还剩 ${days} 天 ${hours} 小时`
  }
  if (hours > 0) {
    return `还剩 ${hours} 小时 ${minutes} 分`
  }
  return `还剩 ${minutes} 分钟`
})

const thresholdProgress = computed(() => {
  if (props.lottery.draw_mode !== 'threshold' || !props.lottery.draw_threshold) {
    return 0
  }
  return Math.min(
    100,
    (props.lottery.entry_count / props.lottery.draw_threshold) * 100
  )
})

const myWin = computed(() =>
  props.lottery.my_prize_id > 0 ? props.lottery.my_prize_name : ''
)

const run = async (task: () => Promise<unknown>) => {
  isLoading.value = true
  try {
    await task()
  } finally {
    isLoading.value = false
  }
}

const handleEnter = async () => {
  if (!requireLogin()) return
  await run(async () => {
    await enter(props.lottery.id)
    emits('refresh')
  })
}

const handleWithdraw = async () => {
  await run(async () => {
    await withdraw(props.lottery.id)
    emits('refresh')
  })
}

const handleDraw = async () => {
  await run(async () => {
    if (await drawNow(props.lottery.id)) {
      emits('refresh')
    }
  })
}

const handleCancel = async () => {
  await run(async () => {
    if (await cancel(props.lottery.id)) {
      emits('refresh')
    }
  })
}

const handleDelete = async () => {
  await run(async () => {
    if (await deleteLottery(props.lottery.id)) {
      emits('refresh')
    }
  })
}

const handleShowEntrants = async () => {
  await run(async () => {
    entrants.value = (await getEntrants(props.lottery.id)) ?? []
    isEntrantsOpen.value = true
  })
}

const handleClaim = async () => {
  await run(async () => {
    const res = await claimCode(props.lottery.id)
    if (!res?.code) {
      return
    }
    revealedCode.value = res.code
    emits('refresh')
  })
}
</script>

<template>
  <KunCard
    :is-hoverable="false"
    :is-transparent="false"
    content-class="space-y-3"
  >
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex items-center gap-2">
        <KunIcon name="lucide:gift" class="text-primary text-xl" />
        <h3 class="text-lg font-bold">{{ lottery.title }}</h3>
      </div>

      <div class="flex flex-wrap items-center gap-1">
        <KunChip color="secondary">
          {{ KUN_LOTTERY_ENTRY_MODE[lottery.entry_mode] }}
        </KunChip>
        <KunChip color="default">
          {{ KUN_LOTTERY_DRAW_MODE[lottery.draw_mode] }}
        </KunChip>
        <KunChip :color="statusColor">
          {{ KUN_LOTTERY_STATUS[lottery.status] }}
        </KunChip>
      </div>
    </div>

    <p v-if="lottery.description" class="text-default-500 text-sm">
      {{ lottery.description }}
    </p>

    <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
      <div
        v-for="prize in lottery.prizes"
        :key="prize.id"
        class="border-default-200 flex items-start gap-3 rounded-md border p-3"
      >
        <KunImage
          v-if="prize.image_url"
          :src="prize.image_url"
          :alt="prize.name"
          class="size-16 shrink-0 rounded-md object-cover"
        />
        <div class="min-w-0 space-y-1">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-semibold">{{ prize.name }}</span>
            <KunChip size="sm" color="primary">{{ prize.slots }} 名</KunChip>
          </div>
          <p class="text-default-500 text-xs">
            {{ KUN_LOTTERY_DELIVERY[prize.delivery] }}
            <template v-if="prize.delivery === 'point'">
              · {{ prize.point_amount }} 萌萌点
            </template>
          </p>
          <p v-if="prize.description" class="text-default-500 text-sm">
            {{ prize.description }}
          </p>
        </div>
      </div>
    </div>

    <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-sm">
      <span v-if="!isFloor" class="text-default-500">
        {{ lottery.entry_count }} 人参与 · 共 {{ lottery.total_slots }} 个名额
      </span>
      <span v-else class="text-default-500">
        中奖楼层: {{ lottery.floor_rule }}
      </span>
      <span v-if="countdown" class="text-primary font-medium">
        {{ countdown }}
      </span>
      <span v-if="lottery.min_moemoepoint > 0" class="text-default-500">
        门槛 {{ lottery.min_moemoepoint }} 萌萌点
      </span>
      <span v-if="lottery.min_account_age_days > 0" class="text-default-500">
        需注册满 {{ lottery.min_account_age_days }} 天
      </span>
    </div>

    <KunProgress
      v-if="lottery.draw_mode === 'threshold' && isOpen"
      :value="thresholdProgress"
      color="primary"
      size="sm"
      :aria-label="`满 ${lottery.draw_threshold} 人开奖`"
    />

    <div v-if="lottery.seed_hash" class="text-default-500 space-y-1 text-xs">
      <button
        type="button"
        class="hover:text-primary flex items-center gap-1"
        @click="isFairnessOpen = !isFairnessOpen"
      >
        <KunIcon name="lucide:shield-check" />
        <span>公平性凭据</span>
        <KunIcon
          :name="isFairnessOpen ? 'lucide:chevron-up' : 'lucide:chevron-down'"
        />
      </button>

      <div v-if="isFairnessOpen" class="space-y-1">
        <p>
          随机数承诺 (创建时公示):
          <code class="break-all">{{ lottery.seed_hash }}</code>
        </p>
        <p v-if="lottery.seed">
          随机数 (开奖时公布):
          <code class="break-all">{{ lottery.seed }}</code>
        </p>
        <p v-else>随机数将在开奖后公布, 届时可自行验算中奖名单。</p>
        <p v-if="lottery.seed">
          验算方式: sha256(随机数) 应等于上面的承诺; 每位参与者的排序值 =
          HMAC-SHA256(随机数, "{{ lottery.id }}:用户 ID"), 取最小的若干名。
        </p>
      </div>
    </div>

    <TopicLotteryResult
      v-if="isDrawn"
      :lottery="lottery"
      :can-manage="canManage"
      @refresh="emits('refresh')"
    />

    <div class="flex flex-wrap items-center justify-between gap-2">
      <div class="flex flex-wrap items-center gap-2">
        <KunButton
          v-if="lottery.can_enter"
          color="primary"
          :loading="isLoading"
          @click="handleEnter"
        >
          参与抽奖
        </KunButton>

        <KunButton
          v-else-if="lottery.has_entered && isOpen"
          variant="bordered"
          :loading="isLoading"
          @click="handleWithdraw"
        >
          退出抽奖
        </KunButton>

        <span
          v-else-if="isOpen && lottery.enter_blocked"
          class="text-default-500 text-sm"
        >
          {{ lottery.enter_blocked }}
        </span>

        <KunButton
          v-if="myWin && lottery.my_code_ready && !revealedCode"
          color="success"
          :loading="isLoading"
          @click="handleClaim"
        >
          <KunIcon name="lucide:key-round" class="mr-1" />
          领取兑换码
        </KunButton>

        <KunButton
          v-if="lottery.show_entrants && !isFloor && lottery.entry_count > 0"
          variant="light"
          size="sm"
          :loading="isLoading"
          @click="handleShowEntrants"
        >
          参与名单
        </KunButton>
      </div>

      <div v-if="canManage" class="flex items-center gap-1">
        <KunTooltip v-if="isOpen" text="立即开奖">
          <KunButton
            variant="light"
            size="lg"
            :is-icon-only="true"
            :loading="isLoading"
            @click="handleDraw"
          >
            <KunIcon name="lucide:dices" />
          </KunButton>
        </KunTooltip>
        <KunTooltip v-if="isOpen" text="编辑抽奖">
          <KunButton
            variant="light"
            size="lg"
            :is-icon-only="true"
            @click="emits('edit', lottery)"
          >
            <KunIcon name="lucide:pencil" />
          </KunButton>
        </KunTooltip>
        <KunTooltip v-if="isOpen" text="取消抽奖">
          <KunButton
            variant="light"
            color="warning"
            size="lg"
            :is-icon-only="true"
            @click="handleCancel"
          >
            <KunIcon name="lucide:ban" />
          </KunButton>
        </KunTooltip>
        <KunTooltip text="删除抽奖">
          <KunButton
            variant="light"
            color="danger"
            size="lg"
            :is-icon-only="true"
            @click="handleDelete"
          >
            <KunIcon name="lucide:trash-2" />
          </KunButton>
        </KunTooltip>
      </div>
    </div>

    <KunInfo v-if="revealedCode" title="您的兑换码">
      <div class="flex flex-wrap items-center gap-3">
        <code class="text-base break-all">{{ revealedCode }}</code>
        <KunCopy :text="revealedCode" color="primary" size="sm" />
      </div>
      <p class="text-default-500 mt-2 text-sm">
        请立即保存, 关闭后需要重新点击「领取兑换码」才能再次查看。
      </p>
    </KunInfo>

    <KunModal v-model="isEntrantsOpen" inner-class-name="max-w-lg">
      <h3 class="mb-3 text-xl font-bold">参与名单</h3>
      <div class="max-h-96 space-y-2 overflow-y-auto">
        <div
          v-for="entrant in entrants"
          :key="entrant.user.id"
          class="flex items-center gap-2"
        >
          <KunAvatar :user="entrant.user" size="sm" />
          <span class="text-sm">{{ entrant.user.name }}</span>
        </div>
        <p v-if="!entrants.length" class="text-default-500 text-sm">
          还没有人参与
        </p>
      </div>
    </KunModal>
  </KunCard>
</template>
