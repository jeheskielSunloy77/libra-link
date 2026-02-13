package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
)

type StoreSyncEventInput struct {
	UserID          uuid.UUID
	EntityType      domain.SyncEntityType
	EntityID        uuid.UUID
	Operation       domain.SyncOperation
	Payload         map[string]any
	BaseVersion     *int64
	ClientTimestamp time.Time
	IdempotencyKey  string
}

func (d *StoreSyncEventInput) ToModel() *domain.SyncEvent {
	return &domain.SyncEvent{
		UserID:          d.UserID,
		EntityType:      d.EntityType,
		EntityID:        d.EntityID,
		Operation:       d.Operation,
		Payload:         d.Payload,
		BaseVersion:     d.BaseVersion,
		ClientTimestamp: d.ClientTimestamp,
		ServerTimestamp: time.Now().UTC(),
		IdempotencyKey:  d.IdempotencyKey,
	}
}
