package service

import (
	"context"
	"strconv"

	"kun-galgame-api/internal/constants"
	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/internal/galgame/model"
	"kun-galgame-api/internal/galgame/repository"
	"kun-galgame-api/internal/moemoepoint"
	userRepo "kun-galgame-api/internal/user/repository"
	"kun-galgame-api/pkg/catalogclient"
	"kun-galgame-api/pkg/dlsite"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/userclient"
	"kun-galgame-api/pkg/utils"

	"gorm.io/gorm"
)

type GalgameService struct {
	galgameRepo        *repository.GalgameRepository
	interactionRepo    *repository.GalgameInteractionRepository
	listRepo           *repository.GalgameListRepository
	resourceMetaRepo   *repository.GalgameResourceMetaRepository
	detailRatingRepo   *repository.GalgameDetailRatingRepository
	contributorRepo    *repository.GalgameContributorRepository
	stateRepo          *userRepo.StateRepository
	galgameClient      *client.GalgameClient
	userClient         *userclient.Client
	catalog            *catalogclient.Client
	helpers            InteractionHelpers
	dlsiteLinkTemplate string
	dlsiteCouponURL    string
}

func NewGalgameService(
	galgameRepo *repository.GalgameRepository,
	interactionRepo *repository.GalgameInteractionRepository,
	listRepo *repository.GalgameListRepository,
	resourceMetaRepo *repository.GalgameResourceMetaRepository,
	detailRatingRepo *repository.GalgameDetailRatingRepository,
	contributorRepo *repository.GalgameContributorRepository,
	stateRepo *userRepo.StateRepository,
	galgameClient *client.GalgameClient,
	userClient *userclient.Client,
	catalog *catalogclient.Client,
	dlsiteLinkTemplate string,
	dlsiteCouponURL string,
) *GalgameService {
	return &GalgameService{
		galgameRepo:        galgameRepo,
		interactionRepo:    interactionRepo,
		listRepo:           listRepo,
		resourceMetaRepo:   resourceMetaRepo,
		dlsiteLinkTemplate: dlsiteLinkTemplate,
		dlsiteCouponURL:    dlsiteCouponURL,
		detailRatingRepo:   detailRatingRepo,
		contributorRepo:    contributorRepo,
		stateRepo:          stateRepo,
		galgameClient:      galgameClient,
		userClient:         userClient,
		catalog:            catalog,
	}
}

func (s *GalgameService) ToggleLike(
	ctx context.Context,
	userID, galgameID int,
) *errors.AppError {
	ownerID, name := s.fetchOwnerAndName(ctx, galgameID)
	if ownerID == userID {
		return errors.ErrBadRequest("您不能给自己点赞")
	}

	txErr := s.galgameRepo.DB().Transaction(func(tx *gorm.DB) error {
		liked, err := s.interactionRepo.ToggleLike(tx, userID, galgameID)
		if err != nil {
			return err
		}
		if !liked {
			s.helpers.AdjustMoemoepoint(tx, ownerID, -1,
				moemoepoint.ReasonLiked, moemoepoint.Ref("galgame", galgameID))
			return nil
		}
		s.helpers.AdjustMoemoepoint(tx, ownerID, 1,
			moemoepoint.ReasonLiked, moemoepoint.Ref("galgame", galgameID))
		return s.helpers.CreateGalgameMessageWithContent(tx, userID, ownerID, "liked", name, galgameID)
	})
	if txErr != nil {
		return errors.ErrInternal("点赞失败")
	}
	return nil
}

func (s *GalgameService) GetMyInteractions(userID int) dto.MyGalgameInteractions {
	liked, favorited := s.interactionRepo.UserGalgameInteractions(userID)
	return dto.MyGalgameInteractions{Liked: liked, Favorited: favorited}
}

func (s *GalgameService) fetchOwnerAndName(ctx context.Context, galgameID int) (int, string) {
	return s.ownerOf(galgameID), truncate(s.entryName(ctx, galgameID), constants.TextPreviewLength)
}

func (s *GalgameService) ownerOf(galgameID int) int {
	if s.galgameRepo == nil || galgameID <= 0 {
		return 0
	}
	row := s.galgameRepo.FindLocal(galgameID)
	if row.CreatorUserID == nil {
		return 0
	}
	return *row.CreatorUserID
}

func (s *GalgameService) entryName(ctx context.Context, galgameID int) string {
	if s.galgameClient == nil {
		return ""
	}
	rows, appErr := s.galgameClient.CatalogRowsByGIDs(ctx, []int{galgameID}, "names", "all")
	if appErr != nil {
		return ""
	}
	row, ok := rows[galgameID]
	if !ok {
		return ""
	}
	brief := client.CatalogItemToBrief(ctx, &row)
	return client.BriefName(&brief)
}

func (s *GalgameService) GetDetail(
	ctx context.Context,
	galgameID, currentUserID int,
	token string,
	isSFW bool,
) (*dto.GalgameDetail, *errors.AppError) {
	d, found, appErr := s.galgameClient.CatalogWorkDetail(ctx, galgameID)
	if appErr != nil {
		return nil, appErr
	}
	if !found {
		return nil, errors.ErrNotFound("未找到该 Galgame")
	}
	g := client.CatalogDetailToFull(ctx, d, galgameID)
	s.galgameClient.HydrateOfficialLinks(ctx, &g)

	go s.galgameRepo.IncrementView(galgameID)

	local := s.galgameRepo.FindLocal(galgameID)
	isLiked, isFavorited := s.interactionRepo.UserInteraction(currentUserID, galgameID)

	platforms, languages, types := s.resourceMetaRepo.FindResourceMetaByGalgame(galgameID)

	ratings := s.buildDetailRatings(ctx, galgameID, currentUserID, g)

	if owner := s.ownerOf(galgameID); owner > 0 {
		g.UserID = owner
	}
	g.Contributor = s.contributorsOf(galgameID)
	users := s.hydrateDetailUsers(ctx, g)
	detail := galgameDetailFromNextMoe(g, users)
	if purchase := dlsite.LinkFor(s.dlsiteLinkTemplate, g.ID, g.Refs["dlsite"]); purchase != "" {
		detail.DlsitePurchaseURL = purchase
		detail.DlsiteCouponURL = s.dlsiteCouponURL
	}
	detail.View = local.View
	detail.ResourceUpdateTime = utils.RFC3339OrEmpty(local.ResourceUpdateTime)
	detail.LikeCount = local.LikeCount
	detail.FavoriteCount = local.FavoriteCount
	detail.ResourcePublishBanned = local.ResourcePublishBanned
	detail.IsOnForum = local.ID != 0
	detail.Indexed = local.Published
	detail.IsLiked = isLiked
	detail.IsFavorited = isFavorited
	detail.Platform = platforms
	detail.Language = languages
	detail.Type = types
	detail.Ratings = ratings
	agg := s.listRepo.BayesianRatings([]int{galgameID})[galgameID]
	detail.Rating = agg.Score
	detail.RatingCount = agg.Count
	s.hydrateCoverVotes(ctx, galgameID, token, detail.Covers)
	detail.MyPlaytime = s.hydrateMyPlaytime(ctx, galgameID, token)
	if isSFW {
		detail.Tag = withoutSexualTags(detail.Tag)
	}
	return &detail, nil
}

func (s *GalgameService) contributorsOf(galgameID int) []dto.NextMoeContributor {
	rows := s.contributorRepo.FindContributors(galgameID, contributorMaxPerGalgame)
	out := make([]dto.NextMoeContributor, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.NextMoeContributor{UserID: int(row.UserID)})
	}
	return out
}

func (s *GalgameService) hydrateDetailUsers(ctx context.Context, g dto.NextMoeGalgameDetailFull) map[string]dto.NextMoeUser {
	uids := make([]int, 0, len(g.Contributor)+1)
	uids = append(uids, g.UserID)
	for _, c := range g.Contributor {
		uids = append(uids, c.UserID)
	}
	umap := s.userClient.Hydrate(ctx, uids)
	users := make(map[string]dto.NextMoeUser, len(umap))
	for id, u := range umap {
		users[strconv.Itoa(id)] = dto.NextMoeUser{ID: u.ID, Name: u.Name, Avatar: u.Avatar}
	}
	return users
}

func (s *GalgameService) buildDetailRatings(
	ctx context.Context,
	galgameID, currentUserID int,
	g dto.NextMoeGalgameDetailFull,
) []dto.GalgameDetailRating {
	rows := s.detailRatingRepo.FindRatingsByGalgame(galgameID)
	if len(rows) == 0 {
		return []dto.GalgameDetailRating{}
	}

	userIDs := make([]int, len(rows))
	ratingIDs := make([]int, len(rows))
	for i, r := range rows {
		userIDs[i] = r.UserID
		ratingIDs[i] = r.ID
	}
	userMap := s.userClient.Hydrate(ctx, userIDs)
	likedSet := s.detailRatingRepo.FindLikedRatingIDs(currentUserID, ratingIDs)

	out := make([]dto.GalgameDetailRating, 0, len(rows))
	for _, r := range rows {
		u := userMap[r.UserID]
		if !userclient.IsRenderable(u) {
			continue
		}
		out = append(out, detailRatingFromRow(r, u, likedSet[r.ID], galgameID, g))
	}
	return out
}

func (s *GalgameService) GetList(
	ctx context.Context,
	req *dto.GalgameListRequest,
	isSFW bool,
) (*dto.GalgameListPage, *errors.AppError) {
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}

	releasedFrom, err := utils.ParseReleaseLowerBound(req.ReleasedFrom)
	if err != nil {
		return nil, errors.ErrBadRequest(err.Error())
	}
	releasedTo, err := utils.ParseReleaseUpperBound(req.ReleasedTo)
	if err != nil {
		return nil, errors.ErrBadRequest(err.Error())
	}
	releasedMonths, err := utils.ParseMonthSet(req.ReleasedMonths)
	if err != nil {
		return nil, errors.ErrBadRequest(err.Error())
	}

	filter := model.GalgameListFilter{
		Type:                 req.Type,
		Language:             req.Language,
		Platform:             req.Platform,
		GameType:             req.GameType,
		SortField:            req.SortField,
		SortOrder:            sortOrder,
		IncludeProviders:     splitCSV(req.IncludeProviders),
		ExcludeOnlyProviders: splitCSV(req.ExcludeOnlyProviders),
		ReleasedFrom:         releasedFrom,
		ReleasedTo:           releasedTo,
		ReleasedMonths:       releasedMonths,
		MinRatingCount:       req.MinRatingCount,
		MinRating:            req.MinRating,
		ShowNoResource:       req.ShowNoResource,
		Indexed:              req.Indexed,
		Page:                 req.Page,
		Limit:                req.Limit,
	}

	if req.Library {
		return s.catalogLibrary(ctx, req, releasedFrom, releasedTo, isSFW)
	}

	return s.hydrateListCards(ctx, filter, isSFW)
}

func (s *GalgameService) hydrateListCards(
	ctx context.Context,
	filter model.GalgameListFilter,
	isSFW bool,
) (*dto.GalgameListPage, *errors.AppError) {
	if len(filter.RestrictIDs) > 0 && !entityUsesLocalList(filter) {
		return s.hydrateIDPage(ctx, filter.RestrictIDs, filter.Page, filter.Limit, isSFW)
	}
	ids, total := s.listRepo.ListIDs(filter)
	if len(ids) == 0 {
		return &dto.GalgameListPage{Galgames: []dto.GalgameListCard{}, Total: total}, nil
	}
	cards, appErr := s.HydrateCardsByIDs(ctx, ids, isSFW)
	if appErr != nil {
		return nil, appErr
	}
	return &dto.GalgameListPage{Galgames: cards, Total: total}, nil
}

// Counts a taxonomy sub-set the way hydrateListCards lists it, or the chip ends
// up counting resource-carrying rows against a list of every catalog member.
func (s *GalgameService) countMembers(filter model.GalgameListFilter) int64 {
	if len(filter.RestrictIDs) > 0 && !entityUsesLocalList(filter) {
		return int64(len(filter.RestrictIDs))
	}
	filter.Page, filter.Limit = 1, 1
	_, total := s.listRepo.ListIDs(filter)
	return total
}

func (s *GalgameService) HydrateCardsByIDs(
	ctx context.Context,
	ids []int,
	isSFW bool,
) ([]dto.GalgameListCard, *errors.AppError) {
	if len(ids) == 0 {
		return []dto.GalgameListCard{}, nil
	}

	briefMap, appErr := s.galgameClient.GetBatchPublic(ctx, ids, isSFW)
	if appErr != nil {
		return nil, appErr
	}

	localMap := s.galgameRepo.FindLocalBatch(ids)

	userMap := s.userClient.Hydrate(ctx, frozenCreatorIDs(ids, localMap))

	ratingMap := s.listRepo.BayesianRatings(ids)

	metaRows := s.resourceMetaRepo.FindResourceMetaBatch(ids)
	platformMap, languageMap := groupResourceMeta(metaRows)

	cards := make([]dto.GalgameListCard, 0, len(ids))
	for _, id := range ids {
		b, ok := briefMap[id]
		if !ok {
			continue
		}
		cards = append(cards, dto.GalgameListCard{
			ID:                         id,
			Name:                       b.Name,
			NameOriginal:               b.NameOriginal,
			User:                       frozenCreatorBrief(localMap[id], userMap),
			ContentLimit:               b.ContentLimit,
			View:                       localMap[id].View,
			LikeCount:                  localMap[id].LikeCount,
			Rating:                     ratingMap[id].Score,
			RatingCount:                ratingMap[id].Count,
			ResourceUpdateTime:         utils.RFC3339OrEmpty(localMap[id].ResourceUpdateTime),
			ReleaseDate:                b.ReleaseDate,
			ReleaseDateTBA:             b.ReleaseDateTBA,
			EffectiveBannerHash:        b.EffectiveBannerHash,
			EffectiveBannerURL:         b.EffectiveBannerURL,
			EffectiveBannerWidth:       b.EffectiveBannerWidth,
			EffectiveBannerHeight:      b.EffectiveBannerHeight,
			EffectiveBannerThumbhash:   b.EffectiveBannerThumbhash,
			EffectivePortraitHash:      b.EffectivePortraitHash,
			EffectivePortraitURL:       b.EffectivePortraitURL,
			EffectivePortraitWidth:     b.EffectivePortraitWidth,
			EffectivePortraitHeight:    b.EffectivePortraitHeight,
			EffectivePortraitThumbhash: b.EffectivePortraitThumbhash,
			Platform:                   emptyStrSliceIfNil(platformMap[id]),
			Language:                   emptyStrSliceIfNil(languageMap[id]),
		})
	}
	return cards, nil
}
