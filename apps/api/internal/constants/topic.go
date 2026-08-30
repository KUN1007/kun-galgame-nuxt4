package constants

var TopicSectionConsume = map[string]bool{
	"g-seeking": true,
	"g-other":   true,
	"t-help":    true,
}

var ValidTopicCategories = []string{"galgame", "technique", "others"}

var ValidTopicSortFields = map[string]string{
	"created":            "created",
	"view":               "view",
	"view_7d":            "view_7d",
	"view_30d":           "view_30d",
	"status_update_time": "status_update_time",
}

var ValidTopicCountSortFields = map[string]string{
	"like":     "like_count",
	"favorite": "favorite_count",
	"upvote":   "upvote_count",
}

const (
	// A lottery hands out scarce goods, so the entry bar is higher than the
	// poll's: either an account old enough to be inconvenient to farm, or enough
	// moemoepoint that the account has actually contributed. Moderators bypass
	// both via lottery.create_any.
	LotteryMinAccountAgeDays = 30
	LotteryMinMoemoepoint    = 100

	MaxLotteriesPerTopic = 10
	MaxPrizesPerLottery  = 10
	MaxSlotsPerPrize     = 500

	MaxPollsPerTopic = 30
	MaxTagsPerTopic  = 7
	MaxTagLength     = 17
)
