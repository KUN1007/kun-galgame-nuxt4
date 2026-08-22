package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kun-galgame-api/pkg/newsclient"
)

func cst(y int, m time.Month, d, h int) time.Time {
	return time.Date(y, m, d, h, 0, 0, 0, time.FixedZone("CST", 8*60*60))
}

// The 2025-01-01 00:00 item is the point of the fixture: published_before is
// inclusive upstream, so a boundary that is not pulled back by a microsecond
// files it under 2024 and both years come out wrong.
var fixture = []time.Time{
	cst(2026, time.March, 5, 9),
	cst(2026, time.January, 10, 12),
	cst(2025, time.July, 1, 8),
	cst(2025, time.January, 1, 0),
	cst(2023, time.November, 11, 20),
}

func newFakeNews(t *testing.T, calls *int) *newsclient.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*calls++
		upper := time.Now().Add(time.Hour)
		if raw := r.URL.Query().Get("published_before"); raw != "" {
			parsed, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				t.Errorf("published_before is not RFC3339: %q", raw)
			}
			upper = parsed
		}
		lower := time.Time{}
		if raw := r.URL.Query().Get("published_after"); raw != "" {
			lower, _ = time.Parse(time.RFC3339, raw)
		}

		var count int64
		var newest time.Time
		for _, at := range fixture {
			if at.Before(lower) || at.After(upper) {
				continue
			}
			count++
			if at.After(newest) {
				newest = at
			}
		}
		items := []map[string]any{}
		if count > 0 {
			items = append(items, map[string]any{
				"id":           1,
				"published_at": newest.Format(time.RFC3339Nano),
				"source":       map[string]any{"key": "hihyou"},
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"items": items, "count": count},
		})
	}))
	t.Cleanup(srv.Close)
	return newsclient.New(newsclient.Config{BaseURL: srv.URL, APIKey: "k"})
}

func TestYearsWalksBackAndStops(t *testing.T) {
	calls := 0
	svc := NewArchiveService(newFakeNews(t, &calls))

	years, err := svc.Years(context.Background(), ArchiveFilter{})
	if err != nil {
		t.Fatalf("Years: %v", err)
	}

	want := []struct {
		year  int
		count int64
	}{{2026, 2}, {2025, 2}, {2024, 0}, {2023, 1}}
	if len(years) != len(want) {
		t.Fatalf("got %d years %+v, want %d", len(years), years, len(want))
	}
	for i, w := range want {
		if years[i].Year != w.year || years[i].Count != w.count {
			t.Errorf("years[%d] = %+v, want %d/%d", i, years[i], w.year, w.count)
		}
	}

	// One head read plus one boundary per year listed; 2022 is never probed
	// because 2023 leaves nothing older.
	if calls != 1+len(want) {
		t.Errorf("made %d upstream calls, want %d", calls, 1+len(want))
	}

	before := calls
	if _, err := svc.Years(context.Background(), ArchiveFilter{}); err != nil {
		t.Fatalf("Years (cached): %v", err)
	}
	if calls != before {
		t.Errorf("cached read made %d more calls", calls-before)
	}
}

func TestMonthsBucketsOneYear(t *testing.T) {
	calls := 0
	svc := NewArchiveService(newFakeNews(t, &calls))

	months, err := svc.Months(context.Background(), ArchiveFilter{}, 2025)
	if err != nil {
		t.Fatalf("Months: %v", err)
	}
	if len(months) != 12 {
		t.Fatalf("got %d months, want 12", len(months))
	}
	for _, m := range months {
		var want int64
		if m.Month == 1 || m.Month == 7 {
			want = 1
		}
		if m.Count != want {
			t.Errorf("month %d = %d, want %d", m.Month, m.Count, want)
		}
	}
}

func TestWindowBounds(t *testing.T) {
	after, before := Window(2025, 0)
	if got := after.Format(time.RFC3339); got != "2025-01-01T00:00:00+08:00" {
		t.Errorf("year after = %s", got)
	}
	if got := before.Format(time.RFC3339Nano); got != "2025-12-31T23:59:59.999999+08:00" {
		t.Errorf("year before = %s", got)
	}

	after, before = Window(2025, 12)
	if got := after.Format(time.RFC3339); got != "2025-12-01T00:00:00+08:00" {
		t.Errorf("month after = %s", got)
	}
	if got := before.Format(time.RFC3339Nano); got != "2025-12-31T23:59:59.999999+08:00" {
		t.Errorf("month before = %s", got)
	}

	if after, before = Window(0, 5); !after.IsZero() || !before.IsZero() {
		t.Errorf("a month without a year should not bound anything, got %s..%s", after, before)
	}
}
