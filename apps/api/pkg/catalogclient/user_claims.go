package catalogclient

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const userClaimBase = userBase

type UserWorkSubmitRequest struct {
	ProductWorkID int64           `json:"product_work_id,omitempty"`
	Fields        map[string]any  `json:"fields"`
	Released      *WorkSubmitDate `json:"released,omitempty"`
}

func (c *Client) SubmitWorkUser(ctx context.Context, accessToken string, req UserWorkSubmitRequest) (*WorkSubmitResult, error) {
	return userEditPost[WorkSubmitResult](ctx, c, accessToken, userClaimBase+"/works/submit", req)
}

type UserClaimActionRequest struct {
	ProductWorkID int64  `json:"product_work_id,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

func (c *Client) ActOnClaimUser(ctx context.Context, accessToken string, workID int64, action string, req UserClaimActionRequest) (*ClaimActionResult, error) {
	return userEditPost[ClaimActionResult](ctx, c, accessToken,
		userClaimBase+"/works/"+strconv.FormatInt(workID, 10)+"/claim-actions/"+action, req)
}

func (c *Client) DeleteDraftUser(ctx context.Context, accessToken string, workID int64) error {
	_, err := userEditDo[struct{}](ctx, c, http.MethodDelete, accessToken,
		userClaimBase+"/works/"+strconv.FormatInt(workID, 10), nil)
	return err
}

func (c *Client) MyClaims(ctx context.Context, accessToken string, f UserClaimFilter) (*UserClaimPage, error) {
	q := url.Values{}
	if len(f.ClaimStates) > 0 {
		q.Set("claim_state", strings.Join(f.ClaimStates, ","))
	}
	if f.Before > 0 {
		q.Set("before", strconv.FormatInt(f.Before, 10))
	}
	if f.Limit > 0 {
		q.Set("limit", strconv.Itoa(f.Limit))
	}
	if f.Kind != "" {
		q.Set("kind", f.Kind)
	}
	path := userClaimBase + "/claims/mine"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	return userEditDo[UserClaimPage](ctx, c, http.MethodGet, accessToken, path, nil)
}
