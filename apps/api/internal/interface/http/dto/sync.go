package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	applicationdto "github.com/jeheskielSunloy77/libra-link/internal/application/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
)

type StoreSyncEventRequest struct {
	EntityType      domain.SyncEntityType `json:"entityType" validate:"required,oneof=progress annotation bookmark preference reader_state"`
	EntityID        uuid.UUID             `json:"entityId" validate:"required"`
	Operation       domain.SyncOperation  `json:"operation" validate:"required,oneof=upsert delete"`
	Payload         map[string]any        `json:"payload"`
	BaseVersion     *int64                `json:"baseVersion" validate:"omitempty,gte=1"`
	ClientTimestamp *time.Time            `json:"clientTimestamp"`
	IdempotencyKey  string                `json:"idempotencyKey" validate:"required,min=8,max=255"`
}

func (d *StoreSyncEventRequest) Validate() error {
	return validator.New().Struct(d)
}

func (d *StoreSyncEventRequest) ToUsecase() *applicationdto.StoreSyncEventInput {
	clientTS := time.Now().UTC()
	if d.ClientTimestamp != nil {
		clientTS = d.ClientTimestamp.UTC()
	}

	return &applicationdto.StoreSyncEventInput{
		EntityType:      d.EntityType,
		EntityID:        d.EntityID,
		Operation:       d.Operation,
		Payload:         d.Payload,
		BaseVersion:     d.BaseVersion,
		ClientTimestamp: clientTS,
		IdempotencyKey:  d.IdempotencyKey,
	}
}

func parseSinceQuery(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	utc := parsed.UTC()
	return &utc, nil
}
