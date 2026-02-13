package handler

import (
	"github.com/jeheskielSunloy77/libra-link/internal/application"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/server"
)

type Handlers struct {
	Health          *HealthHandler
	Auth            *AuthHandler
	User            *UserHandler
	Ebook           *EbookHandler
	Share           *ShareHandler
	ReadingProgress *ReadingProgressHandler
	Bookmark        *BookmarkHandler
	Annotation      *AnnotationHandler
	ReaderSettings  *ReaderSettingsHandler
	Sync            *SyncHandler
	OpenAPI         *OpenAPIHandler
}

func NewHandlers(s *server.Server, services *application.Services) *Handlers {
	h := NewHandler(s)

	return &Handlers{
		Health:          NewHealthHandler(h),
		Auth:            NewAuthHandler(h, services.Auth),
		User:            NewUserHandler(h, services.User),
		Ebook:           NewEbookHandler(h, services.Ebook),
		Share:           NewShareHandler(h, services.Share),
		ReadingProgress: NewReadingProgressHandler(h, services.ReadingProgress),
		Bookmark:        NewBookmarkHandler(h, services.Bookmark),
		Annotation:      NewAnnotationHandler(h, services.Annotation),
		ReaderSettings:  NewReaderSettingsHandler(h, services.UserPreferences, services.UserReaderState),
		Sync:            NewSyncHandler(h, services.Sync),
		OpenAPI:         NewOpenAPIHandler(h),
	}
}
