package dto

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	applicationdto "github.com/jeheskielSunloy77/libra-link/internal/application/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"github.com/jeheskielSunloy77/libra-link/internal/interface/http/validation"
)

type StoreReadingProgressRequest struct {
	EbookID         uuid.UUID          `json:"ebookId" validate:"required"`
	Location        string             `json:"location" validate:"required,min=1"`
	ProgressPercent *float64           `json:"progressPercent" validate:"omitempty,gte=0,lte=100"`
	ReadingMode     domain.ReadingMode `json:"readingMode" validate:"required,oneof=normal zen"`
	LastReadAt      *time.Time         `json:"lastReadAt"`
}

func (d *StoreReadingProgressRequest) Validate() error {
	return validator.New().Struct(d)
}

func (d *StoreReadingProgressRequest) ToUsecase() *applicationdto.StoreReadingProgressInput {
	return &applicationdto.StoreReadingProgressInput{
		EbookID:         d.EbookID,
		Location:        d.Location,
		ProgressPercent: d.ProgressPercent,
		ReadingMode:     d.ReadingMode,
		LastReadAt:      d.LastReadAt,
	}
}

type UpdateReadingProgressRequest struct {
	Location        *string             `json:"location" validate:"omitempty,min=1"`
	ProgressPercent *float64            `json:"progressPercent" validate:"omitempty,gte=0,lte=100"`
	ReadingMode     *domain.ReadingMode `json:"readingMode" validate:"omitempty,oneof=normal zen"`
	LastReadAt      *time.Time          `json:"lastReadAt"`
}

func (d *UpdateReadingProgressRequest) Validate() error {
	return validator.New().Struct(d)
}

func (d *UpdateReadingProgressRequest) ToUsecase() *applicationdto.UpdateReadingProgressInput {
	return &applicationdto.UpdateReadingProgressInput{
		Location:        d.Location,
		ProgressPercent: d.ProgressPercent,
		ReadingMode:     d.ReadingMode,
		LastReadAt:      d.LastReadAt,
	}
}

type StoreBookmarkRequest struct {
	EbookID  uuid.UUID `json:"ebookId" validate:"required"`
	Location string    `json:"location" validate:"required,min=1"`
	Label    *string   `json:"label"`
}

func (d *StoreBookmarkRequest) Validate() error {
	return validator.New().Struct(d)
}

func (d *StoreBookmarkRequest) ToUsecase() *applicationdto.StoreBookmarkInput {
	return &applicationdto.StoreBookmarkInput{
		EbookID:  d.EbookID,
		Location: d.Location,
		Label:    d.Label,
	}
}

type UpdateBookmarkRequest struct {
	Label *string `json:"label"`
}

func (d *UpdateBookmarkRequest) Validate() error {
	return validator.New().Struct(d)
}

func (d *UpdateBookmarkRequest) ToUsecase() *applicationdto.UpdateBookmarkInput {
	return &applicationdto.UpdateBookmarkInput{Label: d.Label}
}

type StoreAnnotationRequest struct {
	EbookID       uuid.UUID `json:"ebookId" validate:"required"`
	LocationStart string    `json:"locationStart" validate:"required,min=1"`
	LocationEnd   string    `json:"locationEnd" validate:"required,min=1"`
	HighlightText *string   `json:"highlightText"`
	Note          *string   `json:"note"`
	Color         *string   `json:"color" validate:"omitempty,max=32"`
}

func (d *StoreAnnotationRequest) Validate() error {
	return validator.New().Struct(d)
}

func (d *StoreAnnotationRequest) ToUsecase() *applicationdto.StoreAnnotationInput {
	return &applicationdto.StoreAnnotationInput{
		EbookID:       d.EbookID,
		LocationStart: d.LocationStart,
		LocationEnd:   d.LocationEnd,
		HighlightText: d.HighlightText,
		Note:          d.Note,
		Color:         d.Color,
	}
}

type UpdateAnnotationRequest struct {
	LocationStart *string `json:"locationStart" validate:"omitempty,min=1"`
	LocationEnd   *string `json:"locationEnd" validate:"omitempty,min=1"`
	HighlightText *string `json:"highlightText"`
	Note          *string `json:"note"`
	Color         *string `json:"color" validate:"omitempty,max=32"`
}

func (d *UpdateAnnotationRequest) Validate() error {
	return validator.New().Struct(d)
}

func (d *UpdateAnnotationRequest) ToUsecase() *applicationdto.UpdateAnnotationInput {
	return &applicationdto.UpdateAnnotationInput{
		LocationStart: d.LocationStart,
		LocationEnd:   d.LocationEnd,
		HighlightText: d.HighlightText,
		Note:          d.Note,
		Color:         d.Color,
	}
}

type UpdateUserPreferencesRequest struct {
	ReadingMode       *domain.ReadingMode       `json:"readingMode" validate:"omitempty,oneof=normal zen"`
	ZenRestoreOnOpen  *bool                     `json:"zenRestoreOnOpen"`
	ThemeMode         *domain.ThemeMode         `json:"themeMode" validate:"omitempty,oneof=light dark sepia high_contrast"`
	ThemeOverrides    map[string]string         `json:"themeOverrides"`
	TypographyProfile *domain.TypographyProfile `json:"typographyProfile" validate:"omitempty,oneof=compact comfortable large"`
}

var (
	themeKeyAllowlist = map[string]struct{}{
		"background": {},
		"text":       {},
		"accent":     {},
		"progress":   {},
	}
	hexColorRegex = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
)

func (d *UpdateUserPreferencesRequest) Validate() error {
	if err := validator.New().Struct(d); err != nil {
		return err
	}

	var errs validation.CustomValidationErrors
	for key, value := range d.ThemeOverrides {
		k := strings.TrimSpace(strings.ToLower(key))
		if _, ok := themeKeyAllowlist[k]; !ok {
			errs = append(errs, validation.CustomValidationError{Field: fmt.Sprintf("themeOverrides.%s", key), Message: "unsupported token"})
			continue
		}
		if !hexColorRegex.MatchString(strings.TrimSpace(value)) {
			errs = append(errs, validation.CustomValidationError{Field: fmt.Sprintf("themeOverrides.%s", key), Message: "must be a valid hex color"})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (d *UpdateUserPreferencesRequest) ToUsecase() *applicationdto.UpdateUserPreferencesInput {
	return &applicationdto.UpdateUserPreferencesInput{
		ReadingMode:       d.ReadingMode,
		ZenRestoreOnOpen:  d.ZenRestoreOnOpen,
		ThemeMode:         d.ThemeMode,
		ThemeOverrides:    d.ThemeOverrides,
		TypographyProfile: d.TypographyProfile,
	}
}

type UpdateUserReaderStateRequest struct {
	CurrentEbookID  *uuid.UUID          `json:"currentEbookId"`
	CurrentLocation *string             `json:"currentLocation"`
	ReadingMode     *domain.ReadingMode `json:"readingMode" validate:"omitempty,oneof=normal zen"`
	LastOpenedAt    *time.Time          `json:"lastOpenedAt"`
}

func (d *UpdateUserReaderStateRequest) Validate() error {
	return validator.New().Struct(d)
}

func (d *UpdateUserReaderStateRequest) ToUsecase() (*applicationdto.UpdateUserReaderStateInput, error) {
	input := &applicationdto.UpdateUserReaderStateInput{
		CurrentLocation: d.CurrentLocation,
		ReadingMode:     d.ReadingMode,
		LastOpenedAt:    d.LastOpenedAt,
	}
	input.CurrentEbookID = d.CurrentEbookID

	return input, nil
}
