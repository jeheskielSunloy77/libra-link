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

type BorrowRepository = port.BorrowRepository

type borrowRepository struct {
	ResourceRepository[domain.Borrow]
	db *gorm.DB
}

func NewBorrowRepository(cfg *config.Config, db *gorm.DB, cacheClient cache.Cache) BorrowRepository {
	return &borrowRepository{
		ResourceRepository: NewResourceRepository[domain.Borrow](cfg, db, cacheClient),
		db:                 db,
	}
}

func (r *borrowRepository) CountActiveByShare(ctx context.Context, shareID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Borrow{}).
		Where("share_id = ? AND status = ?", shareID, domain.BorrowStatusActive).
		Count(&count).
		Error
	return count, err
}

func (r *borrowRepository) GetActiveByShareAndBorrower(ctx context.Context, shareID uuid.UUID, borrowerID uuid.UUID) (*domain.Borrow, error) {
	var borrow domain.Borrow
	err := r.db.WithContext(ctx).
		Where("share_id = ? AND borrower_user_id = ? AND status = ?", shareID, borrowerID, domain.BorrowStatusActive).
		First(&borrow).
		Error
	if err != nil {
		return nil, err
	}
	return &borrow, nil
}
