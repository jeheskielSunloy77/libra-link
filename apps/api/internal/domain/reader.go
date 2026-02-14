package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserPreferences struct {
	UserID            uuid.UUID         `json:"userId" gorm:"type:uuid;primaryKey"`
	ReadingMode       ReadingMode       `json:"readingMode" gorm:"type:reading_mode;not null;default:normal"`
	ZenRestoreOnOpen  bool              `json:"zenRestoreOnOpen" gorm:"not null;default:true"`
	ThemeMode         ThemeMode         `json:"themeMode" gorm:"type:theme_mode;not null;default:dark"`
	ThemeOverrides    map[string]string `json:"themeOverrides" gorm:"type:jsonb;serializer:json;not null;default:'{}'"`
	TypographyProfile TypographyProfile `json:"typographyProfile" gorm:"type:typography_profile;not null;default:comfortable"`
	RowVersion        int64             `json:"rowVersion" gorm:"not null;default:1"`
	CreatedAt         time.Time         `json:"createdAt"`
	UpdatedAt         time.Time         `json:"updatedAt"`
}

func (m UserPreferences) GetID() uuid.UUID {
	return m.UserID
}

func (UserPreferences) TableName() string {
	return "user_preferences"
}

type UserReaderState struct {
	UserID          uuid.UUID   `json:"userId" gorm:"type:uuid;primaryKey"`
	CurrentEbookID  *uuid.UUID  `json:"currentEbookId,omitempty" gorm:"type:uuid"`
	CurrentLocation *string     `json:"currentLocation,omitempty"`
	ReadingMode     ReadingMode `json:"readingMode" gorm:"type:reading_mode;not null;default:normal"`
	RowVersion      int64       `json:"rowVersion" gorm:"not null;default:1"`
	LastOpenedAt    *time.Time  `json:"lastOpenedAt,omitempty"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
}

func (m UserReaderState) GetID() uuid.UUID {
	return m.UserID
}

func (UserReaderState) TableName() string {
	return "user_reader_state"
}

type ReadingProgress struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `json:"deletedAt"`
	UserID          uuid.UUID      `json:"userId" gorm:"type:uuid;not null;index"`
	EbookID         uuid.UUID      `json:"ebookId" gorm:"type:uuid;not null;index"`
	Location        string         `json:"location" gorm:"not null"`
	ProgressPercent *float64       `json:"progressPercent,omitempty"`
	ReadingMode     ReadingMode    `json:"readingMode" gorm:"type:reading_mode;not null"`
	RowVersion      int64          `json:"rowVersion" gorm:"not null;default:1"`
	LastReadAt      *time.Time     `json:"lastReadAt,omitempty"`
}

func (m ReadingProgress) GetID() uuid.UUID {
	return m.ID
}

type Bookmark struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `json:"deletedAt"`
	UserID     uuid.UUID      `json:"userId" gorm:"type:uuid;not null;index"`
	EbookID    uuid.UUID      `json:"ebookId" gorm:"type:uuid;not null;index"`
	Location   string         `json:"location" gorm:"not null"`
	Label      *string        `json:"label,omitempty"`
	RowVersion int64          `json:"rowVersion" gorm:"not null;default:1"`
}

func (m Bookmark) GetID() uuid.UUID {
	return m.ID
}

type Annotation struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `json:"deletedAt"`
	UserID        uuid.UUID      `json:"userId" gorm:"type:uuid;not null;index"`
	EbookID       uuid.UUID      `json:"ebookId" gorm:"type:uuid;not null;index"`
	LocationStart string         `json:"locationStart" gorm:"not null"`
	LocationEnd   string         `json:"locationEnd" gorm:"not null"`
	HighlightText *string        `json:"highlightText,omitempty"`
	Note          *string        `json:"note,omitempty"`
	Color         *string        `json:"color,omitempty" gorm:"type:varchar(32)"`
	RowVersion    int64          `json:"rowVersion" gorm:"not null;default:1"`
}

func (m Annotation) GetID() uuid.UUID {
	return m.ID
}

type SyncEvent struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID      `json:"userId" gorm:"type:uuid;not null;index"`
	EntityType      SyncEntityType `json:"entityType" gorm:"type:sync_entity_type;not null"`
	EntityID        uuid.UUID      `json:"entityId" gorm:"type:uuid;not null"`
	Operation       SyncOperation  `json:"operation" gorm:"type:sync_operation;not null"`
	Payload         map[string]any `json:"payload,omitempty" gorm:"type:jsonb;serializer:json"`
	BaseVersion     *int64         `json:"baseVersion,omitempty"`
	ClientTimestamp time.Time      `json:"clientTimestamp" gorm:"not null"`
	ServerTimestamp time.Time      `json:"serverTimestamp" gorm:"not null"`
	IdempotencyKey  string         `json:"idempotencyKey" gorm:"not null"`
	CreatedAt       time.Time      `json:"createdAt"`
}

func (m SyncEvent) GetID() uuid.UUID {
	return m.ID
}

type SyncCheckpoint struct {
	UserID              uuid.UUID  `json:"userId" gorm:"type:uuid;primaryKey"`
	LastServerTimestamp time.Time  `json:"lastServerTimestamp" gorm:"not null"`
	LastEventID         *uuid.UUID `json:"lastEventId,omitempty" gorm:"type:uuid"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

func (m SyncCheckpoint) GetID() uuid.UUID {
	return m.UserID
}
