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

type ShareReviewRepository = port.ShareReviewRepository

type shareReviewRepository struct {
	ResourceRepository[domain.ShareReview]
	db *gorm.DB
}

func NewShareReviewRepository(cfg *config.Config, db *gorm.DB, cacheClient cache.Cache) ShareReviewRepository {
	return &shareReviewRepository{
		ResourceRepository: NewResourceRepository[domain.ShareReview](cfg, db, cacheClient),
		db:                 db,
	}
}

func (r *shareReviewRepository) GetByShareAndUser(ctx context.Context, shareID uuid.UUID, userID uuid.UUID) (*domain.ShareReview, error) {
	var review domain.ShareReview
	err := r.db.WithContext(ctx).Where("share_id = ? AND user_id = ?", shareID, userID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}
