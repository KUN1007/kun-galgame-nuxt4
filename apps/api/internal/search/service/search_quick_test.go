package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"kun-galgame-api/internal/galgame/client"
	galgameService "kun-galgame-api/internal/galgame/service"
)

func TestQuickSearch_RequiresAKeyword(t *testing.T) {
	svc := NewSearchService(nil, nil, nil, nil, nil, nil)

	if _, appErr := svc.QuickSearch(context.Background(), "   "); appErr == nil {
		t.Fatal("QuickSearch accepted a blank query")
	}
}

func TestQuickSearch_AnswersEvenWhenEveryLaneBreaks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	// One failure mode per lane: a nil repository panics the topic lane, the 500
	// fails the catalog lane, and a nil user client errors the user lane. The
	// palette still has to get an answer back.
	svc := NewSearchService(
		nil,
		client.New(srv.URL, "nm_test_key", ""),
		&galgameService.GalgameEnricher{},
		nil,
		nil,
		nil,
	)

	res, appErr := svc.QuickSearch(context.Background(), "汉化")
	if appErr != nil {
		t.Fatalf("QuickSearch: %v", appErr)
	}
	if len(res.Topics) != 0 || len(res.Galgames) != 0 || len(res.Users) != 0 {
		t.Errorf(
			"want every lane empty, got %d topics, %d galgames, %d users",
			len(res.Topics), len(res.Galgames), len(res.Users),
		)
	}
	if res.Totals.Topic|res.Totals.Galgame|res.Totals.User != 0 {
		t.Errorf("want zero totals, got %+v", res.Totals)
	}
}
