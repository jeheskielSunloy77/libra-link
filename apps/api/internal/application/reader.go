package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/app/errs"
	"github.com/jeheskielSunloy77/libra-link/internal/app/sqlerr"
	applicationdto "github.com/jeheskielSunloy77/libra-link/internal/application/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/application/port"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"gorm.io/gorm"
)

type ReadingProgressService interface {
	ResourceService[domain.ReadingProgress, *applicationdto.StoreReadingProgressInput, *applicationdto.UpdateReadingProgressInput]
}

type BookmarkService interface {
	ResourceService[domain.Bookmark, *applicationdto.StoreBookmarkInput, *applicationdto.UpdateBookmarkInput]
}

type AnnotationService interface {
	ResourceService[domain.Annotation, *applicationdto.StoreAnnotationInput, *applicationdto.UpdateAnnotationInput]
}

type UserPreferencesService interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserPreferences, error)
	Patch(ctx context.Context, userID uuid.UUID, input *applicationdto.UpdateUserPreferencesInput) (*domain.UserPreferences, error)
}

type UserReaderStateService interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserReaderState, error)
	Patch(ctx context.Context, userID uuid.UUID, input *applicationdto.UpdateUserReaderStateInput) (*domain.UserReaderState, error)
}

type userPreferencesService struct {
	repo port.UserPreferencesRepository
}

type userReaderStateService struct {
	repo port.UserReaderStateRepository
}

func NewReadingProgressService(repo port.ReadingProgressRepository) ReadingProgressService {
	return NewResourceService[domain.ReadingProgress, *applicationdto.StoreReadingProgressInput, *applicationdto.UpdateReadingProgressInput]("reading_progress", repo)
}

func NewBookmarkService(repo port.BookmarkRepository) BookmarkService {
	return NewResourceService[domain.Bookmark, *applicationdto.StoreBookmarkInput, *applicationdto.UpdateBookmarkInput]("bookmark", repo)
}

func NewAnnotationService(repo port.AnnotationRepository) AnnotationService {
	return NewResourceService[domain.Annotation, *applicationdto.StoreAnnotationInput, *applicationdto.UpdateAnnotationInput]("annotation", repo)
}

func NewUserPreferencesService(repo port.UserPreferencesRepository) UserPreferencesService {
	return &userPreferencesService{repo: repo}
}

func NewUserReaderStateService(repo port.UserReaderStateRepository) UserReaderStateService {
	return &userReaderStateService{repo: repo}
}

func (s *userPreferencesService) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserPreferences, error) {
	prefs, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaults := &domain.UserPreferences{
				UserID:            userID,
				ReadingMode:       domain.ReadingModeNormal,
				ZenRestoreOnOpen:  true,
				ThemeMode:         domain.ThemeModeDark,
				ThemeOverrides:    map[string]string{},
				TypographyProfile: domain.TypographyProfileComfortable,
				RowVersion:        1,
			}
			if err := s.repo.Upsert(ctx, defaults); err != nil {
				return nil, sqlerr.HandleError(err)
			}
			return defaults, nil
		}
		return nil, sqlerr.HandleError(err)
	}
	if prefs.ThemeOverrides == nil {
		prefs.ThemeOverrides = map[string]string{}
	}
	return prefs, nil
}

func (s *userPreferencesService) Patch(ctx context.Context, userID uuid.UUID, input *applicationdto.UpdateUserPreferencesInput) (*domain.UserPreferences, error) {
	if input == nil {
		return nil, errs.NewBadRequestError("preferences payload is required", true, nil, nil)
	}

	prefs, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.ReadingMode != nil {
		prefs.ReadingMode = *input.ReadingMode
	}
	if input.ZenRestoreOnOpen != nil {
		prefs.ZenRestoreOnOpen = *input.ZenRestoreOnOpen
	}
	if input.ThemeMode != nil {
		prefs.ThemeMode = *input.ThemeMode
	}
	if input.TypographyProfile != nil {
		prefs.TypographyProfile = *input.TypographyProfile
	}
	if input.ThemeOverrides != nil {
		if prefs.ThemeOverrides == nil {
			prefs.ThemeOverrides = map[string]string{}
		}
		for key, value := range input.ThemeOverrides {
			prefs.ThemeOverrides[key] = value
		}
	}

	prefs.RowVersion++
	prefs.UpdatedAt = time.Now().UTC()

	if err := s.repo.Upsert(ctx, prefs); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	return s.GetByUserID(ctx, userID)
}

func (s *userReaderStateService) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserReaderState, error) {
	state, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaults := &domain.UserReaderState{
				UserID:      userID,
				ReadingMode: domain.ReadingModeNormal,
				RowVersion:  1,
			}
			if err := s.repo.Upsert(ctx, defaults); err != nil {
				return nil, sqlerr.HandleError(err)
			}
			return defaults, nil
		}
		return nil, sqlerr.HandleError(err)
	}
	return state, nil
}

func (s *userReaderStateService) Patch(ctx context.Context, userID uuid.UUID, input *applicationdto.UpdateUserReaderStateInput) (*domain.UserReaderState, error) {
	if input == nil {
		return nil, errs.NewBadRequestError("reader state payload is required", true, nil, nil)
	}

	state, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.CurrentEbookID != nil {
		state.CurrentEbookID = input.CurrentEbookID
	}
	if input.CurrentLocation != nil {
		state.CurrentLocation = input.CurrentLocation
	}
	if input.ReadingMode != nil {
		state.ReadingMode = *input.ReadingMode
	}
	if input.LastOpenedAt != nil {
		state.LastOpenedAt = input.LastOpenedAt
	}

	state.RowVersion++
	state.UpdatedAt = time.Now().UTC()

	if err := s.repo.Upsert(ctx, state); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	return s.GetByUserID(ctx, userID)
}
