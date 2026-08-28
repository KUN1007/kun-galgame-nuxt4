// @vitest-environment nuxt
import { describe, expect, it } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import GalgameTag from './Tag.vue'

const tags = [
  {
    id: 1,
    name: '青梅竹马',
    category: 'content',
    spoiler_level: 0,
    galgame_count: 3
  },
  {
    id: 2,
    name: '成人标签',
    category: 'sexual',
    spoiler_level: 0,
    galgame_count: 1
  }
] as unknown as GalgameDetailTag[]

// The category set is read once at setup, so the store has to carry the
// reader's mode before the component mounts.
const mountWith = async (contentLimit: 'sfw' | 'nsfw') => {
  usePersistSettingsStore().showKUNGalgameContentLimit = contentLimit
  return mountSuspended(GalgameTag, { props: { tags, variant: 'mobile' } })
}

describe('GalgameTag', () => {
  // The server strips every sexual tag from the detail payload for an SFW
  // reader, so ticking 成人内容 could never match anything and the panel said
  // nothing about why.
  it('locks the sexual category and says why in SFW mode', async () => {
    const wrapper = await mountWith('sfw')
    const boxes = wrapper.findAll('input[type="checkbox"]')
    expect((boxes[2]!.element as HTMLInputElement).disabled).toBe(true)
    expect(wrapper.text()).toContain('NSFW 开关')
  })

  it('leaves it alone once NSFW is on', async () => {
    const wrapper = await mountWith('nsfw')
    const boxes = wrapper.findAll('input[type="checkbox"]')
    expect((boxes[2]!.element as HTMLInputElement).disabled).toBe(false)
    expect(wrapper.text()).not.toContain('成人内容标签未加载')
  })
})
