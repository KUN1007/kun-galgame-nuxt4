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
	display := ""
	if req.Fields != nil {
		if v, ok := req.Fields["catalog.work.display_name"].(string); ok {
			display = v
		}
	}
	body := map[string]any{"display_name": display}
	if req.ProductWorkID > 0 {
		body["site_work_id"] = strconv.FormatInt(req.ProductWorkID, 10)
	}
	var out v2Claim
	if err := c.userV2JSON(ctx, http.MethodPost, accessToken, "/v2/me/claims", body, &out, nil); err != nil {
		return nil, err
	}
	id := parseFlexID(out.ID)
	return &WorkSubmitResult{WorkID: id, ProductWorkID: id, ClaimState: out.State}, nil
}

type UserClaimActionRequest struct {
	ProductWorkID int64  `json:"product_work_id,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

func (c *Client) ActOnClaimUser(ctx context.Context, accessToken string, workID int64, action string, req UserClaimActionRequest) (*ClaimActionResult, error) {
	id := strconv.FormatInt(workID, 10)
	switch action {
	case ClaimActionClaim:
		body := map[string]any{"work_id": id}
		if req.ProductWorkID > 0 {
			body["site_work_id"] = strconv.FormatInt(req.ProductWorkID, 10)
		}
		var out v2Claim
		if err := c.userV2JSON(ctx, http.MethodPost, accessToken, "/v2/me/claims", body, &out, nil); err != nil {
			return nil, err
		}
		return &ClaimActionResult{WorkID: parseFlexID(out.ID), To: out.State}, nil
	case ClaimActionWithdraw:
		var out v2Claim
		if err := c.userV2JSON(ctx, http.MethodPatch, accessToken, "/v2/me/claims/"+id,
			map[string]any{"state": "withdrawn"}, &out, ifMatchStar()); err != nil {
			return nil, err
		}
		return &ClaimActionResult{WorkID: parseFlexID(out.ID), To: out.State}, nil
	case ClaimActionApprove, ClaimActionDecline:
		decision := "approve"
		if action == ClaimActionDecline {
			decision = "decline"
		}
		if err := c.userV2JSON(ctx, http.MethodPost, accessToken, "/v2/moderation/claims/"+id+"/decisions",
			map[string]any{"decision": decision, "note": req.Reason}, nil, ifMatchStar()); err != nil {
			return nil, err
		}
		to := "live"
		if action == ClaimActionDecline {
			to = "declined"
		}
		return &ClaimActionResult{WorkID: workID, To: to}, nil
	default:
		return userEditPost[ClaimActionResult](ctx, c, accessToken,
			userClaimBase+"/works/"+id+"/claim-actions/"+action, req)
	}
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
	path := "/v2/me/claims"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var page v2List[v2Claim]
	if err := c.userV2JSON(ctx, http.MethodGet, accessToken, path, nil, &page, nil); err != nil {
		return nil, err
	}
	rows := page.rows()
	out := &UserClaimPage{Items: make([]UserClaimItem, 0, len(rows))}
	if page.Total != nil {
		out.Total = *page.Total
	}
	for _, it := range rows {
		out.Items = append(out.Items, it.item())
	}
	return out, nil
}
