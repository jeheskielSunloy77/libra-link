package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/config"
	internaltesting "github.com/jeheskielSunloy77/libra-link/internal/testing"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestResourceRepositoryStore_AssignsUUIDWhenModelIDIsNil(t *testing.T) {
	testDB, cleanup := internaltesting.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := internaltesting.WithRollbackTransaction(ctx, testDB, func(tx *gorm.DB) error {
		userID := seedUser(t, ctx, tx, "uuid-nil@example.com", "uuid_nil")

		repo := NewResourceRepository[domain.SyncEvent](&config.Config{}, tx, nil)
		now := time.Now().UTC()
		event := &domain.SyncEvent{
			UserID:          userID,
			EntityType:      domain.SyncEntityTypeReader,
			EntityID:        userID,
			Operation:       domain.SyncOperationUpsert,
			Payload:         map[string]any{"source": "test"},
			ClientTimestamp: now,
			ServerTimestamp: now,
			IdempotencyKey:  uuid.NewString(),
		}

		require.NoError(t, repo.Store(ctx, event))
		require.NotEqual(t, uuid.Nil, event.ID)

		stored, fetchErr := repo.GetByID(ctx, event.ID, nil)
		require.NoError(t, fetchErr)
		require.Equal(t, event.ID, stored.ID)
		return nil
	})
	require.NoError(t, err)
}

func TestResourceRepositoryStore_PreservesProvidedUUID(t *testing.T) {
	testDB, cleanup := internaltesting.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := internaltesting.WithRollbackTransaction(ctx, testDB, func(tx *gorm.DB) error {
		userID := seedUser(t, ctx, tx, "uuid-existing@example.com", "uuid_existing")

		repo := NewResourceRepository[domain.SyncEvent](&config.Config{}, tx, nil)
		now := time.Now().UTC()
		expectedID := uuid.New()
		event := &domain.SyncEvent{
			ID:              expectedID,
			UserID:          userID,
			EntityType:      domain.SyncEntityTypeReader,
			EntityID:        userID,
			Operation:       domain.SyncOperationUpsert,
			ClientTimestamp: now,
			ServerTimestamp: now,
			IdempotencyKey:  uuid.NewString(),
		}

		require.NoError(t, repo.Store(ctx, event))
		require.Equal(t, expectedID, event.ID)
		return nil
	})
	require.NoError(t, err)
}

func TestResourceRepositoryStore_NoOpForModelWithoutIDField(t *testing.T) {
	testDB, cleanup := internaltesting.SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := internaltesting.WithRollbackTransaction(ctx, testDB, func(tx *gorm.DB) error {
		userID := seedUser(t, ctx, tx, "uuid-userid@example.com", "uuid_userid")

		repo := NewResourceRepository[domain.UserPreferences](&config.Config{}, tx, nil)
		prefs := &domain.UserPreferences{
			UserID:            userID,
			ReadingMode:       domain.ReadingModeNormal,
			ZenRestoreOnOpen:  true,
			ThemeMode:         domain.ThemeModeDark,
			ThemeOverrides:    map[string]string{},
			TypographyProfile: domain.TypographyProfileComfortable,
			RowVersion:        1,
		}

		require.NoError(t, repo.Store(ctx, prefs))
		require.Equal(t, userID, prefs.UserID)

		stored, fetchErr := repo.GetByID(ctx, userID, nil)
		require.NoError(t, fetchErr)
		require.Equal(t, userID, stored.UserID)
		return nil
	})
	require.NoError(t, err)
}

func seedUser(t *testing.T, ctx context.Context, tx *gorm.DB, email string, username string) uuid.UUID {
	t.Helper()
	userRepo := NewUserRepository(&config.Config{}, tx, nil)
	user := &domain.User{
		ID:       uuid.New(),
		Email:    email,
		Username: username,
	}
	require.NoError(t, userRepo.Store(ctx, user))
	return user.ID
}
