package router

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeUserResourceHandler struct{}

func (fakeUserResourceHandler) GetMany() fiber.Handler {
	return func(c *fiber.Ctx) error { return c.SendString("user-get-many") }
}

func (fakeUserResourceHandler) GetByID() fiber.Handler {
	return func(c *fiber.Ctx) error { return c.SendString("user-get-by-id") }
}

func (fakeUserResourceHandler) Store() fiber.Handler {
	return func(c *fiber.Ctx) error { return c.SendString("user-store") }
}

func (fakeUserResourceHandler) Destroy() fiber.Handler {
	return func(c *fiber.Ctx) error { return c.SendString("user-destroy") }
}

func (fakeUserResourceHandler) Kill() fiber.Handler {
	return func(c *fiber.Ctx) error { return c.SendString("user-kill") }
}

func (fakeUserResourceHandler) Restore() fiber.Handler {
	return func(c *fiber.Ctx) error { return c.SendString("user-restore") }
}

func (fakeUserResourceHandler) Update() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if _, err := uuid.Parse(c.Params("id")); err != nil {
			return c.Status(http.StatusBadRequest).JSON(map[string]any{
				"message": "invalid id provided",
				"errors": []map[string]string{{
					"field": "id",
					"error": "must be a valid uuid",
				}},
			})
		}
		return c.SendString("user-update")
	}
}

func TestUsersCustomRoutesTakePrecedenceOverUserIDRoutes(t *testing.T) {
	app := fiber.New()
	api := app.Group("/api/v1")

	api.Get("/users/preferences", func(c *fiber.Ctx) error { return c.SendString("prefs-get") })
	api.Patch("/users/preferences", func(c *fiber.Ctx) error { return c.SendString("prefs-patch") })
	api.Get("/users/reader-state", func(c *fiber.Ctx) error { return c.SendString("reader-state-get") })
	api.Patch("/users/reader-state", func(c *fiber.Ctx) error { return c.SendString("reader-state-patch") })

	resource(api, "/users", fakeUserResourceHandler{})

	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
		wantBody   string
	}{
		{name: "patch reader state hits custom route", method: http.MethodPatch, target: "/api/v1/users/reader-state", wantStatus: http.StatusOK, wantBody: "reader-state-patch"},
		{name: "get reader state hits custom route", method: http.MethodGet, target: "/api/v1/users/reader-state", wantStatus: http.StatusOK, wantBody: "reader-state-get"},
		{name: "patch preferences hits custom route", method: http.MethodPatch, target: "/api/v1/users/preferences", wantStatus: http.StatusOK, wantBody: "prefs-patch"},
		{name: "get preferences hits custom route", method: http.MethodGet, target: "/api/v1/users/preferences", wantStatus: http.StatusOK, wantBody: "prefs-get"},
		{name: "patch valid uuid hits resource update", method: http.MethodPatch, target: "/api/v1/users/9fd6f3d6-39ff-4b8e-9e68-a8bd96d96d2c", wantStatus: http.StatusOK, wantBody: "user-update"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.target, strings.NewReader("{}"))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, resp.StatusCode)

			body := readBody(t, resp)
			require.Equal(t, tc.wantBody, strings.TrimSpace(body))
		})
	}
}

func TestUsersInvalidIDStillReturnsUUIDValidationError(t *testing.T) {
	app := fiber.New()
	api := app.Group("/api/v1")

	api.Get("/users/preferences", func(c *fiber.Ctx) error { return c.SendString("prefs-get") })
	api.Patch("/users/preferences", func(c *fiber.Ctx) error { return c.SendString("prefs-patch") })
	api.Get("/users/reader-state", func(c *fiber.Ctx) error { return c.SendString("reader-state-get") })
	api.Patch("/users/reader-state", func(c *fiber.Ctx) error { return c.SendString("reader-state-patch") })

	resource(api, "/users", fakeUserResourceHandler{})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/not-a-uuid", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var payload struct {
		Message string `json:"message"`
		Errors  []struct {
			Field string `json:"field"`
			Error string `json:"error"`
		} `json:"errors"`
	}

	err = json.Unmarshal([]byte(readBody(t, resp)), &payload)
	require.NoError(t, err)
	require.Equal(t, "invalid id provided", payload.Message)
	require.NotEmpty(t, payload.Errors)
	require.Equal(t, "id", payload.Errors[0].Field)
	require.Equal(t, "must be a valid uuid", payload.Errors[0].Error)
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(buf)
}
