// @vitest-environment nuxt
import { afterEach, describe, expect, it, vi } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import KunTime from './KunTime.vue'

const AUG = '2026-08-05T21:01:00.000Z'
const JUL = '2026-07-30T13:49:00.000Z'

afterEach(() => {
  vi.useRealTimers()
})

describe('KunTime', () => {
  // The message lists render KunTime inside a v-for; paging reuses the instance
  // instead of remounting it. The text used to live in a non-reactive
  // useState(useId()), so from page 2 on every row kept page 1's timestamp.
  it('re-renders when the time prop changes on a reused instance', async () => {
    const wrapper = await mountSuspended(KunTime, {
      props: { time: AUG, type: 'datetime', showYear: true }
    })
    const first = wrapper.text()

    await wrapper.setProps({ time: JUL })
    expect(wrapper.text()).not.toBe(first)
    expect(wrapper.text()).toContain('2026')
  })

  it('keeps two instances of the same shape independent', async () => {
    const a = await mountSuspended(KunTime, {
      props: { time: AUG, type: 'datetime' }
    })
    const b = await mountSuspended(KunTime, {
      props: { time: JUL, type: 'datetime' }
    })
    expect(a.text()).not.toBe(b.text())
  })

  it('still refreshes relative text on the 60s timer', async () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-08-05T21:00:30.000Z'))
    const wrapper = await mountSuspended(KunTime, {
      props: { time: '2026-08-05T21:00:00.000Z', type: 'relative' }
    })
    const first = wrapper.text()

    vi.setSystemTime(new Date('2026-08-05T21:10:30.000Z'))
    await vi.advanceTimersByTimeAsync(60_000)
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).not.toBe(first)
  })
})
