package dto

import (
	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
)

type StoreShareInput struct {
	EbookID              uuid.UUID
	OwnerUserID          uuid.UUID
	TitleOverride        *string
	Description          *string
	Visibility           domain.ShareVisibility
	Status               domain.ShareStatus
	BorrowDurationHours  int
	MaxConcurrentBorrows int
}

func (d *StoreShareInput) ToModel() *domain.Share {
	status := d.Status
	if status == "" {
		status = domain.ShareStatusActive
	}
	visibility := d.Visibility
	if visibility == "" {
		visibility = domain.ShareVisibilityPublic
	}

	return &domain.Share{
		EbookID:              d.EbookID,
		OwnerUserID:          d.OwnerUserID,
		TitleOverride:        d.TitleOverride,
		Description:          d.Description,
		Visibility:           visibility,
		Status:               status,
		BorrowDurationHours:  d.BorrowDurationHours,
		MaxConcurrentBorrows: d.MaxConcurrentBorrows,
	}
}

type UpdateShareInput struct {
	TitleOverride        *string
	Description          *string
	Visibility           *domain.ShareVisibility
	Status               *domain.ShareStatus
	BorrowDurationHours  *int
	MaxConcurrentBorrows *int
}

func (d *UpdateShareInput) ToModel() *domain.Share {
	out := &domain.Share{}
	if d.TitleOverride != nil {
		out.TitleOverride = d.TitleOverride
	}
	if d.Description != nil {
		out.Description = d.Description
	}
	if d.Visibility != nil {
		out.Visibility = *d.Visibility
	}
	if d.Status != nil {
		out.Status = *d.Status
	}
	if d.BorrowDurationHours != nil {
		out.BorrowDurationHours = *d.BorrowDurationHours
	}
	if d.MaxConcurrentBorrows != nil {
		out.MaxConcurrentBorrows = *d.MaxConcurrentBorrows
	}
	return out
}

func (d *UpdateShareInput) ToMap() map[string]any {
	updates := map[string]any{}
	if d.TitleOverride != nil {
		updates["title_override"] = d.TitleOverride
	}
	if d.Description != nil {
		updates["description"] = d.Description
	}
	if d.Visibility != nil {
		updates["visibility"] = *d.Visibility
	}
	if d.Status != nil {
		updates["status"] = *d.Status
	}
	if d.BorrowDurationHours != nil {
		updates["borrow_duration_hours"] = *d.BorrowDurationHours
	}
	if d.MaxConcurrentBorrows != nil {
		updates["max_concurrent_borrows"] = *d.MaxConcurrentBorrows
	}
	return updates
}

type BorrowShareInput struct {
	ShareID             uuid.UUID
	BorrowerUserID      uuid.UUID
	LegalAcknowledgedAt *bool
}

type ReturnBorrowInput struct {
	BorrowID       uuid.UUID
	BorrowerUserID uuid.UUID
	ForceByOwnerID *uuid.UUID
}

type UpsertShareReviewInput struct {
	ShareID    uuid.UUID
	UserID     uuid.UUID
	Rating     int16
	ReviewText *string
}

type CreateShareReportInput struct {
	ShareID        uuid.UUID
	ReporterUserID uuid.UUID
	Reason         domain.ReportReason
	Details        *string
}
