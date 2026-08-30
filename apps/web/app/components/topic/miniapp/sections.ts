import type { Component } from 'vue'
import PollSection from '../poll/Section.vue'
import LotterySection from '../lottery/Section.vue'
import type { TopicMiniAppKey } from './registry'

// Kept out of registry.ts on purpose: the registry is also imported by the
// topic-list badge, and pulling the section components in there would drag the
// whole poll and lottery bundle onto every list page.
//
// The Record is exhaustive over TopicMiniAppKey, so adding a third mini-app to
// the registry fails typecheck here until its section is wired.
export const TOPIC_MINI_APP_SECTIONS: Record<TopicMiniAppKey, Component> = {
  poll: PollSection,
  lottery: LotterySection
}
