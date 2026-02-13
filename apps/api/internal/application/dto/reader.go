package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
)

type StoreReadingProgressInput struct {
	UserID          uuid.UUID
	EbookID         uuid.UUID
	Location        string
	ProgressPercent *float64
	ReadingMode     domain.ReadingMode
	LastReadAt      *time.Time
}

func (d *StoreReadingProgressInput) ToModel() *domain.ReadingProgress {
	return &domain.ReadingProgress{
		UserID:          d.UserID,
		EbookID:         d.EbookID,
		Location:        d.Location,
		ProgressPercent: d.ProgressPercent,
		ReadingMode:     d.ReadingMode,
		LastReadAt:      d.LastReadAt,
	}
}

type UpdateReadingProgressInput struct {
	Location        *string
	ProgressPercent *float64
	ReadingMode     *domain.ReadingMode
	LastReadAt      *time.Time
}

func (d *UpdateReadingProgressInput) ToModel() *domain.ReadingProgress {
	out := &domain.ReadingProgress{}
	if d.Location != nil {
		out.Location = *d.Location
	}
	if d.ProgressPercent != nil {
		out.ProgressPercent = d.ProgressPercent
	}
	if d.ReadingMode != nil {
		out.ReadingMode = *d.ReadingMode
	}
	if d.LastReadAt != nil {
		out.LastReadAt = d.LastReadAt
	}
	return out
}

func (d *UpdateReadingProgressInput) ToMap() map[string]any {
	updates := map[string]any{}
	if d.Location != nil {
		updates["location"] = *d.Location
	}
	if d.ProgressPercent != nil {
		updates["progress_percent"] = *d.ProgressPercent
	}
	if d.ReadingMode != nil {
		updates["reading_mode"] = *d.ReadingMode
	}
	if d.LastReadAt != nil {
		updates["last_read_at"] = *d.LastReadAt
	}
	return updates
}

type StoreBookmarkInput struct {
	UserID   uuid.UUID
	EbookID  uuid.UUID
	Location string
	Label    *string
}

func (d *StoreBookmarkInput) ToModel() *domain.Bookmark {
	return &domain.Bookmark{
		UserID:   d.UserID,
		EbookID:  d.EbookID,
		Location: d.Location,
		Label:    d.Label,
	}
}

type UpdateBookmarkInput struct {
	Label *string
}

func (d *UpdateBookmarkInput) ToModel() *domain.Bookmark {
	out := &domain.Bookmark{}
	if d.Label != nil {
		out.Label = d.Label
	}
	return out
}

func (d *UpdateBookmarkInput) ToMap() map[string]any {
	updates := map[string]any{}
	if d.Label != nil {
		updates["label"] = d.Label
	}
	return updates
}

type StoreAnnotationInput struct {
	UserID        uuid.UUID
	EbookID       uuid.UUID
	LocationStart string
	LocationEnd   string
	HighlightText *string
	Note          *string
	Color         *string
}

func (d *StoreAnnotationInput) ToModel() *domain.Annotation {
	return &domain.Annotation{
		UserID:        d.UserID,
		EbookID:       d.EbookID,
		LocationStart: d.LocationStart,
		LocationEnd:   d.LocationEnd,
		HighlightText: d.HighlightText,
		Note:          d.Note,
		Color:         d.Color,
	}
}

type UpdateAnnotationInput struct {
	LocationStart *string
	LocationEnd   *string
	HighlightText *string
	Note          *string
	Color         *string
}

func (d *UpdateAnnotationInput) ToModel() *domain.Annotation {
	out := &domain.Annotation{}
	if d.LocationStart != nil {
		out.LocationStart = *d.LocationStart
	}
	if d.LocationEnd != nil {
		out.LocationEnd = *d.LocationEnd
	}
	if d.HighlightText != nil {
		out.HighlightText = d.HighlightText
	}
	if d.Note != nil {
		out.Note = d.Note
	}
	if d.Color != nil {
		out.Color = d.Color
	}
	return out
}

func (d *UpdateAnnotationInput) ToMap() map[string]any {
	updates := map[string]any{}
	if d.LocationStart != nil {
		updates["location_start"] = *d.LocationStart
	}
	if d.LocationEnd != nil {
		updates["location_end"] = *d.LocationEnd
	}
	if d.HighlightText != nil {
		updates["highlight_text"] = d.HighlightText
	}
	if d.Note != nil {
		updates["note"] = d.Note
	}
	if d.Color != nil {
		updates["color"] = d.Color
	}
	return updates
}

type UpdateUserPreferencesInput struct {
	ReadingMode       *domain.ReadingMode
	ZenRestoreOnOpen  *bool
	ThemeMode         *domain.ThemeMode
	ThemeOverrides    map[string]string
	TypographyProfile *domain.TypographyProfile
}

type UpdateUserReaderStateInput struct {
	CurrentEbookID  *uuid.UUID
	CurrentLocation *string
	ReadingMode     *domain.ReadingMode
	LastOpenedAt    *time.Time
}
