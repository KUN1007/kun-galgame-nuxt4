package catalogclient

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReportPlaytime_ForwardsBearerAndBody(t *testing.T) {
	var gotPath, gotAuth, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotAuth, gotMethod = r.URL.Path, r.Header.Get("Authorization"), r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"work_id":7,"minutes":720,"status":"finished","client_id":"forum","updated_at":"2026-08-20T00:00:00Z"}}`))
	}))
	defer srv.Close()

	// No client_id/secret: the playtime face is a user-plane call and must not
	// need the s2s credentials the rest of catalogclient carries.
	c := New(Config{BaseURL: srv.URL})
	got, err := c.ReportPlaytime(context.Background(), "user-jwt", 7,
		PlaytimeReport{Minutes: 720, Status: PlaytimeStatusFinished})
	if err != nil {
		t.Fatalf("ReportPlaytime: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method = %s, want PUT", gotMethod)
	}
	if gotPath != "/v2/me/playtimes/7" {
		t.Errorf("path = %s, want /v2/me/playtimes/7", gotPath)
	}
	if gotAuth != "Bearer user-jwt" {
		t.Errorf("auth = %q, want the user's bearer", gotAuth)
	}
	if gotBody["minutes"] != float64(720) {
		t.Errorf("body = %v, want the absolute minutes", gotBody)
	}
	if got.Minutes != 720 {
		t.Errorf("record = %+v", got)
	}
}

func TestMyPlaytime_NullPayloadIsNotAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":null}`))
	}))
	defer srv.Close()

	got, err := New(Config{BaseURL: srv.URL}).MyPlaytime(context.Background(), "user-jwt", 7)
	if err != nil {
		t.Fatalf("MyPlaytime: %v", err)
	}
	if got != nil {
		t.Fatalf("got %+v, want nil — never having reported is a 200 with a null payload", got)
	}
}

func TestPlaytime_ScopeDenialIsDistinctFromForbidden(t *testing.T) {
	cases := []struct {
		name string
		body string
		want error
	}{
		{"missing scope", `{"code":233,"message":"the access token is missing the playtime:write scope"}`, ErrInsufficientScope},
		{"plain forbidden", `{"code":233,"message":"banned"}`, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			_, err := New(Config{BaseURL: srv.URL}).ReportPlaytime(
				context.Background(), "user-jwt", 7, PlaytimeReport{Minutes: 60})
			if tc.want != nil {
				if !errors.Is(err, tc.want) {
					t.Fatalf("err = %v, want ErrInsufficientScope", err)
				}
				return
			}
			if errors.Is(err, ErrInsufficientScope) {
				t.Fatalf("a plain 403 must not read as a scope denial: %v", err)
			}
			var apiErr *UserAPIError
			if !errors.As(err, &apiErr) || apiErr.Status != http.StatusForbidden {
				t.Fatalf("err = %v, want a UserAPIError carrying 403", err)
			}
		})
	}
}

func TestListMyPlaytime_PassesCursorAndReturnsNext(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"items":[{"work_id":7,"minutes":120,"status":"playing","client_id":"forum","updated_at":"2026-08-02T00:00:00Z"}],"cursor":"2026-08-02T00:00:00Z"}}`))
	}))
	defer srv.Close()

	items, cursor, err := New(Config{BaseURL: srv.URL}).ListMyPlaytime(
		context.Background(), "user-jwt", "2026-08-01T00:00:00Z", 200)
	if err != nil {
		t.Fatalf("ListMyPlaytime: %v", err)
	}
	if gotQuery != "limit=200&updated_since=2026-08-01T00%3A00%3A00Z" {
		t.Errorf("query = %q", gotQuery)
	}
	if len(items) != 1 || items[0].WorkID != 7 {
		t.Errorf("items = %+v", items)
	}
	if cursor != "2026-08-02T00:00:00Z" {
		t.Errorf("cursor = %q", cursor)
	}
}
