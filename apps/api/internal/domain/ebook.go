package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Ebook struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt"`

	OwnerUserID    uuid.UUID   `json:"ownerUserId" gorm:"type:uuid;not null;index"`
	Owner          *User       `json:"owner,omitempty" gorm:"foreignKey:OwnerUserID"`
	Title          string      `json:"title" gorm:"not null"`
	Description    *string     `json:"description,omitempty"`
	Format         EbookFormat `json:"format" gorm:"type:ebook_format;not null"`
	LanguageCode   *string     `json:"languageCode,omitempty" gorm:"type:varchar(16)"`
	StorageKey     string      `json:"storageKey" gorm:"not null;uniqueIndex"`
	FileSizeBytes  int64       `json:"fileSizeBytes" gorm:"not null"`
	ChecksumSHA256 string      `json:"checksumSha256" gorm:"type:varchar(64);not null"`
	ImportedAt     time.Time   `json:"importedAt" gorm:"not null"`
}

func (m Ebook) GetID() uuid.UUID {
	return m.ID
}

type Author struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt"`

	Name string `json:"name" gorm:"not null"`
}

func (m Author) GetID() uuid.UUID {
	return m.ID
}

type EbookAuthor struct {
	EbookID  uuid.UUID `json:"ebookId" gorm:"type:uuid;primaryKey"`
	AuthorID uuid.UUID `json:"authorId" gorm:"type:uuid;primaryKey"`
}

type Tag struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt"`

	Name string `json:"name" gorm:"not null"`
}

func (m Tag) GetID() uuid.UUID {
	return m.ID
}

type EbookTag struct {
	EbookID uuid.UUID `json:"ebookId" gorm:"type:uuid;primaryKey"`
	TagID   uuid.UUID `json:"tagId" gorm:"type:uuid;primaryKey"`
}

type EbookGoogleMetadata struct {
	EbookID       uuid.UUID      `json:"ebookId" gorm:"type:uuid;primaryKey"`
	GoogleBooksID string         `json:"googleBooksId" gorm:"not null"`
	ISBN10        *string        `json:"isbn10,omitempty" gorm:"type:varchar(10)"`
	ISBN13        *string        `json:"isbn13,omitempty" gorm:"type:varchar(13)"`
	Publisher     *string        `json:"publisher,omitempty"`
	PublishedDate *string        `json:"publishedDate,omitempty"`
	PageCount     *int           `json:"pageCount,omitempty"`
	Categories    []string       `json:"categories,omitempty" gorm:"type:jsonb;serializer:json"`
	ThumbnailURL  *string        `json:"thumbnailUrl,omitempty"`
	InfoLink      *string        `json:"infoLink,omitempty"`
	RawPayload    map[string]any `json:"rawPayload,omitempty" gorm:"type:jsonb;serializer:json"`
	AttachedAt    time.Time      `json:"attachedAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `json:"deletedAt"`
}

func (m EbookGoogleMetadata) GetID() uuid.UUID {
	return m.EbookID
}
