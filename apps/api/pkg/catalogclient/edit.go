package catalogclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type EditProposal struct {
	ID              int64           `json:"id"`
	EntityType      string          `json:"entity_type"`
	EntityID        int64           `json:"entity_id"`
	BaseRevisionSeq int             `json:"base_revision_seq"`
	Patch           map[string]any  `json:"patch"`
	EffectivePatch  map[string]any  `json:"effective_patch,omitempty"`
	ProposerUID     int64           `json:"proposer_uid"`
	Note            string          `json:"note"`
	Site            string          `json:"site"`
	Status          string          `json:"status"`
	DecidedByUID    *int64          `json:"decided_by_uid,omitempty"`
	DecidedAt       *time.Time      `json:"decided_at,omitempty"`
	DecisionNote    string          `json:"decision_note,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	Amendments      []EditAmendment `json:"amendments,omitempty"`
}

type EditAmendment struct {
	ID         int64          `json:"id"`
	Seq        int            `json:"seq"`
	Set        map[string]any `json:"set,omitempty"`
	Unset      []string       `json:"unset,omitempty"`
	AmenderUID int64          `json:"amender_uid"`
	Note       string         `json:"note"`
	CreatedAt  time.Time      `json:"created_at"`
}

func (a *EditAmendment) UnmarshalJSON(b []byte) error {
	var aux struct {
		ID         json.RawMessage `json:"id"`
		Seq        int             `json:"seq"`
		Set        map[string]any  `json:"set"`
		Unset      []string        `json:"unset"`
		AmenderUID json.RawMessage `json:"amender_uid"`
		Note       string          `json:"note"`
		CreatedAt  time.Time       `json:"created_at"`
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	a.ID = parseFlexID(aux.ID)
	a.Seq = aux.Seq
	a.Set = aux.Set
	a.Unset = aux.Unset
	a.AmenderUID = parseFlexID(aux.AmenderUID)
	a.Note = aux.Note
	a.CreatedAt = aux.CreatedAt
	return nil
}

type EditRevision struct {
	ID            int64          `json:"id"`
	Seq           int            `json:"seq"`
	Action        string         `json:"action"`
	ChangedFields []string       `json:"changed_fields"`
	Snapshot      map[string]any `json:"snapshot"`
	ActorUID      int64          `json:"actor_uid"`
	AmenderUID    *int64         `json:"amender_uid,omitempty"`
	ProposalID    *int64         `json:"proposal_id,omitempty"`
	Site          string         `json:"site"`
	CreatedAt     time.Time      `json:"created_at"`
	LegacyAction  string         `json:"legacy_action,omitempty"`
	LegacyNote    string         `json:"legacy_note,omitempty"`
	LegacyMinor   bool           `json:"legacy_minor,omitempty"`
	LegacyID      *int64         `json:"legacy_id,omitempty"`
}

func (r *EditRevision) UnmarshalJSON(b []byte) error {
	var aux struct {
		ID            json.RawMessage `json:"id"`
		Seq           int             `json:"seq"`
		Action        string          `json:"action"`
		ChangedFields []string        `json:"changed_fields"`
		Snapshot      map[string]any  `json:"snapshot"`
		ActorUID      json.RawMessage `json:"actor_uid"`
		AmenderUID    json.RawMessage `json:"amender_uid"`
		ProposalID    json.RawMessage `json:"proposal_id"`
		Site          string          `json:"site"`
		CreatedAt     time.Time       `json:"created_at"`
		LegacyAction  string          `json:"legacy_action"`
		LegacyNote    string          `json:"legacy_note"`
		LegacyMinor   bool            `json:"legacy_minor"`
		LegacyID      json.RawMessage `json:"legacy_id"`
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	r.ID = parseFlexID(aux.ID)
	r.Seq = aux.Seq
	r.Action = aux.Action
	r.ChangedFields = aux.ChangedFields
	r.Snapshot = aux.Snapshot
	r.ActorUID = parseFlexID(aux.ActorUID)
	r.Site = aux.Site
	r.CreatedAt = aux.CreatedAt
	r.LegacyAction = aux.LegacyAction
	r.LegacyNote = aux.LegacyNote
	r.LegacyMinor = aux.LegacyMinor
	if n := parseFlexID(aux.AmenderUID); n != 0 {
		r.AmenderUID = &n
	}
	if n := parseFlexID(aux.ProposalID); n != 0 {
		r.ProposalID = &n
	}
	if n := parseFlexID(aux.LegacyID); n != 0 {
		r.LegacyID = &n
	}
	return nil
}

type EditRevertResult struct {
	Proposal EditProposal `json:"proposal"`
	Revision EditRevision `json:"revision"`
}

type EditSchemaField struct {
	Key            string `json:"key"`
	Kind           string `json:"kind"`
	DiffHint       string `json:"diff_hint"`
	Deprecated     bool   `json:"deprecated,omitempty"`
	Locked         bool   `json:"locked"`
	CanPropose     bool   `json:"can_propose"`
	CanReview      bool   `json:"can_review"`
	WouldAutomerge bool   `json:"would_automerge"`
}

type EditSchema struct {
	EntityType string            `json:"entity_type"`
	Fields     []EditSchemaField `json:"fields"`
}

type EditFieldDiff struct {
	Key      string `json:"key"`
	Kind     string `json:"kind,omitempty"`
	DiffHint string `json:"diff_hint,omitempty"`
	From     any    `json:"from"`
	To       any    `json:"to"`
}

type EditDiff struct {
	FromSeq int             `json:"from_seq"`
	ToSeq   int             `json:"to_seq"`
	Fields  []EditFieldDiff `json:"fields"`
}

type EditCreateResult struct {
	Proposal EditProposal  `json:"proposal"`
	Merged   bool          `json:"merged"`
	Revision *EditRevision `json:"revision,omitempty"`
}

type EditProposalFilter struct {
	EntityType  string
	EntityID    int64
	Site        string
	ProposerUID int64
	Status      string
	Limit       int
}

type EditAPIError struct {
	Status  int
	Code    int
	Message string
}

func (e *EditAPIError) Error() string {
	return fmt.Sprintf("catalog edit: status=%d code=%d %s", e.Status, e.Code, e.Message)
}

const editBase = "/api/v1/catalog/edit"

const EntityTypeWork = "catalog.work"

const FieldKeyPrefix = EntityTypeWork + "."

func (c *Client) ListEditProposals(ctx context.Context, f EditProposalFilter) ([]EditProposal, error) {
	page, err := c.listProposalsPage(ctx, proposalQuery(f))
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func proposalQuery(f EditProposalFilter) url.Values {
	q := url.Values{}
	if f.EntityType != "" {
		q.Set("entity_type", f.EntityType)
	}
	if f.EntityID > 0 {
		q.Set("entity_id", strconv.FormatInt(f.EntityID, 10))
	}
	if f.Site != "" {
		q.Set("site", f.Site)
	}
	if f.ProposerUID > 0 {
		q.Set("proposer_uid", strconv.FormatInt(f.ProposerUID, 10))
	}
	if f.Status != "" {
		q.Set("status", f.Status)
	}
	if f.Limit > 0 {
		q.Set("limit", strconv.Itoa(f.Limit))
	}
	return q
}

func (c *Client) CountEditProposals(ctx context.Context, f EditProposalFilter) (int64, error) {
	f.Limit = 1
	page, err := c.listProposalsPage(ctx, proposalQuery(f))
	if err != nil {
		return 0, err
	}
	return page.Total, nil
}

type proposalListPage struct {
	Items []EditProposal `json:"items"`
	Total int64          `json:"total"`
}

func (c *Client) listProposalsPage(ctx context.Context, q url.Values) (*proposalListPage, error) {
	return editGet[proposalListPage](ctx, c, editBase+"/proposals?"+q.Encode())
}

func (c *Client) ListEditRevisions(ctx context.Context, entityType string, entityID int64, limit int) ([]EditRevision, error) {
	q := url.Values{}
	q.Set("entity_type", entityType)
	q.Set("entity_id", strconv.FormatInt(entityID, 10))
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	data, err := editGet[struct {
		Items []EditRevision `json:"items"`
	}](ctx, c, editBase+"/revisions?"+q.Encode())
	if err != nil {
		return nil, err
	}
	return data.Items, nil
}

func (c *Client) DiffEditRevisions(ctx context.Context, entityType string, entityID int64, fromSeq, toSeq int) (*EditDiff, error) {
	q := url.Values{}
	q.Set("entity_type", entityType)
	q.Set("entity_id", strconv.FormatInt(entityID, 10))
	q.Set("from_seq", strconv.Itoa(fromSeq))
	q.Set("to_seq", strconv.Itoa(toSeq))
	return editGet[EditDiff](ctx, c, editBase+"/diff?"+q.Encode())
}

func editGet[T any](ctx context.Context, c *Client, path string) (*T, error) {
	return editDo[T](ctx, c, http.MethodGet, path, nil)
}

func editDo[T any](ctx context.Context, c *Client, method, path string, body []byte) (*T, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", c.basicAuth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	defer resp.Body.Close()

	var env struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, fmt.Errorf("%w: malformed envelope", ErrUpstream)
	}
	if resp.StatusCode != http.StatusOK || env.Code != 0 {
		return nil, &EditAPIError{Status: resp.StatusCode, Code: env.Code, Message: env.Message}
	}
	var out T
	if len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, &out); err != nil {
			return nil, fmt.Errorf("%w: malformed data payload", ErrUpstream)
		}
	}
	return &out, nil
}
