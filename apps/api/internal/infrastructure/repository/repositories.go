package repository

import (
	"github.com/jeheskielSunloy77/libra-link/internal/application/port"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/lib/cache"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/server"
)

type Repositories = port.Repositories

func NewRepositories(s *server.Server, cacheClient cache.Cache) *Repositories {
	return &Repositories{
		Auth:              NewAuthRepository(s.DB.DB),
		AuthSession:       NewAuthSessionRepository(s.DB.DB),
		User:              NewUserRepository(s.Config, s.DB.DB, cacheClient),
		EmailVerification: NewEmailVerificationRepository(s.DB.DB),
		Ebook:             NewEbookRepository(s.Config, s.DB.DB, cacheClient),
		EbookMetadata:     NewEbookGoogleMetadataRepository(s.DB.DB),
		UserPreferences:   NewUserPreferencesRepository(s.DB.DB),
		UserReaderState:   NewUserReaderStateRepository(s.DB.DB),
		ReadingProgress:   NewReadingProgressRepository(s.Config, s.DB.DB, cacheClient),
		Bookmark:          NewBookmarkRepository(s.Config, s.DB.DB, cacheClient),
		Annotation:        NewAnnotationRepository(s.Config, s.DB.DB, cacheClient),
		Share:             NewShareRepository(s.Config, s.DB.DB, cacheClient),
		Borrow:            NewBorrowRepository(s.Config, s.DB.DB, cacheClient),
		ShareReview:       NewShareReviewRepository(s.Config, s.DB.DB, cacheClient),
		ShareReport:       NewShareReportRepository(s.Config, s.DB.DB, cacheClient),
		SyncEvent:         NewSyncEventRepository(s.Config, s.DB.DB, cacheClient),
		SyncCheckpoint:    NewSyncCheckpointRepository(s.DB.DB),
	}
}
