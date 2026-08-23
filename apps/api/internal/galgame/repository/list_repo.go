package repository

import (
	"math"
	"strconv"
	"strings"

	"kun-galgame-api/internal/galgame/model"

	"gorm.io/gorm"
)

const bayesianPriorC = 10.0

const ratingAggJoin = "LEFT JOIN (SELECT galgame_id, SUM(overall) AS rsum, " +
	"COUNT(*) AS rcnt FROM galgame_rating GROUP BY galgame_id) rt " +
	"ON rt.galgame_id = g.id"

type GalgameListRepository struct {
	db *gorm.DB
}

func NewGalgameListRepository(db *gorm.DB) *GalgameListRepository {
	return &GalgameListRepository{db: db}
}

var allProviders = []string{
	"baidu", "aliyun", "quark", "pan123", "tianyiyun",
	"caiyun", "xunlei", "uc", "lanzou", "other",
}

func (r *GalgameListRepository) ListIDs(f model.GalgameListFilter) (ids []int, total int64) {
	sortCol := "g.resource_update_time"
	viewOneDay := "COALESCE((SELECT SUM(d.count) FROM galgame_view_daily d " +
		"WHERE d.entity_id = g.id AND d.day = CURRENT_DATE), 0)"
	switch f.SortField {
	case "time":
		sortCol = "g.resource_update_time"
	case "created":
		sortCol = "g.created"
	case "view":
		sortCol = "g.view"
	case "view_1d":
		sortCol = viewOneDay
	case "view_7d":
		sortCol = "g.view_7d"
	case "view_30d":
		sortCol = "g.view_30d"
	case "release_date":
		sortCol = "g.release_date"
	}
	isSubquerySort := f.SortField == "view_1d"

	ratingSort := f.SortField == "rating"
	ratingFilter := f.MinRatingCount > 0 || f.MinRating > 0

	var bayes string
	if ratingSort || ratingFilter {
		bayes = r.bayesianExpr()
	}

	orderClause := sortCol + " " + f.SortOrder
	switch {
	case ratingSort:
		orderClause = "(rt.rcnt IS NULL), " + bayes + " " + f.SortOrder
	case sortCol == "g.release_date":
		orderClause += " NULLS LAST"
	}

	type idRow struct {
		ID int `gorm:"column:id"`
	}

	if !f.HasResourcePredicate() {
		build := func() *gorm.DB {
			q := applyIndexability(r.db.Table("galgame g"), f)
			if f.RestrictIDs != nil {
				q = q.Where("g.id = ANY(?::int[])", intArrayLit(f.RestrictIDs))
			}
			if ratingSort || ratingFilter {
				q = q.Joins(ratingAggJoin)
			}
			q = applyReleaseFilter(q, f)
			q = applyGameTypeFilter(q, f)
			if ratingFilter {
				q = applyRatingFilter(q, f, bayes)
			}
			if !f.Indexed && !f.ShowNoResource {
				q = q.Where("EXISTS (SELECT 1 FROM galgame_resource gr WHERE gr.galgame_id = g.id)")
			}
			return q
		}
		build().Select("COUNT(*)").Scan(&total)
		var rows []idRow
		build().
			Select("g.id").
			Order(orderClause).
			Offset((f.Page - 1) * f.Limit).Limit(f.Limit).
			Scan(&rows)
		ids = make([]int, len(rows))
		for i, row := range rows {
			ids[i] = row.ID
		}
		return
	}

	inner := applyIndexability(r.db.Table("galgame g").
		Select("DISTINCT g.id").
		Joins("JOIN galgame_resource gr ON gr.galgame_id = g.id"), f)
	if f.RestrictIDs != nil {
		inner = inner.Where("g.id = ANY(?::int[])", intArrayLit(f.RestrictIDs))
	}
	if ratingFilter {
		inner = inner.Joins(ratingAggJoin)
	}

	inner = applyReleaseFilter(inner, f)
	inner = applyGameTypeFilter(inner, f)
	if f.Type != "" && f.Type != "all" {
		inner = inner.Where("gr.type = ?", f.Type)
	}
	if f.Language != "" && f.Language != "all" {
		inner = inner.Where("gr.language = ?", f.Language)
	}
	if f.Platform != "" && f.Platform != "all" {
		inner = inner.Where("gr.platform = ?", f.Platform)
	}
	if len(f.IncludeProviders) > 0 {
		inner = inner.Where("gr.provider && ?", providerArrayLit(f.IncludeProviders))
	}
	if len(f.ExcludeOnlyProviders) > 0 {
		allowed := providersExcluding(f.ExcludeOnlyProviders)
		if len(allowed) > 0 {
			inner = inner.Where("gr.provider && ?", providerArrayLit(allowed))
		}
	}
	if ratingFilter {
		inner = applyRatingFilter(inner, f, bayes)
	}

	r.db.Table("(?) AS sub", inner).Select("COUNT(*)").Scan(&total)

	main := applyIndexability(r.db.Table("galgame g").
		Select("g.id").
		Joins("JOIN galgame_resource gr ON gr.galgame_id = g.id"), f)
	groupBy := "g.id, " + sortCol
	if isSubquerySort {
		groupBy = "g.id"
	}
	if ratingSort {
		main = main.Joins(ratingAggJoin)
		groupBy = "g.id, rt.rsum, rt.rcnt"
	}

	var rows []idRow
	main.
		Where("gr.galgame_id IN (?)", inner).
		Group(groupBy).
		Order(orderClause).
		Offset((f.Page - 1) * f.Limit).Limit(f.Limit).
		Scan(&rows)

	ids = make([]int, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	return
}

func (r *GalgameListRepository) bayesianExpr() string {
	var m float64
	r.db.Table("galgame_rating").Select("COALESCE(AVG(overall), 0)").Scan(&m)
	c := strconv.FormatFloat(bayesianPriorC, 'f', -1, 64)
	ms := strconv.FormatFloat(m, 'f', 6, 64)
	return "(" + c + " * " + ms + " + COALESCE(rt.rsum, 0)) / (" +
		c + " + COALESCE(rt.rcnt, 0))"
}

type RatingInfo struct {
	Score float64
	Count int
}

func (r *GalgameListRepository) BayesianRatings(ids []int) map[int]RatingInfo {
	out := make(map[int]RatingInfo, len(ids))
	if len(ids) == 0 {
		return out
	}
	var m float64
	r.db.Table("galgame_rating").Select("COALESCE(AVG(overall), 0)").Scan(&m)

	type aggRow struct {
		GalgameID int     `gorm:"column:galgame_id"`
		Rsum      float64 `gorm:"column:rsum"`
		Rcnt      int     `gorm:"column:rcnt"`
	}
	var rows []aggRow
	r.db.Table("galgame_rating").
		Select("galgame_id, SUM(overall) AS rsum, COUNT(*) AS rcnt").
		Where("galgame_id IN ?", ids).
		Group("galgame_id").
		Scan(&rows)

	for _, row := range rows {
		if row.Rcnt == 0 {
			continue
		}
		score := (bayesianPriorC*m + row.Rsum) / (bayesianPriorC + float64(row.Rcnt))
		out[row.GalgameID] = RatingInfo{
			Score: math.Round(score*10) / 10,
			Count: row.Rcnt,
		}
	}
	return out
}

func applyRatingFilter(q *gorm.DB, f model.GalgameListFilter, bayes string) *gorm.DB {
	if f.MinRatingCount > 0 {
		q = q.Where("COALESCE(rt.rcnt, 0) >= ?", f.MinRatingCount)
	}
	if f.MinRating > 0 {
		q = q.Where(bayes+" >= ?", f.MinRating)
	}
	return q
}

func applyReleaseFilter(q *gorm.DB, f model.GalgameListFilter) *gorm.DB {
	if f.ReleasedFrom != "" {
		q = q.Where("g.release_date >= ?", f.ReleasedFrom)
	}
	if f.ReleasedTo != "" {
		q = q.Where("g.release_date <= ?", f.ReleasedTo)
	}
	if len(f.ReleasedMonths) > 0 {
		q = q.Where("EXTRACT(MONTH FROM g.release_date)::int IN ?", f.ReleasedMonths)
	}
	return q
}

func applyGameTypeFilter(q *gorm.DB, f model.GalgameListFilter) *gorm.DB {
	switch f.GameType {
	case "", "all":
		return q
	case "uncategorized":
		return q.Where("NOT EXISTS (SELECT 1 FROM galgame_rating grt WHERE grt.galgame_id = g.id AND grt.galgame_type <> '[]'::jsonb)")
	default:
		return q.Where(
			"EXISTS (SELECT 1 FROM galgame_rating grt WHERE grt.galgame_id = g.id AND grt.galgame_type @> ?)",
			"[\""+f.GameType+"\"]",
		)
	}
}

func applyIndexability(q *gorm.DB, f model.GalgameListFilter) *gorm.DB {
	if f.Indexed {
		return q.Where("g.published")
	}
	return q
}

func providerArrayLit(providers []string) string {
	return "{" + strings.Join(providers, ",") + "}"
}

func intArrayLit(ids []int) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.Itoa(id)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func providersExcluding(excluded []string) []string {
	exSet := map[string]bool{}
	for _, e := range excluded {
		exSet[e] = true
	}
	out := make([]string, 0, len(allProviders))
	for _, p := range allProviders {
		if !exSet[p] {
			out = append(out, p)
		}
	}
	return out
}
