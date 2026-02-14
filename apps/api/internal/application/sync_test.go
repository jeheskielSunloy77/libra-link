package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/app/errs"
	applicationdto "github.com/jeheskielSunloy77/libra-link/internal/application/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	infrarepo "github.com/jeheskielSunloy77/libra-link/internal/infrastructure/repository"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type testSyncEventRepo struct {
	*infrarepo.MockResourceRepository[domain.SyncEvent]
	getByUserAndIdempotencyKeyFn func(ctx context.Context, userID uuid.UUID, idempotencyKey string) (*domain.SyncEvent, error)
	listSinceFn                  func(ctx context.Context, userID uuid.UUID, since *time.Time, limit int) ([]domain.SyncEvent, error)
	storeFn                      func(ctx context.Context, entity *domain.SyncEvent) error
	storeCalls                   int
	storedEvent                  *domain.SyncEvent
}

func newTestSyncEventRepo() *testSyncEventRepo {
	return &testSyncEventRepo{
		MockResourceRepository: infrarepo.NewMockResourceRepository[domain.SyncEvent](false),
	}
}

func (r *testSyncEventRepo) Store(ctx context.Context, entity *domain.SyncEvent) error {
	r.storeCalls++
	clone := *entity
	r.storedEvent = &clone
	if r.storeFn != nil {
		return r.storeFn(ctx, entity)
	}
	return r.MockResourceRepository.Store(ctx, entity)
}

func (r *testSyncEventRepo) GetByUserAndIdempotencyKey(ctx context.Context, userID uuid.UUID, idempotencyKey string) (*domain.SyncEvent, error) {
	if r.getByUserAndIdempotencyKeyFn != nil {
		return r.getByUserAndIdempotencyKeyFn(ctx, userID, idempotencyKey)
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *testSyncEventRepo) ListSince(ctx context.Context, userID uuid.UUID, since *time.Time, limit int) ([]domain.SyncEvent, error) {
	if r.listSinceFn != nil {
		return r.listSinceFn(ctx, userID, since, limit)
	}
	return nil, nil
}

type testSyncCheckpointRepo struct {
	getByUserIDFn func(ctx context.Context, userID uuid.UUID) (*domain.SyncCheckpoint, error)
	upsertFn      func(ctx context.Context, checkpoint *domain.SyncCheckpoint) error
	upsertCalls   int
	lastUpserted  *domain.SyncCheckpoint
}

func (r *testSyncCheckpointRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.SyncCheckpoint, error) {
	if r.getByUserIDFn != nil {
		return r.getByUserIDFn(ctx, userID)
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *testSyncCheckpointRepo) Upsert(ctx context.Context, checkpoint *domain.SyncCheckpoint) error {
	r.upsertCalls++
	clone := *checkpoint
	r.lastUpserted = &clone
	if r.upsertFn != nil {
		return r.upsertFn(ctx, checkpoint)
	}
	return nil
}

func TestSyncServiceStoreEvent_RejectsNilInput(t *testing.T) {
	service := NewSyncService(newTestSyncEventRepo(), &testSyncCheckpointRepo{})

	_, err := service.StoreEvent(context.Background(), nil)
	require.Error(t, err)

	var httpErr *errs.ErrorResponse
	require.ErrorAs(t, err, &httpErr)
	require.Equal(t, 400, httpErr.Status)
}

func TestSyncServiceStoreEvent_ReturnsExistingForIdempotencyKey(t *testing.T) {
	eventRepo := newTestSyncEventRepo()
	checkpointRepo := &testSyncCheckpointRepo{}
	service := NewSyncService(eventRepo, checkpointRepo)

	existing := &domain.SyncEvent{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		IdempotencyKey: "existing-idempotency",
	}
	eventRepo.getByUserAndIdempotencyKeyFn = func(ctx context.Context, userID uuid.UUID, idempotencyKey string) (*domain.SyncEvent, error) {
		return existing, nil
	}

	stored, err := service.StoreEvent(context.Background(), &applicationdto.StoreSyncEventInput{
		UserID:          existing.UserID,
		EntityType:      domain.SyncEntityTypeReader,
		EntityID:        uuid.New(),
		Operation:       domain.SyncOperationUpsert,
		ClientTimestamp: time.Now().UTC(),
		IdempotencyKey:  existing.IdempotencyKey,
	})
	require.NoError(t, err)
	require.Same(t, existing, stored)
	require.Equal(t, 0, eventRepo.storeCalls)
	require.Equal(t, 0, checkpointRepo.upsertCalls)
}

func TestSyncServiceStoreEvent_AssignsIDAndUpdatesCheckpoint(t *testing.T) {
	eventRepo := newTestSyncEventRepo()
	checkpointRepo := &testSyncCheckpointRepo{}
	service := NewSyncService(eventRepo, checkpointRepo)

	eventRepo.getByUserAndIdempotencyKeyFn = func(ctx context.Context, userID uuid.UUID, idempotencyKey string) (*domain.SyncEvent, error) {
		return nil, gorm.ErrRecordNotFound
	}

	input := &applicationdto.StoreSyncEventInput{
		UserID:          uuid.New(),
		EntityType:      domain.SyncEntityTypeReader,
		EntityID:        uuid.New(),
		Operation:       domain.SyncOperationUpsert,
		ClientTimestamp: time.Now().UTC(),
		IdempotencyKey:  uuid.NewString(),
	}

	stored, err := service.StoreEvent(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, stored)
	require.NotEqual(t, uuid.Nil, stored.ID)
	require.Equal(t, 1, eventRepo.storeCalls)
	require.NotNil(t, eventRepo.storedEvent)
	require.Equal(t, stored.ID, eventRepo.storedEvent.ID)

	require.Equal(t, 1, checkpointRepo.upsertCalls)
	require.NotNil(t, checkpointRepo.lastUpserted)
	require.NotNil(t, checkpointRepo.lastUpserted.LastEventID)
	require.Equal(t, stored.ID, *checkpointRepo.lastUpserted.LastEventID)
}
