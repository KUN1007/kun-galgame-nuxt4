package handler

import (
	"net/http/httptest"
	"testing"

	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

// library=true is the only thing that sends /galgame to the catalog, so a
// binder that quietly dropped it would answer 200 with the resource list —
// the failure mode is wrong data, never an error.
func TestGalgameListRequest_BindsTheLibraryFlag(t *testing.T) {
	for name, tc := range map[string]struct {
		query string
		want  bool
	}{
		"absent":       {"?page=1&limit=24", false},
		"library=true": {"?page=1&limit=24&library=true", true},
		"library=1":    {"?page=1&limit=24&library=1", true},
	} {
		t.Run(name, func(t *testing.T) {
			var got dto.GalgameListRequest
			app := fiber.New()
			app.Get("/galgame", func(c fiber.Ctx) error {
				if appErr := utils.ParseQueryAndValidate(c, &got); appErr != nil {
					return c.SendStatus(fiber.StatusBadRequest)
				}
				return c.SendStatus(fiber.StatusOK)
			})

			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/galgame"+tc.query, nil))
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			if resp.StatusCode != fiber.StatusOK {
				t.Fatalf("status = %d, want 200", resp.StatusCode)
			}
			if got.Library != tc.want {
				t.Errorf("Library = %v, want %v", got.Library, tc.want)
			}
		})
	}
}
