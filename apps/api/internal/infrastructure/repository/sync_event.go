package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/application/port"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/config"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/lib/cache"
	"gorm.io/gorm"
)

type SyncEventRepository = port.SyncEventRepository

type syncEventRepository struct {
	ResourceRepository[domain.SyncEvent]
	db *gorm.DB
}

func NewSyncEventRepository(cfg *config.Config, db *gorm.DB, cacheClient cache.Cache) SyncEventRepository {
	return &syncEventRepository{
		ResourceRepository: NewResourceRepository[domain.SyncEvent](cfg, db, cacheClient),
		db:                 db,
	}
}

func (r *syncEventRepository) GetByUserAndIdempotencyKey(ctx context.Context, userID uuid.UUID, idempotencyKey string) (*domain.SyncEvent, error) {
	var event domain.SyncEvent
	err := r.db.WithContext(ctx).Where("user_id = ? AND idempotency_key = ?", userID, idempotencyKey).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *syncEventRepository) ListSince(ctx context.Context, userID uuid.UUID, since *time.Time, limit int) ([]domain.SyncEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if since != nil {
		query = query.Where("server_timestamp > ?", *since)
	}

	var events []domain.SyncEvent
	if err := query.Order("server_timestamp asc").Limit(limit).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}
