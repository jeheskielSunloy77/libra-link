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

type UserReaderStateRepository = port.UserReaderStateRepository

type userReaderStateRepository struct {
	db *gorm.DB
}

func NewUserReaderStateRepository(db *gorm.DB) UserReaderStateRepository {
	return &userReaderStateRepository{db: db}
}

func (r *userReaderStateRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserReaderState, error) {
	var state domain.UserReaderState
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *userReaderStateRepository) Upsert(ctx context.Context, state *domain.UserReaderState) error {
	now := time.Now().UTC()
	if state.CreatedAt.IsZero() {
		state.CreatedAt = now
	}
	state.UpdatedAt = now

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"current_ebook_id": state.CurrentEbookID,
				"current_location": state.CurrentLocation,
				"reading_mode":     state.ReadingMode,
				"row_version":      state.RowVersion,
				"last_opened_at":   state.LastOpenedAt,
				"updated_at":       now,
			}),
		}).
		Create(state).
		Error
}
