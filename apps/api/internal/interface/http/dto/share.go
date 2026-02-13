package dto

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	applicationdto "github.com/jeheskielSunloy77/libra-link/internal/application/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
)

type StoreShareRequest struct {
	EbookID              uuid.UUID              `json:"ebookId" validate:"required"`
	TitleOverride        *string                `json:"titleOverride" validate:"omitempty,max=255"`
	Description          *string                `json:"description"`
	Visibility           domain.ShareVisibility `json:"visibility" validate:"omitempty,oneof=public unlisted"`
	Status               domain.ShareStatus     `json:"status" validate:"omitempty,oneof=active disabled removed"`
	BorrowDurationHours  int                    `json:"borrowDurationHours" validate:"required,gt=0"`
	MaxConcurrentBorrows int                    `json:"maxConcurrentBorrows" validate:"omitempty,gte=1"`
}

func (d *StoreShareRequest) Validate() error {
	return validator.New().Struct(d)
}

func (d *StoreShareRequest) ToUsecase() *applicationdto.StoreShareInput {
	maxConcurrent := d.MaxConcurrentBorrows
	if maxConcurrent == 0 {
		maxConcurrent = 1
	}

	return &applicationdto.StoreShareInput{
		EbookID:              d.EbookID,
		TitleOverride:        d.TitleOverride,
		Description:          d.Description,
		Visibility:           d.Visibility,
		Status:               d.Status,
		BorrowDurationHours:  d.BorrowDurationHours,
		MaxConcurrentBorrows: maxConcurrent,
	}
}

type UpdateShareRequest struct {
	TitleOverride        *string                 `json:"titleOverride" validate:"omitempty,max=255"`
	Description          *string                 `json:"description"`
	Visibility           *domain.ShareVisibility `json:"visibility" validate:"omitempty,oneof=public unlisted"`
	Status               *domain.ShareStatus     `json:"status" validate:"omitempty,oneof=active disabled removed"`
	BorrowDurationHours  *int                    `json:"borrowDurationHours" validate:"omitempty,gt=0"`
	MaxConcurrentBorrows *int                    `json:"maxConcurrentBorrows" validate:"omitempty,gte=1"`
}

func (d *UpdateShareRequest) Validate() error {
	return validator.New().Struct(d)
}

func (d *UpdateShareRequest) ToUsecase() *applicationdto.UpdateShareInput {
	return &applicationdto.UpdateShareInput{
		TitleOverride:        d.TitleOverride,
		Description:          d.Description,
		Visibility:           d.Visibility,
		Status:               d.Status,
		BorrowDurationHours:  d.BorrowDurationHours,
		MaxConcurrentBorrows: d.MaxConcurrentBorrows,
	}
}

type BorrowShareRequest struct {
	LegalAcknowledged bool `json:"legalAcknowledged" validate:"required,eq=true"`
}

func (d *BorrowShareRequest) Validate() error {
	return validator.New().Struct(d)
}

type UpsertShareReviewRequest struct {
	Rating     int16   `json:"rating" validate:"required,min=1,max=5"`
	ReviewText *string `json:"reviewText"`
}

func (d *UpsertShareReviewRequest) Validate() error {
	return validator.New().Struct(d)
}

type CreateShareReportRequest struct {
	Reason  domain.ReportReason `json:"reason" validate:"required,oneof=copyright abuse spam other"`
	Details *string             `json:"details"`
}

func (d *CreateShareReportRequest) Validate() error {
	return validator.New().Struct(d)
}
