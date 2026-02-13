package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
)

type StoreEbookInput struct {
	OwnerUserID    uuid.UUID
	Title          string
	Description    *string
	Format         domain.EbookFormat
	LanguageCode   *string
	StorageKey     string
	FileSizeBytes  int64
	ChecksumSHA256 string
	ImportedAt     *time.Time
}

func (d *StoreEbookInput) ToModel() *domain.Ebook {
	importedAt := time.Now().UTC()
	if d.ImportedAt != nil {
		importedAt = d.ImportedAt.UTC()
	}

	return &domain.Ebook{
		OwnerUserID:    d.OwnerUserID,
		Title:          d.Title,
		Description:    d.Description,
		Format:         d.Format,
		LanguageCode:   d.LanguageCode,
		StorageKey:     d.StorageKey,
		FileSizeBytes:  d.FileSizeBytes,
		ChecksumSHA256: d.ChecksumSHA256,
		ImportedAt:     importedAt,
	}
}

type UpdateEbookInput struct {
	Title        *string
	Description  *string
	LanguageCode *string
}

func (d *UpdateEbookInput) ToModel() *domain.Ebook {
	out := &domain.Ebook{}
	if d.Title != nil {
		out.Title = *d.Title
	}
	if d.Description != nil {
		out.Description = d.Description
	}
	if d.LanguageCode != nil {
		out.LanguageCode = d.LanguageCode
	}
	return out
}

func (d *UpdateEbookInput) ToMap() map[string]any {
	updates := map[string]any{}
	if d.Title != nil {
		updates["title"] = *d.Title
	}
	if d.Description != nil {
		updates["description"] = d.Description
	}
	if d.LanguageCode != nil {
		updates["language_code"] = d.LanguageCode
	}
	return updates
}

type AttachGoogleMetadataInput struct {
	EbookID       uuid.UUID
	GoogleBooksID string
	ISBN10        *string
	ISBN13        *string
	Publisher     *string
	PublishedDate *string
	PageCount     *int
	Categories    []string
	ThumbnailURL  *string
	InfoLink      *string
	RawPayload    map[string]any
}
