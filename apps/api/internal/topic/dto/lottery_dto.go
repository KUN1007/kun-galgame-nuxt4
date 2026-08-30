package dto

import "time"

type LotteryPrizeInput struct {
	Name        string   `json:"name" validate:"required,min=1,max=100"`
	Description string   `json:"description" validate:"max=500"`
	ImageHashes []string `json:"image_hashes" validate:"max=9,dive,min=1,max=128"`
	NSFWHashes  []string `json:"nsfw_hashes" validate:"max=9,dive,min=1,max=128"`
	Delivery    string   `json:"delivery" validate:"required,oneof=code manual point"`
	PointMode   string   `json:"point_mode" validate:"omitempty,oneof=fixed split random"`
	PointAmount int      `json:"point_amount" validate:"min=0,max=10000"`
	Slots       int      `json:"slots" validate:"required,min=1,max=500"`
	Codes       []string `json:"codes" validate:"max=500,dive,min=1,max=200"`
}

type CreateLotteryRequest struct {
	TopicID     int    `json:"topic_id" validate:"required,min=1"`
	Title       string `json:"title" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=1000"`

	EntryMode     string  `json:"entry_mode" validate:"required,oneof=signup reply floor"`
	FloorRule     string  `json:"floor_rule" validate:"max=200"`
	DrawMode      string  `json:"draw_mode" validate:"required,oneof=deadline manual threshold"`
	DrawThreshold int     `json:"draw_threshold" validate:"min=0,max=100000"`
	Deadline      *string `json:"deadline"`

	MinAccountAgeDays int  `json:"min_account_age_days" validate:"min=0,max=3650"`
	MinMoemoepoint    int  `json:"min_moemoepoint" validate:"min=0,max=1000000"`
	ShowEntrants      bool `json:"show_entrants"`

	Prizes []LotteryPrizeInput `json:"prizes" validate:"required,min=1,max=10,dive"`
}

// UpdateLotteryRequest carries the whole prize list because prizes are only
// editable while the lottery has no entries at all; once someone has entered,
// only the scalar fields below are writable. Nothing has to be diffed.
type UpdateLotteryRequest struct {
	LotteryID   int    `json:"lottery_id" validate:"required,min=1"`
	Title       string `json:"title" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=1000"`

	EntryMode     string  `json:"entry_mode" validate:"required,oneof=signup reply floor"`
	FloorRule     string  `json:"floor_rule" validate:"max=200"`
	DrawMode      string  `json:"draw_mode" validate:"required,oneof=deadline manual threshold"`
	DrawThreshold int     `json:"draw_threshold" validate:"min=0,max=100000"`
	Deadline      *string `json:"deadline"`

	MinAccountAgeDays int  `json:"min_account_age_days" validate:"min=0,max=3650"`
	MinMoemoepoint    int  `json:"min_moemoepoint" validate:"min=0,max=1000000"`
	ShowEntrants      bool `json:"show_entrants"`

	Prizes []LotteryPrizeInput `json:"prizes" validate:"max=10,dive"`
}

type LotteryIDRequest struct {
	LotteryID int `json:"lottery_id" validate:"required,min=1"`
}

type LotteryFulfillRequest struct {
	LotteryID   int    `json:"lottery_id" validate:"required,min=1"`
	EntryID     int    `json:"entry_id" validate:"required,min=1"`
	Fulfillment string `json:"fulfillment" validate:"required,oneof=pending shipped received forfeited"`
}

type GetLotteryByTopicRequest struct {
	TopicID int `query:"topic_id" validate:"required,min=1"`
}

type LotteryPrizeResponse struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ImageHashes []string `json:"image_hashes"`
	// Parallel to ImageHashes. An entry is empty when the reader is not allowed
	// to see that image, which keeps the hash list intact for the author's edit
	// form instead of silently shortening what a save would write back.
	ImageURLs  []string `json:"image_urls"`
	NSFWHashes []string `json:"nsfw_hashes"`
	// Hashes the image service's grader called explicit. Kept apart from
	// NSFWHashes so an edit still writes back only what the author marked, and
	// so the form can show a machine mark as one the author cannot take off.
	MachineNSFWHashes []string `json:"machine_nsfw_hashes"`
	Delivery          string   `json:"delivery"`
	PointMode         string   `json:"point_mode"`
	PointAmount       int      `json:"point_amount"`
	// What the prize hands out in total, so a reader does not have to know
	// whether point_amount means per winner or per pool.
	PointTotal  int `json:"point_total"`
	Slots       int `json:"slots"`
	CodesLoaded int `json:"codes_loaded"`
}

type LotteryWinnerResponse struct {
	EntryID      int       `json:"entry_id"`
	PrizeID      int       `json:"prize_id"`
	PrizeName    string    `json:"prize_name"`
	User         KunUser   `json:"user"`
	ReplyFloor   int       `json:"reply_floor"`
	RankKey      string    `json:"rank_key"`
	Fulfillment  string    `json:"fulfillment"`
	PointAwarded int       `json:"point_awarded"`
	WonAt        time.Time `json:"won_at"`
}

// MyPrizeName / MyCodeReady describe the reader's own win. The code itself is
// never in this struct — it only comes back from the claim endpoint, because
// Nuxt writes every fetched payload into the SSR __NUXT__ blob and a code in a
// detail response is a code in the page source for every reader.
type TopicLotteryResponse struct {
	ID          int    `json:"id"`
	TopicID     int    `json:"topic_id"`
	Title       string `json:"title"`
	Description string `json:"description"`

	EntryMode     string     `json:"entry_mode"`
	FloorRule     string     `json:"floor_rule"`
	DrawMode      string     `json:"draw_mode"`
	DrawThreshold int        `json:"draw_threshold"`
	Deadline      *time.Time `json:"deadline"`

	MinAccountAgeDays int  `json:"min_account_age_days"`
	MinMoemoepoint    int  `json:"min_moemoepoint"`
	ShowEntrants      bool `json:"show_entrants"`

	Status     string     `json:"status"`
	SeedHash   string     `json:"seed_hash"`
	Seed       string     `json:"seed"`
	EntryCount int        `json:"entry_count"`
	TotalSlots int        `json:"total_slots"`
	DrawnAt    *time.Time `json:"drawn_at"`

	User    KunUser                 `json:"user"`
	Prizes  []LotteryPrizeResponse  `json:"prizes"`
	Winners []LotteryWinnerResponse `json:"winners"`

	HasEntered   bool   `json:"has_entered"`
	CanEnter     bool   `json:"can_enter"`
	EnterBlocked string `json:"enter_blocked"`

	MyEntryID      int    `json:"my_entry_id"`
	MyPrizeID      int    `json:"my_prize_id"`
	MyPrizeName    string `json:"my_prize_name"`
	MyDelivery     string `json:"my_delivery"`
	MyFulfillment  string `json:"my_fulfillment"`
	MyCodeReady    bool   `json:"my_code_ready"`
	MyPointAwarded int    `json:"my_point_awarded"`
	// Only set for a code prize: the moment the sweep stops letting the winner
	// reveal it.
	MyClaimDeadline *time.Time `json:"my_claim_deadline"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type LotteryEntrantResponse struct {
	User    KunUser   `json:"user"`
	Floor   int       `json:"reply_floor"`
	Created time.Time `json:"created"`
}

type LotteryClaimResponse struct {
	Code string `json:"code"`
}
