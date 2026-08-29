<script setup lang="ts">
import { KUN_DLSITE_ANNOUNCE_TOPIC_ID } from '~/constants/dlsite'

const props = defineProps<{
  purchaseUrl: string
  couponUrl?: string
  campaignName?: string
}>()

const hasCoupon = computed(() => Boolean(props.couponUrl))

// Falls back to the standing landing page's own terms: an empty campaignName
// means the coupon link is that page, not a campaign infra is running.
const couponLabel = computed(
  () => props.campaignName || '每 3 个月一张 ¥400 券, 单笔满 ¥1200 可用'
)
</script>

<template>
  <KunPopover v-if="hasCoupon" position="bottom-end">
    <template #trigger>
      <KunButton variant="flat" color="primary" size="sm">
        <span class="flex items-center gap-1">
          <KunIcon name="lucide:shopping-cart" />正版购买
        </span>
      </KunButton>
    </template>

    <div class="w-80 max-w-[88vw] p-2">
      <div class="px-2 pt-1 pb-2">
        <p class="text-default-800 text-sm font-medium">与 DLsite 官方合作</p>
        <p class="text-default-500 mt-0.5 text-xs">
          本站不从中获得任何收益, 全部分成用于回馈用户
        </p>
      </div>

      <KunDivider />

      <div class="mt-2 space-y-1">
        <a
          :href="couponUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="hover:bg-default-100 flex items-start gap-3 rounded-lg p-2 transition-colors"
        >
          <span
            class="bg-primary/10 text-primary flex size-8 shrink-0 items-center justify-center rounded-lg"
          >
            <KunIcon name="lucide:ticket" />
          </span>
          <span class="min-w-0">
            <span class="text-default-800 flex items-center gap-1.5 text-sm">
              领取优惠券
              <KunChip color="primary" variant="flat" size="sm"
                >建议先领</KunChip
              >
            </span>
            <span class="text-default-500 mt-0.5 block text-xs">
              {{ couponLabel }}
            </span>
          </span>
        </a>

        <a
          :href="purchaseUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="hover:bg-default-100 flex items-start gap-3 rounded-lg p-2 transition-colors"
        >
          <span
            class="bg-primary/10 text-primary flex size-8 shrink-0 items-center justify-center rounded-lg"
          >
            <KunIcon name="lucide:shopping-cart" />
          </span>
          <span class="min-w-0">
            <span class="text-default-800 block text-sm">前往 DLsite 购买</span>
            <span class="text-default-500 mt-0.5 block text-xs">
              打开本作商品页。订单额 12% 返还为点数, 汇入论坛公共账户,
              用于购买游戏分享给大家
            </span>
          </span>
        </a>
      </div>

      <template v-if="KUN_DLSITE_ANNOUNCE_TOPIC_ID">
        <KunDivider class-name="my-2" />
        <div class="px-2 pb-1">
          <KunLink size="sm" :to="`/topic/${KUN_DLSITE_ANNOUNCE_TOPIC_ID}`">
            了解这次合作的全部细节
          </KunLink>
        </div>
      </template>
    </div>
  </KunPopover>

  <KunTooltip
    v-else
    text="通过本站链接购买, 全部分成用于回馈用户, 本站不取分成"
  >
    <KunButton
      :href="purchaseUrl"
      target="_blank"
      rel="noopener noreferrer"
      variant="flat"
      color="primary"
      size="sm"
    >
      <span class="flex items-center gap-1">
        <KunIcon name="lucide:shopping-cart" />正版购买
      </span>
    </KunButton>
  </KunTooltip>
</template>
