<script setup lang="ts">
import { computed, ref } from 'vue'
import { useLottery } from '~/composables/topic/useLottery'
import { KUN_LOTTERY_FULFILLMENT } from '~/constants/topic'
import type { KunUIColor } from '@kungal/ui-core'

const props = defineProps<{
  lottery: TopicLottery
  canManage: boolean
}>()

const emits = defineEmits<{ refresh: [] }>()

const { setFulfillment } = useLottery(props.lottery.topic_id)
const isLoading = ref(false)

const { id: currentUserId } = usePersistUserStore()

const grouped = computed(() => {
  const byPrize = new Map<number, TopicLotteryWinner[]>()
  for (const winner of props.lottery.winners) {
    const list = byPrize.get(winner.prize_id) ?? []
    list.push(winner)
    byPrize.set(winner.prize_id, list)
  }
  return props.lottery.prizes
    .map((prize) => ({ prize, winners: byPrize.get(prize.id) ?? [] }))
    .filter((group) => group.winners.length > 0)
})

const isEmpty = computed(() => props.lottery.winners.length === 0)

const FULFILLMENT_COLOR: Record<string, KunUIColor> = {
  pending: 'warning',
  shipped: 'primary',
  received: 'success',
  forfeited: 'default'
}

const myEntry = computed(() =>
  props.lottery.winners.find((w) => w.user.id === currentUserId)
)

const advance = async (entryId: number, fulfillment: string) => {
  isLoading.value = true
  try {
    await setFulfillment(props.lottery.id, entryId, fulfillment)
    emits('refresh')
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="border-default-200 space-y-3 rounded-lg border p-3">
    <div class="flex items-center gap-2">
      <KunIcon name="lucide:party-popper" class="text-secondary-500" />
      <span class="text-sm font-semibold">中奖名单</span>
    </div>

    <p v-if="isEmpty" class="text-default-500 text-sm">
      本次抽奖没有产生中奖者。
    </p>

    <div v-for="group in grouped" :key="group.prize.id" class="space-y-2">
      <div class="text-default-500 text-xs">
        {{ group.prize.name }} · {{ group.winners.length }}/{{
          group.prize.slots
        }}
        名
      </div>
      <div class="flex flex-wrap gap-2">
        <div
          v-for="winner in group.winners"
          :key="winner.entry_id"
          class="border-default-200 flex items-center gap-2 rounded-lg border py-1 pr-2 pl-1"
        >
          <KunAvatar :user="winner.user" size="sm" />
          <span class="text-sm">{{ winner.user.name }}</span>
          <KunChip
            v-if="winner.reply_floor > 0"
            size="sm"
            variant="flat"
            color="secondary"
          >
            {{ winner.reply_floor }} 楼
          </KunChip>
          <KunChip
            v-if="winner.point_awarded > 0"
            size="sm"
            variant="flat"
            color="primary"
          >
            {{ winner.point_awarded }} 萌萌点
          </KunChip>
          <KunChip
            size="sm"
            variant="flat"
            :color="FULFILLMENT_COLOR[winner.fulfillment] ?? 'warning'"
          >
            {{ KUN_LOTTERY_FULFILLMENT[winner.fulfillment] ?? '待发放' }}
          </KunChip>
          <KunButton
            v-if="canManage && winner.fulfillment === 'pending'"
            variant="light"
            size="sm"
            :loading="isLoading"
            @click="advance(winner.entry_id, 'shipped')"
          >
            标记已发出
          </KunButton>
        </div>
      </div>
    </div>

    <KunInfo v-if="myEntry" title="您中奖了">
      <div class="space-y-2 text-sm">
        <p>
          您获得了「{{ myEntry.prize_name }}」, 当前状态:
          {{ KUN_LOTTERY_FULFILLMENT[myEntry.fulfillment] ?? '待发放' }}。
        </p>
        <p v-if="lottery.my_delivery === 'manual'">
          实物类奖品由发起人直接联系您。本站只负责产生名单, 不担保履约,
          请自行与对方确认发货方式, 不要在公开楼层留下收货地址。
        </p>
        <p v-else-if="lottery.my_delivery === 'point'">
          {{ myEntry.point_awarded }} 萌萌点已自动发放到您的账户。
        </p>
        <p v-else-if="myEntry.fulfillment === 'forfeited'">
          该兑换码已作废, 无法再领取。
        </p>
        <p v-else-if="lottery.my_claim_deadline">
          请在
          <KunTime
            :time="lottery.my_claim_deadline"
            type="datetime"
            show-year
          />
          前领取兑换码, 逾期作废。
        </p>
        <div
          v-if="
            myEntry.fulfillment !== 'received' &&
            myEntry.fulfillment !== 'forfeited'
          "
          class="flex gap-2"
        >
          <KunButton
            size="sm"
            color="primary"
            :loading="isLoading"
            @click="advance(myEntry.entry_id, 'received')"
          >
            确认已收到
          </KunButton>
          <KunButton
            size="sm"
            variant="light"
            color="danger"
            :loading="isLoading"
            @click="advance(myEntry.entry_id, 'forfeited')"
          >
            放弃奖品
          </KunButton>
        </div>
      </div>
    </KunInfo>
  </div>
</template>
