package service

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/pkg/errors"
)

const upcomingMonthCap = 24

const upcomingFetchConcurrency = 8

const calendarPageLimit = 100

// v2's calendar face answers with items and a cursor and nothing else — no
// month, no today, no min/max month. Month.vue lays the day grid out from
// `month`, so an empty one drew "0 年 0 月" and put day 1 under whatever
// weekday NaN produced. Reconstruct the two the page cannot do without, in
// the JST the month boundaries are cut on.
var calendarLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.UTC
	}
	return loc
}()

func calendarNow(layout string) string {
	return time.Now().In(calendarLocation).Format(layout)
}

func orElse(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

type CalendarService struct {
	galgameClient *client.GalgameClient
	enricher      *GalgameEnricher
}

func NewCalendarService(galgameClient *client.GalgameClient, enricher *GalgameEnricher) *CalendarService {
	return &CalendarService{galgameClient: galgameClient, enricher: enricher}
}

func bucketQuery(isSFW bool) url.Values {
	q := url.Values{
		"limit":   {strconv.Itoa(calendarPageLimit)},
		"include": {CatalogCardInclude},
	}
	client.ApplyWorksGate(q, isSFW)
	return q
}

func (s *CalendarService) fetchMonthRaw(
	ctx context.Context,
	month string,
	isSFW bool,
) (*client.CatalogWorksPage, *errors.AppError) {
	q := bucketQuery(isSFW)
	if month != "" {
		q.Set("month", month)
	}
	return s.galgameClient.CatalogCalendar(ctx, "", q)
}

func (s *CalendarService) GetMonth(
	ctx context.Context,
	rawQuery url.Values,
	isSFW bool,
) (*dto.CalendarMonthPage, *errors.AppError) {
	month := orElse(rawQuery.Get("month"), calendarNow("2006-01"))
	page, appErr := s.fetchMonthRaw(ctx, month, isSFW)
	if appErr != nil {
		return nil, appErr
	}
	page.Month = orElse(page.Month, month)

	return &dto.CalendarMonthPage{
		Month: page.Month,
		Today: orElse(page.Meta.Today, calendarNow("2006-01-02")),
		Items: s.enricher.ToCards(ctx, catalogItemsToNextMoe(ctx, page.Items)),
		Meta: dto.CalendarMeta{
			PrevMonth: shiftMonth(page.Month, -1),
			NextMonth: shiftMonth(page.Month, +1),
			HasPrev:   derefBool(page.Meta.HasPrev),
			HasNext:   derefBool(page.Meta.HasNext),
			MinMonth:  page.Meta.MinMonth,
			MaxMonth:  page.Meta.MaxMonth,
			Count:     int(page.Count),
		},
	}, nil
}

// GetTodayFlag answers the one boolean the sidebar rail asks of the calendar.
func (s *CalendarService) GetTodayFlag(
	ctx context.Context,
	isSFW bool,
) (*dto.CalendarTodayFlag, *errors.AppError) {
	page, appErr := s.fetchMonthRaw(ctx, "", isSFW)
	if appErr != nil {
		return nil, appErr
	}
	today := orElse(page.Meta.Today, calendarNow("2006-01-02"))

	flag := &dto.CalendarTodayFlag{Today: today, ExpiresIn: secondsUntilCalendarMidnight()}
	for _, item := range page.Items {
		if item.ReleaseDate != nil && *item.ReleaseDate == today {
			flag.HasRelease = true
			break
		}
	}
	return flag, nil
}

func secondsUntilCalendarMidnight() int {
	now := time.Now().In(calendarLocation)
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, calendarLocation)
	return int(midnight.Sub(now).Seconds())
}

func (s *CalendarService) GetPending(
	ctx context.Context,
	rawQuery url.Values,
	isSFW bool,
) (*dto.CalendarPendingPage, *errors.AppError) {
	q := bucketQuery(isSFW)
	if year := rawQuery.Get("year"); year != "" {
		q.Set("year", year)
	}
	page, appErr := s.galgameClient.CatalogCalendar(ctx, "/pending", q)
	if appErr != nil {
		return nil, appErr
	}
	return &dto.CalendarPendingPage{
		Year:  orElse(page.Year, calendarNow("2006")),
		Items: s.enricher.ToCards(ctx, catalogItemsToNextMoe(ctx, page.Items)),
		Count: int(page.Count),
	}, nil
}

func (s *CalendarService) GetTBA(
	ctx context.Context,
	rawQuery url.Values,
	isSFW bool,
) (*dto.CalendarTBAPage, *errors.AppError) {
	page, appErr := s.galgameClient.CatalogCalendar(ctx, "/tba", bucketQuery(isSFW))
	if appErr != nil {
		return nil, appErr
	}
	return &dto.CalendarTBAPage{
		Items: s.enricher.ToCards(ctx, catalogItemsToNextMoe(ctx, page.Items)),
		Count: int(page.Count),
	}, nil
}

func (s *CalendarService) GetUpcoming(
	ctx context.Context,
	isSFW bool,
) (*dto.CalendarUpcomingPage, *errors.AppError) {
	base, appErr := s.fetchMonthRaw(ctx, "", isSFW)
	if appErr != nil {
		return nil, appErr
	}
	today := orElse(base.Meta.Today, calendarNow("2006-01-02"))
	start := orElse(base.Month, calendarNow("2006-01"))
	last := orElse(base.Meta.MaxMonth, shiftMonth(start, upcomingMonthCap))
	months := monthRange(start, last, upcomingMonthCap)

	rawByMonth := make([][]client.CatalogWorkListItem, len(months))
	rawByMonth[0] = base.Items
	if len(months) > 1 {
		var wg sync.WaitGroup
		sem := make(chan struct{}, upcomingFetchConcurrency)
		for i := 1; i < len(months); i++ {
			wg.Add(1)
			go func(idx int, m string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				if resp, err := s.fetchMonthRaw(ctx, m, isSFW); err == nil {
					rawByMonth[idx] = resp.Items
				}
			}(i, months[i])
		}
		wg.Wait()
	}

	var flat []dto.NextMoeGalgameItem
	for _, items := range rawByMonth {
		for i := range items {
			it := &items[i]
			if !client.CatalogItemRenderable(it) {
				continue
			}
			// Lexicographic compare on the partial-ISO date is correct and
			// intentional: a month-precision "2026-08" sorts before every day in
			// that month, so a month-precision entry in the current month stays
			// visible instead of dropping out.
			if it.ReleaseDate != nil && *it.ReleaseDate >= monthOf(today) {
				flat = append(flat, client.CatalogItemToNextMoeItem(ctx, it))
			}
		}
	}

	cards := s.enricher.ToCards(ctx, flat)
	byMonth := make(map[string][]dto.GalgameCard, len(months))
	for _, c := range cards {
		if c.ReleaseDate == nil || len(*c.ReleaseDate) < 7 {
			continue
		}
		key := (*c.ReleaseDate)[:7]
		byMonth[key] = append(byMonth[key], c)
	}

	out := make([]dto.CalendarUpcomingMonth, 0, len(months))
	total := 0
	for _, m := range months {
		if g := byMonth[m]; len(g) > 0 {
			out = append(out, dto.CalendarUpcomingMonth{Month: m, Items: g})
			total += len(g)
		}
	}
	return &dto.CalendarUpcomingPage{Today: today, Months: out, Count: total}, nil
}

func monthOf(today string) string {
	if len(today) < 7 {
		return today
	}
	return today[:7]
}

func derefBool(p *bool) bool {
	return p != nil && *p
}

func shiftMonth(ym string, n int) string {
	y, m := parseYM(ym)
	if y == 0 {
		return ""
	}
	total := y*12 + (m - 1) + n
	return formatYM(total/12, total%12+1)
}

func monthRange(start, end string, cap int) []string {
	sy, sm := parseYM(start)
	ey, em := parseYM(end)
	if ey < sy || (ey == sy && em < sm) {
		return []string{formatYM(sy, sm)}
	}
	out := make([]string, 0, cap)
	y, m := sy, sm
	for len(out) < cap {
		out = append(out, formatYM(y, m))
		if y == ey && m == em {
			break
		}
		if m++; m > 12 {
			m = 1
			y++
		}
	}
	return out
}

func parseYM(s string) (int, int) {
	if len(s) < 7 {
		return 0, 1
	}
	y, _ := strconv.Atoi(s[:4])
	m, _ := strconv.Atoi(s[5:7])
	if m < 1 || m > 12 {
		m = 1
	}
	return y, m
}

func formatYM(y, m int) string {
	return fmt.Sprintf("%04d-%02d", y, m)
}
