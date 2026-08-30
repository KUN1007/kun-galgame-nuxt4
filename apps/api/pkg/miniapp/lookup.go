// Package miniapp answers "which mini-apps does this topic carry" for the
// surfaces that only list topics. The home feed, the search results and the
// activity feed had each grown a private copy of the same topic_poll lookup,
// so the lottery shipped invisible in all three: a topic running one looked
// like any other topic everywhere except its own page.
package miniapp

import "gorm.io/gorm"

// Kind must match the keys in
// apps/web/app/components/topic/miniapp/registry.ts — the frontend picks the
// badge label and icon out of that registry by this string.
const (
	KindPoll    = "poll"
	KindLottery = "lottery"
)

var sources = []struct {
	kind  string
	table string
}{
	{KindPoll, "topic_poll"},
	{KindLottery, "topic_lottery"},
}

// ByTopic maps a topic id to the mini-app kinds it carries, in registry order.
func ByTopic(db *gorm.DB, topicIDs []int) map[int][]string {
	out := map[int][]string{}
	if len(topicIDs) == 0 {
		return out
	}
	for _, src := range sources {
		var ids []int
		if err := db.Table(src.table).
			Distinct("topic_id").
			Where("topic_id IN ?", topicIDs).
			Pluck("topic_id", &ids).Error; err != nil {
			continue
		}
		for _, id := range ids {
			out[id] = append(out[id], src.kind)
		}
	}
	return out
}
