package model

import "time"

const (
	LotteryEntrySignup = "signup"
	LotteryEntryReply  = "reply"
	LotteryEntryFloor  = "floor"

	LotteryDrawDeadline  = "deadline"
	LotteryDrawManual    = "manual"
	LotteryDrawThreshold = "threshold"

	LotteryStatusOpen = "open"
	// The sweep flips open -> drawing in the same statement that selects the
	// lottery, so a draw that outlives its minute is not started a second time
	// and the prizes are not handed out twice.
	LotteryStatusDrawing   = "drawing"
	LotteryStatusDrawn     = "drawn"
	LotteryStatusCancelled = "cancelled"

	LotteryDeliveryCode   = "code"
	LotteryDeliveryManual = "manual"
	LotteryDeliveryPoint  = "point"

	LotteryFulfillPending   = "pending"
	LotteryFulfillShipped   = "shipped"
	LotteryFulfillReceived  = "received"
	LotteryFulfillForfeited = "forfeited"
)

type TopicLottery struct {
	ID          int    `gorm:"primaryKey;autoIncrement" json:"id"`
	TopicID     int    `gorm:"column:topic_id;not null" json:"topic_id"`
	UserID      int    `gorm:"column:user_id;not null" json:"user_id"`
	Title       string `gorm:"type:varchar(100);not null" json:"title"`
	Description string `gorm:"type:varchar(1000);default:''" json:"description"`

	EntryMode     string     `gorm:"column:entry_mode;default:'signup'" json:"entry_mode"`
	FloorRule     string     `gorm:"column:floor_rule;default:''" json:"floor_rule"`
	DrawMode      string     `gorm:"column:draw_mode;default:'deadline'" json:"draw_mode"`
	DrawThreshold int        `gorm:"column:draw_threshold;default:0" json:"draw_threshold"`
	Deadline      *time.Time `gorm:"column:deadline" json:"deadline"`

	MinAccountAgeDays int  `gorm:"column:min_account_age_days;default:0" json:"min_account_age_days"`
	MinMoemoepoint    int  `gorm:"column:min_moemoepoint;default:0" json:"min_moemoepoint"`
	ShowEntrants      bool `gorm:"column:show_entrants;default:true" json:"show_entrants"`

	Status     string     `gorm:"default:'open'" json:"status"`
	SeedHash   string     `gorm:"column:seed_hash;default:''" json:"seed_hash"`
	Seed       string     `gorm:"column:seed;default:''" json:"-"`
	EntryCount int        `gorm:"column:entry_count;default:0" json:"entry_count"`
	DrawnAt    *time.Time `gorm:"column:drawn_at" json:"drawn_at"`

	CreatedAt time.Time `gorm:"column:created" json:"created"`
	UpdatedAt time.Time `gorm:"column:updated" json:"updated"`
}

func (TopicLottery) TableName() string { return "topic_lottery" }

type TopicLotteryPrize struct {
	ID          int    `gorm:"primaryKey;autoIncrement" json:"id"`
	LotteryID   int    `gorm:"column:lottery_id;not null" json:"lottery_id"`
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `gorm:"type:varchar(500);default:''" json:"description"`
	ImageHash   string `gorm:"column:image_hash;default:''" json:"image_hash"`
	Delivery    string `gorm:"column:delivery;default:'manual'" json:"delivery"`
	PointAmount int    `gorm:"column:point_amount;default:0" json:"point_amount"`
	Slots       int    `gorm:"column:slots;default:1" json:"slots"`
	SortOrder   int    `gorm:"column:sort_order;default:0" json:"sort_order"`

	CreatedAt time.Time `gorm:"column:created" json:"created"`
	UpdatedAt time.Time `gorm:"column:updated" json:"updated"`
}

func (TopicLotteryPrize) TableName() string { return "topic_lottery_prize" }

// Secret is the AES-GCM sealed activation code. It has no json tag on purpose:
// this struct must never be reachable from a response body.
type TopicLotteryCode struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	LotteryID int    `gorm:"column:lottery_id;not null"`
	PrizeID   int    `gorm:"column:prize_id;not null"`
	Secret    string `gorm:"column:secret;not null"`
	ClaimedBy int    `gorm:"column:claimed_by;default:0"`

	ClaimedAt *time.Time `gorm:"column:claimed_at"`
	CreatedAt time.Time  `gorm:"column:created"`
}

func (TopicLotteryCode) TableName() string { return "topic_lottery_code" }

type TopicLotteryEntry struct {
	ID         int `gorm:"primaryKey;autoIncrement" json:"id"`
	LotteryID  int `gorm:"column:lottery_id;not null" json:"lottery_id"`
	UserID     int `gorm:"column:user_id;not null" json:"user_id"`
	ReplyFloor int `gorm:"column:reply_floor;default:0" json:"reply_floor"`

	PrizeID     int        `gorm:"column:prize_id;default:0" json:"prize_id"`
	CodeID      int        `gorm:"column:code_id;default:0" json:"-"`
	RankKey     string     `gorm:"column:rank_key;default:''" json:"rank_key"`
	Fulfillment string     `gorm:"column:fulfillment;default:''" json:"fulfillment"`
	WonAt       *time.Time `gorm:"column:won_at" json:"won_at"`

	CreatedAt time.Time `gorm:"column:created" json:"created"`
	UpdatedAt time.Time `gorm:"column:updated" json:"updated"`
}

func (TopicLotteryEntry) TableName() string { return "topic_lottery_entry" }
