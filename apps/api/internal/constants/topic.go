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
	MaxImagesPerPrize    = 9

	// A point prize mints moemoepoint rather than moving it off the author's
	// balance, so the whole lottery is capped instead of only the per-prize
	// amount: ten prizes of 10000 across 500 slots each would otherwise be a
	// legal way to create fifty million points.
	MaxLotteryPointTotal = 100000

	MaxPollsPerTopic = 30
	MaxTagsPerTopic  = 7
	MaxTagLength     = 17
)
