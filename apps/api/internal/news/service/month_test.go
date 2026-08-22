package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"kun-galgame-api/pkg/newsclient"
)

var monthFixture = []time.Time{
	cst(2025, time.March, 30, 9),
	cst(2025, time.March, 30, 8),
	cst(2025, time.March, 16, 20),
	cst(2025, time.March, 16, 19),
	cst(2025, time.March, 2, 11),
	cst(2025, time.April, 1, 0),
	cst(2025, time.February, 28, 23),
}

// The fake hands out three items at a time whatever limit is asked for, because
// what the walk has to get right is following next_cursor to the end — not the
// page size it happens to be served.
func newPagedNews(t *testing.T, calls *int) *newsclient.Client {
	t.Helper()
	const pageSize = 3
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*calls++
		q := r.URL.Query()
		lower, _ := time.Parse(time.RFC3339, q.Get("published_after"))
		upper, err := time.Parse(time.RFC3339, q.Get("published_before"))
		if err != nil {
			t.Errorf("published_before is not RFC3339: %q", q.Get("published_before"))
		}

		var matched []time.Time
		for _, at := range monthFixture {
			if !at.Before(lower) && !at.After(upper) {
				matched = append(matched, at)
			}
		}
		offset, _ := strconv.Atoi(q.Get("cursor"))
		end := min(offset+pageSize, len(matched))

		items := []map[string]any{}
		for i := offset; i < end; i++ {
			items = append(items, map[string]any{
				"id":           i + 1,
				"published_at": matched[i].Format(time.RFC3339Nano),
				"source":       map[string]any{"key": "hihyou"},
			})
		}
		cursor := ""
		if end < len(matched) {
			cursor = strconv.Itoa(end)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"items": items, "count": len(matched), "next_cursor": cursor},
		})
	}))
	t.Cleanup(srv.Close)
	return newsclient.New(newsclient.Config{BaseURL: srv.URL, APIKey: "k"})
}

func TestItemsWalksEveryPage(t *testing.T) {
	calls := 0
	svc := NewMonthService(newPagedNews(t, &calls))

	items, err := svc.Items(context.Background(), ArchiveFilter{}, 2025, 3)
	if err != nil {
		t.Fatalf("Items: %v", err)
	}
	// Five of the seven fixture items fall inside March in CST; the 1 April
	// 00:00 one is the guard, since UTC would file it under March.
	if len(items) != 5 {
		t.Fatalf("got %d items, want 5", len(items))
	}
	if calls != 2 {
		t.Errorf("made %d upstream calls, want 2 pages", calls)
	}

	before := calls
	if _, err := svc.Items(context.Background(), ArchiveFilter{}, 2025, 3); err != nil {
		t.Fatalf("Items (cached): %v", err)
	}
	if calls != before {
		t.Errorf("cached read made %d more calls", calls-before)
	}

	days := DayCounts(items, 2025, 3)
	if len(days) != 31 {
		t.Fatalf("got %d days, want 31", len(days))
	}
	for _, d := range days {
		want := 0
		switch d.Day {
		case 2:
			want = 1
		case 16, 30:
			want = 2
		}
		if d.Count != want {
			t.Errorf("day %d = %d, want %d", d.Day, d.Count, want)
		}
	}

	if got := OnDay(items, 16); len(got) != 2 {
		t.Errorf("OnDay(16) = %d items, want 2", len(got))
	}
	if got := OnDay(items, 3); len(got) != 0 {
		t.Errorf("OnDay(3) = %d items, want 0", len(got))
	}
}

func TestDayCountsSpansTheWholeMonth(t *testing.T) {
	for _, tc := range []struct{ year, month, days int }{
		{2025, 2, 28}, {2024, 2, 29}, {2025, 4, 30}, {2025, 12, 31},
	} {
		if got := len(DayCounts(nil, tc.year, tc.month)); got != tc.days {
			t.Errorf("%d-%02d has %d days, want %d", tc.year, tc.month, got, tc.days)
		}
	}
}
