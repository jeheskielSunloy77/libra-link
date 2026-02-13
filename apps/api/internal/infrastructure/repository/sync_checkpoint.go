package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/application/port"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SyncCheckpointRepository = port.SyncCheckpointRepository

type syncCheckpointRepository struct {
	db *gorm.DB
}

func NewSyncCheckpointRepository(db *gorm.DB) SyncCheckpointRepository {
	return &syncCheckpointRepository{db: db}
}

func (r *syncCheckpointRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.SyncCheckpoint, error) {
	var checkpoint domain.SyncCheckpoint
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&checkpoint).Error; err != nil {
		return nil, err
	}
	return &checkpoint, nil
}

func (r *syncCheckpointRepository) Upsert(ctx context.Context, checkpoint *domain.SyncCheckpoint) error {
	now := time.Now().UTC()
	checkpoint.UpdatedAt = now
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"last_server_timestamp": checkpoint.LastServerTimestamp,
				"last_event_id":         checkpoint.LastEventID,
				"updated_at":            checkpoint.UpdatedAt,
			}),
		}).
		Create(checkpoint).
		Error
}
