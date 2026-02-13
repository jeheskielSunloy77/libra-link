package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/application/port"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EbookGoogleMetadataRepository = port.EbookGoogleMetadataRepository

type ebookGoogleMetadataRepository struct {
	db *gorm.DB
}

func NewEbookGoogleMetadataRepository(db *gorm.DB) EbookGoogleMetadataRepository {
	return &ebookGoogleMetadataRepository{db: db}
}

func (r *ebookGoogleMetadataRepository) GetByEbookID(ctx context.Context, ebookID uuid.UUID) (*domain.EbookGoogleMetadata, error) {
	var metadata domain.EbookGoogleMetadata
	if err := r.db.WithContext(ctx).Where("ebook_id = ?", ebookID).First(&metadata).Error; err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (r *ebookGoogleMetadataRepository) Upsert(ctx context.Context, metadata *domain.EbookGoogleMetadata) error {
	if metadata == nil {
		return errors.New("metadata is required")
	}

	now := time.Now().UTC()
	metadata.UpdatedAt = now
	if metadata.AttachedAt.IsZero() {
		metadata.AttachedAt = now
	}

	return r.db.WithContext(ctx).
		Unscoped().
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "ebook_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"google_books_id", "isbn_10", "isbn_13", "publisher", "published_date", "page_count", "categories", "thumbnail_url", "info_link", "raw_payload", "updated_at", "attached_at", "deleted_at"}),
		}).
		Create(metadata).
		Error
}

func (r *ebookGoogleMetadataRepository) SoftDeleteByEbookID(ctx context.Context, ebookID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&domain.EbookGoogleMetadata{}).
		Where("ebook_id = ? AND deleted_at IS NULL", ebookID).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).
		Error
}
