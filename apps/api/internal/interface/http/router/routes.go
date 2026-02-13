package router

import (
	"github.com/gofiber/fiber/v2"
	applicationdto "github.com/jeheskielSunloy77/libra-link/internal/application/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	httpdto "github.com/jeheskielSunloy77/libra-link/internal/interface/http/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/interface/http/handler"
	"github.com/jeheskielSunloy77/libra-link/internal/interface/http/middleware"
)

func registerRoutes(
	r *fiber.App,
	h *handler.Handlers,
	middlewares *middleware.Middlewares,
) {
	// system routes
	r.Get("/health", h.Health.GetHealth)
	r.Static("/static", "static")
	r.Get("/api/docs", h.OpenAPI.ServeOpenAPIUI)

	// versioned routes
	api := r.Group("/api/v1")

	authGroup := api.Group("/auth")
	authGroup.Post("/register", h.Auth.Register())
	authGroup.Post("/login", h.Auth.Login())
	authGroup.Post("/google/device/start", h.Auth.GoogleDeviceStart())
	authGroup.Post("/google/device/poll", h.Auth.GoogleDevicePoll())
	authGroup.Get("/google", h.Auth.GoogleLogin())
	authGroup.Get("/google/callback", h.Auth.GoogleCallback())
	authGroup.Post("/verify-email", h.Auth.VerifyEmail())
	authGroup.Post("/refresh", h.Auth.Refresh())
	authGroup.Post("/logout", h.Auth.Logout())

	authProtected := authGroup.Group("", middlewares.Auth.RequireAuth())
	authProtected.Get("/me", h.Auth.Me())
	authProtected.Post("/resend-verification", h.Auth.ResendVerification())
	authProtected.Post("/logout-all", h.Auth.LogoutAll())

	// protected routes
	protected := api.Group("", middlewares.Auth.RequireAuth())

	resource(protected, "/users", h.User.ResourceHandler)
	resource(protected, "/ebooks", h.Ebook.ResourceHandler)
	resource(protected, "/shares", h.Share.ResourceHandler)
	resource(protected, "/reading-progress", h.ReadingProgress.ResourceHandler)
	resource(protected, "/bookmarks", h.Bookmark.ResourceHandler)
	resource(protected, "/annotations", h.Annotation.ResourceHandler)

	protected.Post("/ebooks/:id/metadata", h.Ebook.AttachMetadata())
	protected.Delete("/ebooks/:id/metadata", h.Ebook.DetachMetadata())

	protected.Post("/shares/:id/borrow", h.Share.Borrow())
	protected.Post("/borrows/:id/return", h.Share.ReturnBorrow())
	protected.Put("/shares/:id/review", h.Share.UpsertReview())
	protected.Post("/shares/:id/report", h.Share.CreateReport())

	protected.Get("/users/preferences", h.ReaderSettings.GetPreferences())
	protected.Patch("/users/preferences", h.ReaderSettings.PatchPreferences())
	protected.Get("/users/reader-state", h.ReaderSettings.GetReaderState())
	protected.Patch("/users/reader-state", h.ReaderSettings.PatchReaderState())

	protected.Post("/sync/events", h.Sync.StoreEvent())
	protected.Get("/sync/events", h.Sync.ListEvents())
}

func resource[T domain.BaseModel, S applicationdto.StoreDTO[T], U applicationdto.UpdateDTO[T], TS httpdto.StoreDTO[S], TU httpdto.UpdateDTO[U]](group fiber.Router, path string, h *handler.ResourceHandler[T, S, U, TS, TU], authMiddleware ...fiber.Handler) {
	g := group.Group(path, authMiddleware...)
	g.Get("/", h.GetMany())
	g.Get("/:id", h.GetByID())
	g.Post("/", h.Store())
	g.Delete("/:id", h.Destroy())
	g.Delete("/:id/kill", h.Kill())
	g.Patch("/:id/restore", h.Restore())
	g.Patch("/:id", h.Update())
}
