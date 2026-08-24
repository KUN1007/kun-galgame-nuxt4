package service

import (
	"context"
	stderrors "errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"kun-galgame-api/internal/constants"
	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/repository"
	"kun-galgame-api/internal/moemoepoint"
	"kun-galgame-api/pkg/catalogclient"
	"kun-galgame-api/pkg/errors"

	"gorm.io/gorm"
)

type SubmissionService struct {
	galgameClient *client.GalgameClient
	catalog       *catalogclient.Client
	galgameRepo   *repository.GalgameRepository
}

func NewSubmissionService(
	galgameClient *client.GalgameClient,
	catalog *catalogclient.Client,
	galgameRepo *repository.GalgameRepository,
) *SubmissionService {
	return &SubmissionService{
		galgameClient: galgameClient,
		catalog:       catalog,
		galgameRepo:   galgameRepo,
	}
}

const submissionSite = client.ClaimSiteKungal

type SubmitResult struct {
	GID        int    `json:"gid"`
	WorkID     int64  `json:"work_id"`
	ClaimState string `json:"claim_state"`
	// False when the submitter uploaded a banner that did not make it onto the
	// work. The submission itself still stands, so this is reported rather than
	// raised — but it has to be reported: silence here hid a broken cover patch
	// for the whole life of the feature.
	BannerAttached bool `json:"banner_attached"`
}

func (s *SubmissionService) Submit(
	ctx context.Context,
	accessToken string,
	uid int,
	form *SubmissionForm,
) (*SubmitResult, *errors.AppError) {
	if form.DisplayName() == "" {
		return nil, errors.ErrValidation("请至少填写一个语言的标题")
	}
	released, appErr := form.Released()
	if appErr != nil {
		return nil, appErr
	}
	res, err := s.catalog.SubmitWorkUser(ctx, accessToken, catalogclient.UserWorkSubmitRequest{
		Fields: form.Fields(), Released: released,
	})
	if err != nil {
		return nil, claimActionError(err)
	}
	gid := int(res.ProductWorkID)
	// Stamp the submitter now rather than waiting for the claim feed to say so
	// ten minutes later: it is what authorises the submitter to open their own
	// unpublished entry, and it survives a Redis flush. The row stays
	// published=false, so it shows up in no list and no feed.
	if uid > 0 {
		if err := s.galgameRepo.SubmitLocal(s.galgameRepo.DB().WithContext(ctx), gid, uid); err != nil {
			slog.Warn("submit: 记录本地投稿人失败", "gid", gid, "uid", uid, "error", err)
		}
	}
	patch := form.CoverPatch()
	attached := patch == nil
	if patch != nil {
		created, err := s.catalog.CreateEditProposalUser(ctx, accessToken, catalogclient.UserEditCreateRequest{
			EntityType: catalogclient.EntityTypeWork, EntityID: res.WorkID,
			Patch: patch, Note: "投稿时提交的横幅图",
		})
		switch {
		case err != nil:
			slog.Error("submit: 附加横幅图失败, 投稿已建立但封面丢失", "work", res.WorkID, "error", err)
		case created.Merged:
			attached = true
		default:
			// Approving the claim moves only the claim state, never the open
			// edit proposals, so a banner that stays open here is never shown.
			// The submitter owns the work, so merge it now instead.
			if _, err := s.catalog.MergeEditProposalUser(ctx, accessToken, created.Proposal.ID, ""); err != nil {
				slog.Error("submit: 合并横幅图提案失败", "proposal", created.Proposal.ID, "error", err)
			} else {
				attached = true
			}
		}
	}
	return &SubmitResult{
		GID: gid, WorkID: res.WorkID, ClaimState: res.ClaimState, BannerAttached: attached,
	}, nil
}

func (s *SubmissionService) Claim(
	ctx context.Context,
	accessToken string,
	uid int64,
	gid int,
) (*catalogclient.ClaimActionResult, *errors.AppError) {
	res, appErr := s.act(ctx, accessToken, gid, catalogclient.ClaimActionPublish, "")
	if appErr != nil {
		return nil, appErr
	}
	if err := s.galgameRepo.Touch(s.galgameRepo.DB().WithContext(ctx), gid); err != nil {
		slog.Warn("claim: 刷新本地 galgame resource_update_time 失败", "gid", gid, "error", err)
	}
	moemoepoint.Award(int(uid), constants.RewardCreateGalgame,
		moemoepoint.ReasonContentApproved, moemoepoint.Ref("galgame", gid),
		moemoepoint.Key("claim", strconv.Itoa(gid), strconv.FormatInt(uid, 10)))
	return res, nil
}

// The wizard row for an unclaimed work carries no gid, so the response has to
// name the one the claim just minted rather than let the caller assume it.
type ClaimUnclaimedResult struct {
	catalogclient.ClaimActionResult
	GID int `json:"gid"`
}

// ClaimUnclaimed adopts a bodyless catalog work (claim state "none") for the
// user and publishes it in one call: claim (none → draft) then publish
// (draft → live).
//
// The claim anchors product_work_id at the catalog work id, which is what
// Submit ends up with too (it omits the field and the registry mints an
// identity the claim then adopts). It is also the only id the forum has here:
// catalog rejects a claim with no product_work_id.
//
// That anchor is NOT free of collisions in principle. catalog_work's unique
// index is (medium_id, site, product_work_id), and for the legacy rows the two
// id spaces differ — 61,329 of the 62,250 claims in the dev registry carry a
// product_work_id that is not their own catalog id. An unclaimed work whose
// catalog id happens to equal one of those anchors cannot be adopted. Measured
// against dev: 310 of 152,635 unclaimed works sit inside the anchor range at
// all, and none of them collides. If one ever does, catalog refuses the claim
// and the user sees the upstream error; there is no id the forum could pick
// instead, because a forum-chosen anchor would break the gid = catalog id
// convention every other path depends on.
func (s *SubmissionService) ClaimUnclaimed(
	ctx context.Context,
	accessToken string,
	uid int64,
	workID int64,
) (*ClaimUnclaimedResult, *errors.AppError) {
	res, appErr := adoptAndPublish(ctx, s.catalog, accessToken, workID)
	if appErr != nil {
		return nil, appErr
	}

	gid := int(workID)
	// Stamp the row now instead of waiting for the claim-event cron, exactly as
	// Submit does: the caller is redirected to /galgame/:gid the moment this
	// returns, and until the row exists that page renders IsOnForum=false — a
	// "未收录" banner and no resource or comment section, right after telling
	// the user 认领成功, 已发布.
	if err := s.galgameRepo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.galgameRepo.PublishLocal(tx, gid); err != nil {
			return err
		}
		return s.galgameRepo.SetCreatorIfUnset(tx, gid, int(uid))
	}); err != nil {
		slog.Warn("claim: 建立本地 galgame 行失败, 等待 claim 事件同步补齐",
			"gid", gid, "uid", uid, "error", err)
	}
	if err := s.galgameRepo.Touch(s.galgameRepo.DB().WithContext(ctx), gid); err != nil {
		slog.Warn("claim: 刷新本地 galgame resource_update_time 失败", "gid", gid, "error", err)
	}
	moemoepoint.Award(int(uid), constants.RewardCreateGalgame,
		moemoepoint.ReasonContentApproved, moemoepoint.Ref("galgame", gid),
		moemoepoint.Key("claim", strconv.Itoa(gid), strconv.FormatInt(uid, 10)))
	return &ClaimUnclaimedResult{ClaimActionResult: *res, GID: gid}, nil
}

// adoptAndPublish is one user gesture but two transactions upstream, so the
// claim failing is only fatal when the publish fails as well. That is what
// makes a retry work: the second attempt's claim is refused (the work is
// already this user's draft) while its publish succeeds. Returning early on the
// claim error instead left a half-adopted work with no way to finish it — and
// the wizard would keep offering it as unclaimed until the next daily reindex,
// so every retry hit the same refused claim.
func adoptAndPublish(
	ctx context.Context,
	catalog *catalogclient.Client,
	accessToken string,
	workID int64,
) (*catalogclient.ClaimActionResult, *errors.AppError) {
	_, claimErr := catalog.ActOnClaimUser(ctx, accessToken, workID,
		catalogclient.ClaimActionClaim, catalogclient.UserClaimActionRequest{ProductWorkID: workID})

	res, pubErr := catalog.ActOnClaimUser(ctx, accessToken, workID,
		catalogclient.ClaimActionPublish, catalogclient.UserClaimActionRequest{})
	if pubErr == nil {
		return res, nil
	}
	if claimErr != nil {
		return nil, claimActionError(claimErr)
	}
	return nil, claimActionError(pubErr)
}

func (s *SubmissionService) Resubmit(
	ctx context.Context,
	accessToken string,
	gid int,
) (*catalogclient.ClaimActionResult, *errors.AppError) {
	return s.act(ctx, accessToken, gid, catalogclient.ClaimActionSubmit, "")
}

func (s *SubmissionService) Withdraw(
	ctx context.Context,
	accessToken string,
	gid int,
) (*catalogclient.ClaimActionResult, *errors.AppError) {
	return s.act(ctx, accessToken, gid, catalogclient.ClaimActionWithdraw, "")
}

func (s *SubmissionService) DeleteDraft(
	ctx context.Context,
	accessToken string,
	gid int,
) *errors.AppError {
	workID, appErr := s.workIDOf(ctx, gid)
	if appErr != nil {
		return appErr
	}
	if err := s.catalog.DeleteDraftUser(ctx, accessToken, workID); err != nil {
		return claimActionError(err)
	}
	if err := s.galgameRepo.DeleteLocal(gid); err != nil {
		slog.Warn("delete draft: 删除本地 galgame 行失败", "gid", gid, "error", err)
	}
	return nil
}

func (s *SubmissionService) act(
	ctx context.Context,
	accessToken string,
	gid int,
	action string,
	reason string,
) (*catalogclient.ClaimActionResult, *errors.AppError) {
	workID, appErr := s.workIDOf(ctx, gid)
	if appErr != nil {
		return nil, appErr
	}
	res, err := s.catalog.ActOnClaimUser(ctx, accessToken, workID, action, catalogclient.UserClaimActionRequest{
		Reason: reason,
	})
	if err != nil {
		return nil, claimActionError(err)
	}
	return res, nil
}

func (s *SubmissionService) workIDOf(ctx context.Context, gid int) (int64, *errors.AppError) {
	ids, appErr := s.galgameClient.CatalogWorkIDs(ctx, []int{gid})
	if appErr != nil {
		return 0, appErr
	}
	workID, ok := ids[gid]
	if !ok {
		return 0, errors.ErrNotFound("条目不存在")
	}
	return workID, nil
}

func claimActionError(err error) *errors.AppError {
	var apiErr *catalogclient.UserAPIError
	switch {
	case stderrors.Is(err, catalogclient.ErrInsufficientScope):
		return errors.ErrReauthRequired("投稿需要新的授权，请退出登录后重新登录以授予该权限")
	case stderrors.Is(err, catalogclient.ErrUnauthorized):
		return errors.ErrAuthExpired()
	case stderrors.Is(err, catalogclient.ErrNotFound):
		return errors.ErrNotFound("条目不存在")
	case stderrors.Is(err, catalogclient.ErrNotConfigured):
		return errors.New(errors.CodeBiz, "资料库服务暂不可用", http.StatusServiceUnavailable)
	case stderrors.As(err, &apiErr):
		switch apiErr.Status {
		case http.StatusForbidden:
			return errors.ErrForbidden("你没有权限执行此操作")
		case http.StatusUnprocessableEntity:
			return errors.ErrValidation(apiErr.Message)
		case http.StatusConflict:
			return errors.New(errors.CodeBiz, apiErr.Message, http.StatusConflict)
		}
		slog.Error("claim action: 上游错误", "status", apiErr.Status, "code", apiErr.Code, "msg", apiErr.Message)
	default:
		slog.Warn("claim action: catalog 不可达", "error", err)
	}
	return errors.New(errors.CodeBiz, "资料库服务暂不可用", http.StatusServiceUnavailable)
}

const claimPageLimit = 20

var mineStates = []string{
	catalogclient.ClaimStatePending,
	catalogclient.ClaimStateDeclined,
	catalogclient.ClaimStateDraft,
}

var claimStates = map[string]bool{
	catalogclient.ClaimStateNone:     true,
	catalogclient.ClaimStateLive:     true,
	catalogclient.ClaimStateDraft:    true,
	catalogclient.ClaimStatePending:  true,
	catalogclient.ClaimStateDeclined: true,
	catalogclient.ClaimStateHidden:   true,
}

func (s *SubmissionService) ListMine(
	ctx context.Context,
	accessToken string,
	query url.Values,
) (*catalogclient.UserClaimPage, *errors.AppError) {
	states := mineStates
	if raw := query.Get("claim_state"); raw != "" {
		states = splitCSV(raw)
		for _, st := range states {
			if !claimStates[st] {
				return nil, errors.ErrBadRequest("未知的申请状态: " + st)
			}
		}
	}
	return s.listClaims(ctx, accessToken, query, catalogclient.ClaimKindSubmitted, states)
}

func (s *SubmissionService) ListAudit(
	ctx context.Context,
	accessToken string,
	query url.Values,
) (*catalogclient.UserClaimPage, *errors.AppError) {
	return s.listClaims(ctx, accessToken, query, catalogclient.ClaimKindAudited, nil)
}

func (s *SubmissionService) listClaims(
	ctx context.Context,
	accessToken string,
	query url.Values,
	kind string,
	states []string,
) (*catalogclient.UserClaimPage, *errors.AppError) {
	page, err := s.catalog.MyClaims(ctx, accessToken, catalogclient.UserClaimFilter{
		ClaimStates: states,
		Before:      int64(atoiOr(query.Get("before"), 0)),
		Limit:       atoiOr(query.Get("limit"), claimPageLimit),
		Kind:        kind,
	})
	if err != nil {
		return nil, claimActionError(err)
	}
	if page.Items == nil {
		page.Items = []catalogclient.UserClaimItem{}
	}
	return page, nil
}

const wizardSearchInclude = "names,covers,refs"

const wizardDefaultLimit = 12

type WizardSearchPage struct {
	Items   []client.GalgameBrief         `json:"items"`
	Pending []catalogclient.UserClaimItem `json:"pending"`
	Total   int64                         `json:"total"`
}

func (s *SubmissionService) SearchWithPending(
	ctx context.Context,
	accessToken string,
	query url.Values,
) (*WizardSearchPage, *errors.AppError) {
	items, total, appErr := s.wizardItems(ctx, query)
	if appErr != nil {
		return nil, appErr
	}
	pending, appErr := s.wizardPending(ctx, accessToken)
	if appErr != nil {
		return nil, appErr
	}
	return &WizardSearchPage{Items: items, Pending: pending, Total: total}, nil
}

func (s *SubmissionService) wizardItems(
	ctx context.Context,
	query url.Values,
) ([]client.GalgameBrief, int64, *errors.AppError) {
	// Neither claim_state nor claimed here on purpose: both are the index's and
	// lag a day, while the claimed_by CatalogItemWizardEligible reads is the
	// registry's. Gating on both hid every work that entered an actionable state
	// since the last reindex — an approved submission stayed unfindable for a
	// day — and claimed=true additionally hid the whole unclaimed supply, which
	// is exactly what the wizard's 认领并发布 button exists to adopt.
	q := url.Values{
		"q":       {query.Get("q")},
		"page":    {strconv.Itoa(atoiOr(query.Get("page"), 1))},
		"limit":   {strconv.Itoa(atoiOr(query.Get("limit"), wizardDefaultLimit))},
		"include": {wizardSearchInclude},
	}
	client.OpenPopulation(q)

	res, appErr := s.galgameClient.CatalogWorksSearch(ctx, q)
	if appErr != nil {
		return nil, 0, appErr
	}
	items := make([]client.GalgameBrief, 0, len(res.Items))
	for i := range res.Items {
		row := &res.Items[i]
		if !client.CatalogItemWizardEligible(row) {
			continue
		}
		b := client.CatalogItemToBrief(ctx, row)
		items = append(items, b)
	}
	return items, res.Total, nil
}

func (s *SubmissionService) wizardPending(
	ctx context.Context,
	accessToken string,
) ([]catalogclient.UserClaimItem, *errors.AppError) {
	page, err := s.catalog.MyClaims(ctx, accessToken, catalogclient.UserClaimFilter{
		ClaimStates: []string{catalogclient.ClaimStatePending, catalogclient.ClaimStateDeclined},
		Limit:       wizardDefaultLimit,
		Kind:        catalogclient.ClaimKindSubmitted,
	})
	if err != nil {
		return nil, claimActionError(err)
	}
	if page.Items == nil {
		return []catalogclient.UserClaimItem{}, nil
	}
	return page.Items, nil
}
