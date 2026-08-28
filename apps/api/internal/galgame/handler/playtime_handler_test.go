package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/service"
	"kun-galgame-api/internal/middleware"
	"kun-galgame-api/pkg/catalogclient"
	"kun-galgame-api/pkg/errors"

	"github.com/gofiber/fiber/v3"
)

type fakePlaytimeFace struct {
	mu       sync.Mutex
	requests []recordedRequest
	status   int
	body     string
	// What the self-read reports back after the write. A second application of
	// the same user can hold a larger number than the one just written.
	foldMinutes int
	foldClients int
}

func (f *fakePlaytimeFace) server(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &body)
		}
		f.mu.Lock()
		f.requests = append(f.requests, recordedRequest{
			Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery, Body: body,
			Auth: r.Header.Get("Authorization"),
		})
		f.mu.Unlock()

		if f.status != 0 {
			w.WriteHeader(f.status)
			_, _ = w.Write([]byte(f.body))
			return
		}
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"work_id":1000,"minutes":` +
				strconv.Itoa(f.foldMinutes) + `,"status":"finished","last_played_at":null,"clients":` +
				strconv.Itoa(f.foldClients) + `}}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"work_id":1000,"minutes":720,` +
			`"status":"finished","client_id":"forum","updated_at":"2026-08-20T00:00:00Z"}}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func playtimeTestApp(t *testing.T, catalogURL string, user *middleware.UserInfo) *fiber.App {
	t.Helper()
	cc := catalogclient.New(catalogclient.Config{BaseURL: catalogURL})
	gc := client.New(fakeGalgame(t).URL, "nm_test", "")
	h := NewPlaytimeHandler(service.NewPlaytimeService(nil, gc, cc, "forum"))

	app := fiber.New()
	api := app.Group("/api")
	authed := api.Group("", func(c fiber.Ctx) error {
		if user == nil {
			return c.Status(401).JSON(fiber.Map{"code": 205, "message": "用户登录失效"})
		}
		c.Locals(string(middleware.UserInfoKey), user)
		c.Locals(string(middleware.OAuthAccessTokenKey), "user-jwt")
		return c.Next()
	})
	authed.Put("/galgame/:gid/playtime", h.Report)
	return app
}

func TestReportPlaytimeTravelsAsTheUserAndAnswersTheFold(t *testing.T) {
	fake := &fakePlaytimeFace{foldMinutes: 900, foldClients: 2}
	app := playtimeTestApp(t, fake.server(t).URL, plainUser)

	status, raw := doJSON(t, app, "PUT", "/api/galgame/1/playtime",
		`{"minutes":720,"status":"finished"}`)
	if status != http.StatusOK {
		t.Fatalf("report: status = %d body %s", status, raw)
	}

	if len(fake.requests) != 2 {
		t.Fatalf("want a write then a fold read, got %+v", fake.requests)
	}
	write := fake.requests[0]
	if write.Method != http.MethodPut || write.Path != "/v2/me/playtimes/1000" {
		t.Errorf("write went to %s %s, want PUT /v2/me/playtimes/1000", write.Method, write.Path)
	}
	if write.Auth != "Bearer user-jwt" {
		t.Errorf("auth = %q, want the user's own bearer, never the s2s pair", write.Auth)
	}
	if write.Body["minutes"] != float64(720) {
		t.Errorf("body = %v", write.Body)
	}
	read := fake.requests[1]
	if read.Method != http.MethodGet || read.Path != "/v2/me/playtimes/1000" {
		t.Errorf("fold read = %s %s", read.Method, read.Path)
	}

	var env struct {
		Data struct {
			Minutes int `json:"minutes"`
			Clients int `json:"clients"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("decode: %v (%s)", err, raw)
	}
	// 900, not the 720 just written: another app of the same user holds more,
	// and echoing our own figure would contradict the number on the page.
	if env.Data.Minutes != 900 || env.Data.Clients != 2 {
		t.Errorf("answered %+v, want the folded 900 over 2 clients", env.Data)
	}
}

func TestReportPlaytimeRejectsBadInputWithoutCallingCatalog(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"negative", `{"minutes":-1}`},
		{"over the ceiling", `{"minutes":60001}`},
		{"unknown status", `{"minutes":600,"status":"finished_all"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakePlaytimeFace{}
			app := playtimeTestApp(t, fake.server(t).URL, plainUser)

			status, raw := doJSON(t, app, "PUT", "/api/galgame/1/playtime", tc.body)
			if status != http.StatusBadRequest {
				t.Fatalf("status = %d body %s", status, raw)
			}
			if len(fake.requests) != 0 {
				t.Fatalf("catalog was called anyway: %+v", fake.requests)
			}
		})
	}
}

func TestReportPlaytimeAsksForReauthOnAScopeDenial(t *testing.T) {
	fake := &fakePlaytimeFace{
		status: http.StatusForbidden,
		body:   `{"code":233,"message":"the access token is missing the playtime:write scope"}`,
	}
	app := playtimeTestApp(t, fake.server(t).URL, plainUser)

	status, raw := doJSON(t, app, "PUT", "/api/galgame/1/playtime", `{"minutes":720,"status":"finished"}`)
	if status != http.StatusForbidden {
		t.Fatalf("status = %d body %s", status, raw)
	}
	var env struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(raw, &env)
	if env.Code != errors.CodeReauthRequired {
		t.Errorf("code = %d, want CodeReauthRequired so the client offers a re-login", env.Code)
	}
	if !strings.Contains(env.Message, "重新登录") {
		t.Errorf("message = %q", env.Message)
	}
}
