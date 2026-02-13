package application

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/app/errs"
	applicationdto "github.com/jeheskielSunloy77/libra-link/internal/application/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type testUserPreferencesRepo struct {
	data map[uuid.UUID]*domain.UserPreferences
}

func newTestUserPreferencesRepo() *testUserPreferencesRepo {
	return &testUserPreferencesRepo{data: map[uuid.UUID]*domain.UserPreferences{}}
}

func (r *testUserPreferencesRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserPreferences, error) {
	prefs, ok := r.data[userID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return prefs, nil
}

func (r *testUserPreferencesRepo) Upsert(ctx context.Context, prefs *domain.UserPreferences) error {
	clone := *prefs
	if clone.ThemeOverrides == nil {
		clone.ThemeOverrides = map[string]string{}
	}
	r.data[prefs.UserID] = &clone
	return nil
}

type testUserReaderStateRepo struct {
	data map[uuid.UUID]*domain.UserReaderState
}

func newTestUserReaderStateRepo() *testUserReaderStateRepo {
	return &testUserReaderStateRepo{data: map[uuid.UUID]*domain.UserReaderState{}}
}

func (r *testUserReaderStateRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserReaderState, error) {
	state, ok := r.data[userID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return state, nil
}

func (r *testUserReaderStateRepo) Upsert(ctx context.Context, state *domain.UserReaderState) error {
	clone := *state
	r.data[state.UserID] = &clone
	return nil
}

func TestUserPreferencesServicePatch_CreatesDefaultsAndIncrementsVersion(t *testing.T) {
	repo := newTestUserPreferencesRepo()
	service := NewUserPreferencesService(repo)

	userID := uuid.New()
	theme := domain.ThemeModeSepia
	updates := map[string]string{"background": "#101010"}

	updated, err := service.Patch(context.Background(), userID, &applicationdto.UpdateUserPreferencesInput{
		ThemeMode:      &theme,
		ThemeOverrides: updates,
	})
	require.NoError(t, err)
	require.Equal(t, userID, updated.UserID)
	require.Equal(t, domain.ThemeModeSepia, updated.ThemeMode)
	require.Equal(t, int64(2), updated.RowVersion)
	require.Equal(t, "#101010", updated.ThemeOverrides["background"])
	require.Equal(t, domain.ReadingModeNormal, updated.ReadingMode)
}

func TestUserReaderStateServicePatch_CreatesDefaultsAndUpdatesMode(t *testing.T) {
	repo := newTestUserReaderStateRepo()
	service := NewUserReaderStateService(repo)

	userID := uuid.New()
	mode := domain.ReadingModeZen

	updated, err := service.Patch(context.Background(), userID, &applicationdto.UpdateUserReaderStateInput{
		ReadingMode: &mode,
	})
	require.NoError(t, err)
	require.Equal(t, userID, updated.UserID)
	require.Equal(t, domain.ReadingModeZen, updated.ReadingMode)
	require.Equal(t, int64(2), updated.RowVersion)
}

func TestUserPreferencesServicePatch_RejectsNilPayload(t *testing.T) {
	repo := newTestUserPreferencesRepo()
	service := NewUserPreferencesService(repo)

	_, err := service.Patch(context.Background(), uuid.New(), nil)
	require.Error(t, err)

	var httpErr *errs.ErrorResponse
	require.ErrorAs(t, err, &httpErr)
	require.Equal(t, http.StatusBadRequest, httpErr.Status)
}
