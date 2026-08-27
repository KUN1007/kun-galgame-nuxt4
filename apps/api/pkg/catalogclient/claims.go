package catalogclient

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

const (
	ClaimStateNone     = "none"
	ClaimStateLive     = "live"
	ClaimStateDraft    = "draft"
	ClaimStatePending  = "pending"
	ClaimStateDeclined = "declined"
	ClaimStateHidden   = "hidden"
)

const (
	ClaimKindSubmitted = "submitted"
	ClaimKindAudited   = "audited"
)

const (
	ClaimActionClaim    = "claim"
	ClaimActionSubmit   = "submit"
	ClaimActionPublish  = "publish"
	ClaimActionWithdraw = "withdraw"
	ClaimActionApprove  = "approve"
	ClaimActionDecline  = "decline"
	ClaimActionBan      = "ban"
	ClaimActionUnban    = "unban"
)

type ClaimActionResult struct {
	WorkID  int64   `json:"work_id"`
	From    *string `json:"from_state"`
	To      string  `json:"to_state"`
	EventID int64   `json:"event_id"`
}

type WorkSubmitDate struct {
	Y int16 `json:"y"`
	M int16 `json:"m,omitempty"`
	D int16 `json:"d,omitempty"`
}

type WorkSubmitResult struct {
	WorkID        int64  `json:"work_id"`
	ProductWorkID int64  `json:"product_work_id"`
	ClaimState    string `json:"claim_state"`
	EventID       int64  `json:"event_id"`
	ReleaseID     int64  `json:"release_id,omitempty"`
}

type ClaimEventFeedItem struct {
	ID            int64     `json:"id"`
	WorkID        int64     `json:"work_id"`
	FromState     *string   `json:"from_state"`
	ToState       string    `json:"to_state"`
	ActorUID      int64     `json:"actor_uid"`
	Reason        *string   `json:"reason"`
	Site          string    `json:"site"`
	ProductWorkID *int64    `json:"product_work_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type ClaimEventFeedPage struct {
	Items     []ClaimEventFeedItem `json:"items"`
	NextSince int64                `json:"next_since"`
}

// The last route the forum still reads over v1, and it has no replacement: as of
// 2026-08-28 /v2/catalog/changes answers only {target_object, id, updated_at} —
// no state transition, no actor, no site — and it accepts from= and actor=
// without applying them, so three calls with different filters returned the same
// first row. The moemoepoint award, the stub cleanup and the unpublish cron all
// need the transition and the actor, so they cannot move until v2 grows a feed
// that carries them. Retiring v1 without that replacement stops moemoepoint
// site-wide, silently, exactly the way the 2026-06 outage did.
func (c *Client) ClaimEventsSince(ctx context.Context, since int64, limit int, site string) (*ClaimEventFeedPage, error) {
	q := url.Values{
		"since": {strconv.FormatInt(since, 10)},
		"limit": {strconv.Itoa(limit)},
	}
	if site != "" {
		q.Set("site", site)
	}
	return editGetQuery[ClaimEventFeedPage](ctx, c, "/api/v1/catalog/claim-events/feed", q)
}

type UserClaimItem struct {
	WorkID        int64  `json:"work_id"`
	DisplayName   string `json:"display_name"`
	Site          string `json:"site"`
	ProductWorkID *int64 `json:"product_work_id"`
	ClaimState    string `json:"claim_state"`

	LastEventID   int64     `json:"last_event_id"`
	LastFromState *string   `json:"last_from_state"`
	LastToState   string    `json:"last_to_state"`
	LastReason    *string   `json:"last_reason"`
	LastActorUID  int64     `json:"last_actor_uid"`
	LastEventAt   time.Time `json:"last_event_at"`

	FirstActedAt time.Time `json:"first_acted_at"`
	ActedCount   int       `json:"acted_count"`
}

type UserClaimPage struct {
	Items      []UserClaimItem `json:"items"`
	NextBefore int64           `json:"next_before"`
	Total      int64           `json:"total"`
}

// Kind ("submitted"/"audited") only exists on the bearer /claims/mine face.
// The by-uid face /catalog/users/{uid}/claims answers ClaimsByActor — every work
// the user TOUCHED — and huma drops unknown query params without erroring, so
// asking it to filter by kind returns someone else's works silently. That face
// has no client here on purpose; a per-user owned list comes from
// GalgameRepository.PublishedIDsByCreator.
type UserClaimFilter struct {
	ClaimStates []string
	Before      int64
	Limit       int
	Kind        string
}

func editGetQuery[T any](ctx context.Context, c *Client, path string, q url.Values) (*T, error) {
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	return editGet[T](ctx, c, path)
}
