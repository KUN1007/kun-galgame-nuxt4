import type { LotteryFormData } from '~/components/topic/lottery/types'

const toPayload = (data: LotteryFormData) => ({
  title: data.title,
  description: data.description,
  entry_mode: data.entry_mode,
  floor_rule: data.entry_mode === 'floor' ? data.floor_rule : '',
  draw_mode: data.draw_mode,
  draw_threshold: data.draw_threshold,
  deadline: data.deadline,
  min_account_age_days: data.min_account_age_days,
  min_moemoepoint: data.min_moemoepoint,
  show_entrants: data.show_entrants,
  prizes: data.prizes.map((p) => ({
    name: p.name,
    description: p.description,
    image_hash: p.image_hash,
    delivery: p.delivery,
    point_amount: p.delivery === 'point' ? p.point_amount : 0,
    slots: p.slots,
    codes:
      p.delivery === 'code'
        ? p.codes
            .split('\n')
            .map((c) => c.trim())
            .filter(Boolean)
        : []
  }))
})

export const useLottery = (topicId: number) => {
  const base = `/topic/${topicId}/lottery`

  const getLotteries = () =>
    useKunFetch<TopicLottery[]>(`${base}/topic`, {
      query: { topic_id: topicId },
      lazy: true
    })

  const getEntrants = (lotteryId: number) =>
    kunFetch<TopicLotteryEntrant[]>(`${base}/entrants`, {
      query: { lottery_id: lotteryId }
    })

  const createLottery = (data: LotteryFormData) =>
    kunFetch<string>(base, {
      method: 'POST',
      body: { topic_id: topicId, ...toPayload(data) }
    })

  const updateLottery = (lotteryId: number, data: LotteryFormData) =>
    kunFetch<string>(base, {
      method: 'PUT',
      body: { lottery_id: lotteryId, ...toPayload(data) }
    })

  const deleteLottery = async (lotteryId: number) => {
    const confirmed = await useComponentMessageStore().alert(
      '确定要删除这个抽奖吗？',
      '删除后所有参与记录、中奖名单与托管的兑换码都会一并丢失, 该操作不可恢复!'
    )
    if (!confirmed) {
      return false
    }
    await kunFetch<string>(base, {
      method: 'DELETE',
      query: { lottery_id: lotteryId }
    })
    return true
  }

  const enter = (lotteryId: number) =>
    kunFetch<string>(`${base}/enter`, {
      method: 'POST',
      body: { lottery_id: lotteryId }
    })

  const withdraw = (lotteryId: number) =>
    kunFetch<string>(`${base}/withdraw`, {
      method: 'POST',
      body: { lottery_id: lotteryId }
    })

  const drawNow = async (lotteryId: number) => {
    const confirmed = await useComponentMessageStore().alert(
      '确定现在开奖吗？',
      '开奖后中奖名单立即固定, 无法撤销, 也不能再修改奖项。'
    )
    if (!confirmed) {
      return false
    }
    await kunFetch<string>(`${base}/draw`, {
      method: 'POST',
      body: { lottery_id: lotteryId }
    })
    return true
  }

  const cancel = async (lotteryId: number) => {
    const confirmed = await useComponentMessageStore().alert(
      '确定要取消这个抽奖吗？',
      '取消后将不再开奖, 已参与的用户不会获得任何奖品。'
    )
    if (!confirmed) {
      return false
    }
    await kunFetch<string>(`${base}/cancel`, {
      method: 'POST',
      body: { lottery_id: lotteryId }
    })
    return true
  }

  // POST on purpose. The code must not travel on a page-load fetch: Nuxt inlines
  // every SSR payload into the __NUXT__ blob, which is readable in page source.
  const claimCode = (lotteryId: number) =>
    kunFetch<{ code: string }>(`${base}/claim`, {
      method: 'POST',
      body: { lottery_id: lotteryId }
    })

  const setFulfillment = (
    lotteryId: number,
    entryId: number,
    fulfillment: string
  ) =>
    kunFetch<string>(`${base}/fulfillment`, {
      method: 'PUT',
      body: { lottery_id: lotteryId, entry_id: entryId, fulfillment }
    })

  return {
    getLotteries,
    getEntrants,
    createLottery,
    updateLottery,
    deleteLottery,
    enter,
    withdraw,
    drawNow,
    cancel,
    claimCode,
    setFulfillment
  }
}
