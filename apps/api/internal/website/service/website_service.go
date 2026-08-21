package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"

	"kun-galgame-api/internal/infrastructure/markdown"
	"kun-galgame-api/internal/website/dto"
	"kun-galgame-api/internal/website/model"
	"kun-galgame-api/internal/website/repository"
	"kun-galgame-api/pkg/communityclient"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/imageclient"
	"kun-galgame-api/pkg/userclient"

	"gorm.io/gorm"
)

func marshalDomain(domains []string) json.RawMessage {
	if len(domains) == 0 {
		return json.RawMessage("[]")
	}
	b, err := json.Marshal(domains)
	if err != nil {
		return json.RawMessage("[]")
	}
	return json.RawMessage(b)
}

type WebsiteService struct {
	websiteRepo  *repository.WebsiteRepository
	categoryRepo *repository.CategoryRepository
	tagRepo      *repository.TagRepository
	userClient   *userclient.Client
	community    *communityclient.Client
	cdnBase      string
}

func NewWebsiteService(
	websiteRepo *repository.WebsiteRepository,
	categoryRepo *repository.CategoryRepository,
	tagRepo *repository.TagRepository,
	userClient *userclient.Client,
	community *communityclient.Client,
	cdnBase string,
) *WebsiteService {
	return &WebsiteService{
		websiteRepo:  websiteRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		userClient:   userClient,
		community:    community,
		cdnBase:      cdnBase,
	}
}

func (s *WebsiteService) GetList(isSFW bool) []dto.WebsiteCard {
	rows := s.websiteRepo.FindAll(isSFW)
	catMap := s.categoryRepo.FindNamesByIDs(collectCategoryIDs(rows))
	levelMap := s.tagRepo.LevelSumsAll()
	return websiteCardsFromRows(rows, catMap, levelMap, s.cdnBase)
}

func normalizeStatus(status string) string {
	if status == "" {
		return "normal"
	}
	return status
}

// `name` and `url` are both UNIQUE. Without this the duplicate came back as a
// 500 "创建网站失败" with the real cause — `duplicate key value violates unique
// constraint "galgame_website_name_key"` — swallowed, so re-submitting an
// already-listed site looked like a server fault.
func (s *WebsiteService) conflictError(name, url string, excludeID int) *errors.AppError {
	existing, err := s.websiteRepo.FindConflict(name, url, excludeID)
	if err != nil || existing == nil {
		return nil
	}
	if existing.Name == name {
		return errors.ErrBadRequest("网站名称「" + name + "」已被 " + existing.URL + " 使用")
	}
	return errors.ErrBadRequest("主域名 " + url + " 已经收录过了: 「" + existing.Name + "」")
}

func (s *WebsiteService) Create(userID int, req *dto.CreateWebsiteRequest) *errors.AppError {
	if appErr := s.conflictError(req.Name, req.URL, 0); appErr != nil {
		return appErr
	}

	txErr := s.websiteRepo.DB().Transaction(func(tx *gorm.DB) error {
		website := model.GalgameWebsite{
			Name:          req.Name,
			URL:           req.URL,
			Description:   req.Description,
			Icon:          req.Icon,
			IconImageHash: req.IconImageHash,
			Language:      req.Language,
			AgeLimit:      req.AgeLimit,
			Status:        normalizeStatus(req.Status),
			CategoryID:    req.CategoryID,
			UserID:        userID,
			CreateTime:    req.CreateTime,
			Domain:        marshalDomain(req.Domain),
		}
		if err := s.websiteRepo.Create(tx, &website); err != nil {
			return err
		}
		return s.tagRepo.InsertWebsiteTagRelations(tx, website.ID, req.TagIDs)
	})
	if txErr != nil {
		slog.Error("website create failed", "name", req.Name, "url", req.URL, "error", txErr)
		return errors.ErrInternal("创建网站失败")
	}
	return nil
}

func (s *WebsiteService) GetDetail(
	ctx context.Context,
	domain string,
	currentUserID int,
) (*dto.WebsiteDetailResponse, *errors.AppError) {
	website, err := s.websiteRepo.FindByDomain(domain)
	if err != nil {
		return nil, errors.ErrNotFound("未找到该网站")
	}

	go s.websiteRepo.IncrementView(website.ID)

	category, _ := s.categoryRepo.FindByID(website.CategoryID)
	catBrief := dto.WebsiteCategoryBrief{}
	if category != nil {
		catBrief = dto.WebsiteCategoryBrief{
			ID:          category.ID,
			Name:        category.Name,
			Label:       category.Label,
			Description: category.Description,
		}
	}

	rels := s.tagRepo.FindRelationsByWebsiteWithTag(website.ID)
	tags := make([]dto.WebsiteTagBrief, len(rels))
	for i, tr := range rels {
		tags[i] = dto.WebsiteTagBrief{
			ID:          tr.Tag.ID,
			Name:        tr.Tag.Name,
			Description: tr.Tag.Description,
			Label:       tr.Tag.Label,
			Level:       tr.Tag.Level,
			GroupID:     tr.Tag.GroupID,
		}
	}

	commentList := s.resolveDetailComments(ctx, website.ID)

	isLiked, isFavorited := false, false
	if currentUserID > 0 {
		isLiked = s.websiteRepo.HasLike(currentUserID, website.ID)
		isFavorited = s.websiteRepo.HasFavorite(currentUserID, website.ID)
	}

	return &dto.WebsiteDetailResponse{
		ID:            website.ID,
		Name:          website.Name,
		URL:           website.URL,
		Description:   website.Description,
		Icon:          website.Icon,
		IconImageHash: website.IconImageHash,
		IconURL:       imageclient.ResolveURL(s.cdnBase, website.IconImageHash, website.Icon),
		View:          website.View,
		Language:      website.Language,
		AgeLimit:      website.AgeLimit,
		Status:        website.Status,
		Category:      catBrief,
		Tags:          tags,
		LikeCount:     website.LikeCount,
		IsLiked:       isLiked,
		FavoriteCount: website.FavoriteCount,
		IsFavorited:   isFavorited,
		Domain:        website.Domain,
		CreateTime:    website.CreateTime,
		Comment:       commentList,
		Created:       website.CreatedAt,
		Updated:       website.UpdatedAt,
	}, nil
}

const websiteDetailCommentCap = 20

func (s *WebsiteService) resolveDetailComments(ctx context.Context, websiteID int) []dto.WebsiteDetailComment {
	out := []dto.WebsiteDetailComment{}
	thread, err := s.community.ResolveComments(ctx, communityclient.ResolveCommentsRequest{
		AnchorKind:    communityclient.AnchorSiteResource,
		AnchorID:      "website:" + strconv.Itoa(websiteID),
		ContentRating: communityclient.RatingAll,
	})
	if err != nil {
		slog.Warn("website detail: community resolve failed (best-effort)", "website_id", websiteID, "error", err)
		return out
	}

	uids := make([]int, 0, len(thread.Posts))
	for _, p := range thread.Posts {
		uids = append(uids, int(p.AuthorID))
	}
	userMap := s.userClient.Hydrate(ctx, uids)

	for _, p := range thread.Posts {
		if len(out) >= websiteDetailCommentCap {
			break
		}
		if p.Status != communityclient.PostVisible {
			continue
		}
		u := userMap[int(p.AuthorID)]
		if !userclient.IsRenderable(u) {
			continue
		}
		updated := p.EditedAt
		if updated == "" {
			updated = p.CreatedAt
		}
		out = append(out, dto.WebsiteDetailComment{
			ID:      int(p.ID),
			Content: markdown.ToPlainText(p.ContentRaw, 1007),
			User:    dto.UserBriefCompact{ID: u.ID, Name: u.Name, Avatar: u.Avatar},
			Created: p.CreatedAt,
			Updated: updated,
		})
	}
	return out
}

func (s *WebsiteService) Update(req *dto.UpdateWebsiteRequest) *errors.AppError {
	if appErr := s.conflictError(req.Name, req.URL, req.WebsiteID); appErr != nil {
		return appErr
	}

	txErr := s.websiteRepo.DB().Transaction(func(tx *gorm.DB) error {
		if err := s.websiteRepo.UpdateFields(tx, req.WebsiteID, map[string]any{
			"name":            req.Name,
			"url":             req.URL,
			"description":     req.Description,
			"icon":            req.Icon,
			"icon_image_hash": req.IconImageHash,
			"category_id":     req.CategoryID,
			"age_limit":       req.AgeLimit,
			"status":          normalizeStatus(req.Status),
			"language":        req.Language,
			"create_time":     req.CreateTime,
			"domain":          marshalDomain(req.Domain),
		}); err != nil {
			return err
		}
		return s.tagRepo.ReplaceWebsiteTagRelations(tx, req.WebsiteID, req.TagIDs)
	})
	if txErr != nil {
		slog.Error("website update failed", "website_id", req.WebsiteID, "error", txErr)
		return errors.ErrInternal("更新网站失败")
	}
	return nil
}

func (s *WebsiteService) Delete(websiteID int) *errors.AppError {
	if err := s.websiteRepo.DeleteByID(websiteID); err != nil {
		return errors.ErrInternal("删除网站失败")
	}
	return nil
}

func (s *WebsiteService) ToggleLike(userID, websiteID int) *errors.AppError {
	txErr := s.websiteRepo.DB().Transaction(func(tx *gorm.DB) error {
		existing, err := s.websiteRepo.FindLike(tx, userID, websiteID)
		if err == gorm.ErrRecordNotFound {
			if err := s.websiteRepo.CreateLike(tx, userID, websiteID); err != nil {
				return err
			}
			return s.websiteRepo.AdjustLikeCount(tx, websiteID, 1)
		}
		if err == nil && existing != nil {
			if err := s.websiteRepo.DeleteLike(tx, existing); err != nil {
				return err
			}
			return s.websiteRepo.AdjustLikeCount(tx, websiteID, -1)
		}
		return err
	})
	if txErr != nil {
		return errors.ErrInternal("点赞失败")
	}
	return nil
}

func (s *WebsiteService) ToggleFavorite(userID, websiteID int) *errors.AppError {
	txErr := s.websiteRepo.DB().Transaction(func(tx *gorm.DB) error {
		existing, err := s.websiteRepo.FindFavorite(tx, userID, websiteID)
		if err == gorm.ErrRecordNotFound {
			if err := s.websiteRepo.CreateFavorite(tx, userID, websiteID); err != nil {
				return err
			}
			return s.websiteRepo.AdjustFavoriteCount(tx, websiteID, 1)
		}
		if err == nil && existing != nil {
			if err := s.websiteRepo.DeleteFavorite(tx, existing); err != nil {
				return err
			}
			return s.websiteRepo.AdjustFavoriteCount(tx, websiteID, -1)
		}
		return err
	})
	if txErr != nil {
		return errors.ErrInternal("收藏失败")
	}
	return nil
}
