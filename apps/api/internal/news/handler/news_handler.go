package handler

import (
	"context"
	stderrors "errors"
	"log/slog"
	"slices"

	"kun-galgame-api/internal/news/dto"
	"kun-galgame-api/internal/news/service"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/newsclient"
	"kun-galgame-api/pkg/response"
	"kun-galgame-api/pkg/userclient"
	"kun-galgame-api/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

const (
	defaultLimit = 20
	// The month archive is read as pages of fifty rather than the infinite
	// scroll the feed uses: it is a reference page, and a reader who wants the
	// third week of a month should not have to scroll the first two.
	monthPageSize = 50
)

type NewsHandler struct {
	news    *newsclient.Client
	user    *userclient.Client
	archive *service.ArchiveService
	month   *service.MonthService
}

func NewNewsHandler(news *newsclient.Client, user *userclient.Client) *NewsHandler {
	return &NewsHandler{
		news:    news,
		user:    user,
		archive: service.NewArchiveService(news),
		month:   service.NewMonthService(news),
	}
}

type feedRequest struct {
	Lane   string `query:"lane" validate:"omitempty,oneof=news column"`
	Source string `query:"source"`
	Cursor string `query:"cursor"`
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=50"`
	Year   int    `query:"year" validate:"omitempty,min=1970,max=9999"`
	Month  int    `query:"month" validate:"omitempty,min=1,max=12"`
}

func (h *NewsHandler) GetFeed(c fiber.Ctx) error {
	var req feedRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	if req.Limit == 0 {
		req.Limit = defaultLimit
	}

	after, before := service.Window(req.Year, req.Month)
	feed, err := h.news.Feed(c.Context(), newsclient.FeedQuery{
		Lane:            req.Lane,
		Source:          req.Source,
		Cursor:          req.Cursor,
		Limit:           req.Limit,
		PublishedAfter:  after,
		PublishedBefore: before,
	})
	if err != nil {
		return response.Error(c, feedError(err))
	}
	return response.OK(c, h.mapFeed(c.Context(), feed))
}

// GetSources answers the partner directory. It is a separate read from the feed
// on purpose: the feed only names the partners whose items are on the page you
// asked for, so a filter built from the feed silently loses whichever partner
// has not published lately.
func (h *NewsHandler) GetSources(c fiber.Ctx) error {
	sources, err := h.news.Sources(c.Context())
	if err != nil {
		return response.Error(c, feedError(err))
	}

	upstream := make(map[string]newsclient.Source, len(sources))
	for _, src := range sources {
		upstream[src.Key] = src
	}
	publishers := h.publishers(c.Context(), upstream)

	out := make([]dto.NewsSource, 0, len(sources))
	for _, src := range sources {
		out = append(out, newsSource(src, publishers[src.PublisherUID]))
	}
	return response.OK(c, out)
}

type archiveRequest struct {
	Lane   string `query:"lane" validate:"omitempty,oneof=news column"`
	Source string `query:"source"`
	Year   int    `query:"year" validate:"omitempty,min=1970,max=9999"`
}

// GetArchive answers the year index, plus the month breakdown of the one year
// the caller named. Both are counted under the caller's lane and source, so the
// numbers always describe the list the reader is actually looking at.
func (h *NewsHandler) GetArchive(c fiber.Ctx) error {
	var req archiveRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}

	filter := service.ArchiveFilter{Lane: req.Lane, Source: req.Source}
	years, err := h.archive.Years(c.Context(), filter)
	if err != nil {
		return response.Error(c, feedError(err))
	}

	out := dto.NewsArchive{Years: years, Months: []dto.NewsArchiveMonth{}}
	known := slices.ContainsFunc(years, func(y dto.NewsArchiveYear) bool { return y.Year == req.Year })
	if known {
		months, err := h.archive.Months(c.Context(), filter, req.Year)
		if err != nil {
			return response.Error(c, feedError(err))
		}
		out.Months = months
	}
	return response.OK(c, out)
}

type monthRequest struct {
	Year   int    `query:"year" validate:"required,min=1970,max=9999"`
	Month  int    `query:"month" validate:"required,min=1,max=12"`
	Lane   string `query:"lane" validate:"omitempty,oneof=news column"`
	Source string `query:"source"`
	Day    int    `query:"day" validate:"omitempty,min=1,max=31"`
	Page   int    `query:"page" validate:"omitempty,min=1"`
}

// GetMonth answers one calendar month as a numbered page, plus how many items
// each day of that month holds.
func (h *NewsHandler) GetMonth(c fiber.Ctx) error {
	var req monthRequest
	if appErr := utils.ParseQueryAndValidate(c, &req); appErr != nil {
		return response.Error(c, appErr)
	}
	if req.Page == 0 {
		req.Page = 1
	}

	filter := service.ArchiveFilter{Lane: req.Lane, Source: req.Source}
	items, err := h.month.Items(c.Context(), filter, req.Year, req.Month)
	if err != nil {
		return response.Error(c, feedError(err))
	}

	out := dto.NewsMonth{
		Days:  service.DayCounts(items, req.Year, req.Month),
		Total: len(items),
		Page:  req.Page,
		Limit: monthPageSize,
	}
	if req.Day > 0 {
		items = service.OnDay(items, req.Day)
	}
	out.Count = len(items)

	start := min((req.Page-1)*monthPageSize, len(items))
	end := min(start+monthPageSize, len(items))
	out.Items, out.Sources = h.mapItems(c.Context(), items[start:end])
	return response.OK(c, out)
}

func newsSource(src newsclient.Source, publisher *dto.UserBrief) dto.NewsSource {
	return dto.NewsSource{
		Key:         src.Key,
		Name:        src.DisplayName,
		HomepageURL: src.HomepageURL,
		ColumnURL:   src.ColumnURL,
		Attribution: src.Attribution,
		Publisher:   publisher,
	}
}

func (h *NewsHandler) mapFeed(ctx context.Context, feed *newsclient.Feed) dto.NewsFeed {
	items, sources := h.mapItems(ctx, feed.Items)
	return dto.NewsFeed{
		Items:      items,
		Sources:    sources,
		Count:      feed.Count,
		NextCursor: feed.NextCursor,
	}
}

func (h *NewsHandler) mapItems(ctx context.Context, in []newsclient.Item) ([]dto.NewsItem, map[string]dto.NewsSource) {
	upstream := make(map[string]newsclient.Source)
	items := make([]dto.NewsItem, 0, len(in))
	for _, it := range in {
		items = append(items, dto.NewsItem{
			ID:          it.ID,
			SourceKey:   it.Source.Key,
			Lane:        it.Lane,
			Title:       it.Title,
			Preview:     it.Preview,
			SourceURL:   it.SourceURL,
			BannerURL:   it.BannerURL,
			PublishedAt: it.PublishedAt,
		})
		if _, ok := upstream[it.Source.Key]; !ok {
			upstream[it.Source.Key] = it.Source
		}
	}

	publishers := h.publishers(ctx, upstream)
	sources := make(map[string]dto.NewsSource, len(upstream))
	for key, src := range upstream {
		sources[key] = newsSource(src, publishers[src.PublisherUID])
	}
	return items, sources
}

// publishers hydrates the forum account each partner publishes under, which is
// what news_source.publisher_uid names. A miss leaves the source without one
// rather than failing the page: the display name and the attribution text stand
// on their own, and an unreachable OAuth is not worth losing the feed over.
func (h *NewsHandler) publishers(ctx context.Context, sources map[string]newsclient.Source) map[int64]*dto.UserBrief {
	out := make(map[int64]*dto.UserBrief, len(sources))
	if h.user == nil {
		return out
	}
	ids := make([]int, 0, len(sources))
	for _, src := range sources {
		if src.PublisherUID > 0 {
			ids = append(ids, int(src.PublisherUID))
		}
	}
	if len(ids) == 0 {
		return out
	}
	users, err := h.user.Users(ctx, ids)
	if err != nil {
		slog.Warn("news: 获取情报发布者信息失败", "error", err)
		return out
	}
	for id, u := range users {
		if !userclient.IsRenderable(u) {
			continue
		}
		out[int64(id)] = &dto.UserBrief{ID: u.ID, Name: u.Name, Avatar: u.Avatar}
	}
	return out
}

func feedError(err error) *errors.AppError {
	switch {
	case stderrors.Is(err, newsclient.ErrNotConfigured):
		return errors.New(errors.CodeBiz, "情报服务未配置", fiber.StatusServiceUnavailable)
	case stderrors.Is(err, newsclient.ErrBadRequest):
		return errors.ErrBadRequest("情报服务拒绝了该查询")
	}
	slog.Error("news: 情报服务请求失败", "error", err)
	return errors.New(errors.CodeBiz, "情报服务暂时不可用", fiber.StatusBadGateway)
}
