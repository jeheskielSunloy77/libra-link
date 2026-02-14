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

type SyncService interface {
	StoreEvent(ctx context.Context, input *applicationdto.StoreSyncEventInput) (*domain.SyncEvent, error)
	ListEvents(ctx context.Context, userID uuid.UUID, since *time.Time, limit int) ([]domain.SyncEvent, error)
}

type syncService struct {
	eventRepo      port.SyncEventRepository
	checkpointRepo port.SyncCheckpointRepository
}

func NewSyncService(eventRepo port.SyncEventRepository, checkpointRepo port.SyncCheckpointRepository) SyncService {
	return &syncService{eventRepo: eventRepo, checkpointRepo: checkpointRepo}
}

func (s *syncService) StoreEvent(ctx context.Context, input *applicationdto.StoreSyncEventInput) (*domain.SyncEvent, error) {
	if input == nil {
		return nil, errs.NewBadRequestError("sync event payload is required", true, nil, nil)
	}

	existing, err := s.eventRepo.GetByUserAndIdempotencyKey(ctx, input.UserID, input.IdempotencyKey)
	if err == nil {
		return existing, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sqlerr.HandleError(err)
	}

	event := input.ToModel()
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	event.ServerTimestamp = time.Now().UTC()

	if err := s.eventRepo.Store(ctx, event); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	checkpoint := &domain.SyncCheckpoint{
		UserID:              input.UserID,
		LastServerTimestamp: event.ServerTimestamp,
		LastEventID:         &event.ID,
	}
	if err := s.checkpointRepo.Upsert(ctx, checkpoint); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	return event, nil
}

func (s *syncService) ListEvents(ctx context.Context, userID uuid.UUID, since *time.Time, limit int) ([]domain.SyncEvent, error) {
	events, err := s.eventRepo.ListSince(ctx, userID, since, limit)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	return events, nil
}
