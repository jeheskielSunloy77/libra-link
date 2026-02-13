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

type UserPreferencesRepository = port.UserPreferencesRepository

type userPreferencesRepository struct {
	db *gorm.DB
}

func NewUserPreferencesRepository(db *gorm.DB) UserPreferencesRepository {
	return &userPreferencesRepository{db: db}
}

func (r *userPreferencesRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserPreferences, error) {
	var prefs domain.UserPreferences
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&prefs).Error; err != nil {
		return nil, err
	}
	return &prefs, nil
}

func (r *userPreferencesRepository) Upsert(ctx context.Context, prefs *domain.UserPreferences) error {
	now := time.Now().UTC()
	if prefs.CreatedAt.IsZero() {
		prefs.CreatedAt = now
	}
	prefs.UpdatedAt = now

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"reading_mode":        prefs.ReadingMode,
				"zen_restore_on_open": prefs.ZenRestoreOnOpen,
				"theme_mode":          prefs.ThemeMode,
				"theme_overrides":     prefs.ThemeOverrides,
				"typography_profile":  prefs.TypographyProfile,
				"row_version":         prefs.RowVersion,
				"updated_at":          now,
			}),
		}).
		Create(prefs).
		Error
}
