package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Share struct {
	ID                   uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt            time.Time       `json:"createdAt"`
	UpdatedAt            time.Time       `json:"updatedAt"`
	DeletedAt            gorm.DeletedAt  `json:"deletedAt"`
	EbookID              uuid.UUID       `json:"ebookId" gorm:"type:uuid;not null;index"`
	OwnerUserID          uuid.UUID       `json:"ownerUserId" gorm:"type:uuid;not null;index"`
	TitleOverride        *string         `json:"titleOverride,omitempty"`
	Description          *string         `json:"description,omitempty"`
	Visibility           ShareVisibility `json:"visibility" gorm:"type:share_visibility;not null;default:public"`
	Status               ShareStatus     `json:"status" gorm:"type:share_status;not null;default:active"`
	BorrowDurationHours  int             `json:"borrowDurationHours" gorm:"not null"`
	MaxConcurrentBorrows int             `json:"maxConcurrentBorrows" gorm:"not null;default:1"`
}

func (m Share) GetID() uuid.UUID {
	return m.ID
}

type Borrow struct {
	ID                  uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt           time.Time    `json:"createdAt"`
	UpdatedAt           time.Time    `json:"updatedAt"`
	ShareID             uuid.UUID    `json:"shareId" gorm:"type:uuid;not null;index"`
	BorrowerUserID      uuid.UUID    `json:"borrowerUserId" gorm:"type:uuid;not null;index"`
	StartedAt           time.Time    `json:"startedAt" gorm:"not null"`
	DueAt               time.Time    `json:"dueAt" gorm:"not null"`
	ReturnedAt          *time.Time   `json:"returnedAt,omitempty"`
	ExpiredAt           *time.Time   `json:"expiredAt,omitempty"`
	Status              BorrowStatus `json:"status" gorm:"type:borrow_status;not null"`
	LegalAcknowledgedAt time.Time    `json:"legalAcknowledgedAt" gorm:"not null"`
}

func (m Borrow) GetID() uuid.UUID {
	return m.ID
}

type ShareReview struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `json:"deletedAt"`
	ShareID    uuid.UUID      `json:"shareId" gorm:"type:uuid;not null;index"`
	UserID     uuid.UUID      `json:"userId" gorm:"type:uuid;not null;index"`
	Rating     int16          `json:"rating" gorm:"not null"`
	ReviewText *string        `json:"reviewText,omitempty"`
}

func (m ShareReview) GetID() uuid.UUID {
	return m.ID
}

type ShareReport struct {
	ID               uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt        time.Time    `json:"createdAt"`
	UpdatedAt        time.Time    `json:"updatedAt"`
	ShareID          uuid.UUID    `json:"shareId" gorm:"type:uuid;not null;index"`
	ReporterUserID   uuid.UUID    `json:"reporterUserId" gorm:"type:uuid;not null;index"`
	Reason           ReportReason `json:"reason" gorm:"type:report_reason;not null"`
	Details          *string      `json:"details,omitempty"`
	Status           ReportStatus `json:"status" gorm:"type:report_status;not null;default:open"`
	ReviewedByUserID *uuid.UUID   `json:"reviewedByUserId,omitempty" gorm:"type:uuid"`
	ReviewedAt       *time.Time   `json:"reviewedAt,omitempty"`
	ResolutionNote   *string      `json:"resolutionNote,omitempty"`
}

func (m ShareReport) GetID() uuid.UUID {
	return m.ID
}
