package dto

import (
	"time"

	"github.com/go-playground/validator/v10"
	applicationdto "github.com/jeheskielSunloy77/libra-link/internal/application/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
)

type StoreEbookRequest struct {
	Title          string             `json:"title" validate:"required,min=1,max=255"`
	Description    *string            `json:"description"`
	Format         domain.EbookFormat `json:"format" validate:"required,oneof=epub pdf txt"`
	LanguageCode   *string            `json:"languageCode" validate:"omitempty,max=16"`
	StorageKey     string             `json:"storageKey" validate:"required,min=1,max=500"`
	FileSizeBytes  int64              `json:"fileSizeBytes" validate:"required,gt=0"`
	ChecksumSHA256 string             `json:"checksumSha256" validate:"required,len=64,hexadecimal"`
	ImportedAt     *time.Time         `json:"importedAt"`
}

func (d *StoreEbookRequest) Validate() error {
	return validator.New().Struct(d)
}

func (d *StoreEbookRequest) ToUsecase() *applicationdto.StoreEbookInput {
	return &applicationdto.StoreEbookInput{
		Title:          d.Title,
		Description:    d.Description,
		Format:         d.Format,
		LanguageCode:   d.LanguageCode,
		StorageKey:     d.StorageKey,
		FileSizeBytes:  d.FileSizeBytes,
		ChecksumSHA256: d.ChecksumSHA256,
		ImportedAt:     d.ImportedAt,
	}
}

type UpdateEbookRequest struct {
	Title        *string `json:"title" validate:"omitempty,min=1,max=255"`
	Description  *string `json:"description"`
	LanguageCode *string `json:"languageCode" validate:"omitempty,max=16"`
}

func (d *UpdateEbookRequest) Validate() error {
	return validator.New().Struct(d)
}

func (d *UpdateEbookRequest) ToUsecase() *applicationdto.UpdateEbookInput {
	return &applicationdto.UpdateEbookInput{
		Title:        d.Title,
		Description:  d.Description,
		LanguageCode: d.LanguageCode,
	}
}

type AttachGoogleMetadataRequest struct {
	GoogleBooksID string         `json:"googleBooksId" validate:"required,min=1,max=128"`
	ISBN10        *string        `json:"isbn10" validate:"omitempty,len=10"`
	ISBN13        *string        `json:"isbn13" validate:"omitempty,len=13"`
	Publisher     *string        `json:"publisher"`
	PublishedDate *string        `json:"publishedDate"`
	PageCount     *int           `json:"pageCount" validate:"omitempty,gte=1"`
	Categories    []string       `json:"categories"`
	ThumbnailURL  *string        `json:"thumbnailUrl" validate:"omitempty,url"`
	InfoLink      *string        `json:"infoLink" validate:"omitempty,url"`
	RawPayload    map[string]any `json:"rawPayload"`
}

func (d *AttachGoogleMetadataRequest) Validate() error {
	return validator.New().Struct(d)
}

func (d *AttachGoogleMetadataRequest) ToUsecase() *applicationdto.AttachGoogleMetadataInput {
	return &applicationdto.AttachGoogleMetadataInput{
		GoogleBooksID: d.GoogleBooksID,
		ISBN10:        d.ISBN10,
		ISBN13:        d.ISBN13,
		Publisher:     d.Publisher,
		PublishedDate: d.PublishedDate,
		PageCount:     d.PageCount,
		Categories:    d.Categories,
		ThumbnailURL:  d.ThumbnailURL,
		InfoLink:      d.InfoLink,
		RawPayload:    d.RawPayload,
	}
}
