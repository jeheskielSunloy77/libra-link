package application

import (
	"github.com/jeheskielSunloy77/libra-link/internal/application/port"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/lib/job"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/server"
)

type Services struct {
	Auth            AuthService
	User            UserService
	Ebook           EbookService
	Share           ShareService
	ReadingProgress ReadingProgressService
	Bookmark        BookmarkService
	Annotation      AnnotationService
	UserPreferences UserPreferencesService
	UserReaderState UserReaderStateService
	Sync            SyncService
	Authorization   *AuthorizationService
	Job             *job.JobService
}

func NewServices(s *server.Server, repos *port.Repositories) (*Services, error) {
	var enqueuer TaskEnqueuer
	if s.Job != nil {
		enqueuer = s.Job.Client
	}
	authService := NewAuthService(&s.Config.Auth, repos.Auth, repos.AuthSession, repos.EmailVerification, enqueuer, s.Logger)
	userService := NewUserService(repos.User)
	ebookService := NewEbookService(repos.Ebook, repos.EbookMetadata)
	shareService := NewShareService(repos.Share, repos.Borrow, repos.ShareReview, repos.ShareReport)
	readingProgressService := NewReadingProgressService(repos.ReadingProgress)
	bookmarkService := NewBookmarkService(repos.Bookmark)
	annotationService := NewAnnotationService(repos.Annotation)
	userPreferencesService := NewUserPreferencesService(repos.UserPreferences)
	userReaderStateService := NewUserReaderStateService(repos.UserReaderState)
	syncService := NewSyncService(repos.SyncEvent, repos.SyncCheckpoint)
	authorizationService, err := NewAuthorizationService(s.DB.DB, s.Logger)
	if err != nil {
		return nil, err
	}

	return &Services{
		Job:             s.Job,
		Auth:            authService,
		User:            userService,
		Ebook:           ebookService,
		Share:           shareService,
		ReadingProgress: readingProgressService,
		Bookmark:        bookmarkService,
		Annotation:      annotationService,
		UserPreferences: userPreferencesService,
		UserReaderState: userReaderStateService,
		Sync:            syncService,
		Authorization:   authorizationService,
	}, nil
}
