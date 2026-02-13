package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/application/port"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/config"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/lib/cache"
	"gorm.io/gorm"
)

type ReadingProgressRepository = port.ReadingProgressRepository

type readingProgressRepository struct {
	ResourceRepository[domain.ReadingProgress]
	db *gorm.DB
}

func NewReadingProgressRepository(cfg *config.Config, db *gorm.DB, cacheClient cache.Cache) ReadingProgressRepository {
	return &readingProgressRepository{
		ResourceRepository: NewResourceRepository[domain.ReadingProgress](cfg, db, cacheClient),
		db:                 db,
	}
}

func (r *readingProgressRepository) GetByUserAndEbook(ctx context.Context, userID uuid.UUID, ebookID uuid.UUID) (*domain.ReadingProgress, error) {
	var progress domain.ReadingProgress
	if err := r.db.WithContext(ctx).Where("user_id = ? AND ebook_id = ?", userID, ebookID).First(&progress).Error; err != nil {
		return nil, err
	}
	return &progress, nil
}
