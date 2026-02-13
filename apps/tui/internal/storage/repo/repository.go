package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/api"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/storage/sqlite/sqlcdb"
)

type Repository struct {
	q *sqlcdb.Queries
}

type SessionState struct {
	AccessToken  string
	RefreshToken string
	UserID       string
	UpdatedAt    time.Time
}

type PreferencesCache struct {
	UserID            string
	ReadingMode       string
	ZenRestoreOnOpen  bool
	ThemeMode         string
	ThemeOverrides    map[string]string
	TypographyProfile string
	RowVersion        int64
	UpdatedAt         time.Time
}

type ReaderStateCache struct {
	UserID          string
	CurrentEbookID  string
	CurrentLocation string
	ReadingMode     string
	RowVersion      int64
	LastOpenedAt    *time.Time
	UpdatedAt       time.Time
}

type EbookCache struct {
	ID         string
	Title      string
	Author     string
	Format     string
	FilePath   string
	RowVersion int64
	UpdatedAt  time.Time
}

type ShareCache struct {
	ID          string
	EbookID     string
	OwnerID     string
	Status      string
	Title       string
	BorrowUntil *time.Time
	RowVersion  int64
	UpdatedAt   time.Time
}

type OutboxEvent struct {
	ID             string
	EntityType     string
	EntityID       string
	Operation      string
	Payload        map[string]any
	BaseVersion    *int64
	IdempotencyKey string
	AttemptCount   int64
	NextAttemptAt  time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type SyncCheckpoint struct {
	LastServerTimestamp *time.Time
	LastEventID         string
	UpdatedAt           time.Time
}

func New(q *sqlcdb.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) GetSessionState(ctx context.Context) (*SessionState, error) {
	row, err := r.q.GetSessionState(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	updatedAt, _ := time.Parse(time.RFC3339Nano, row.UpdatedAt)
	state := &SessionState{
		AccessToken:  row.AccessToken,
		RefreshToken: row.RefreshToken,
		UpdatedAt:    updatedAt,
	}
	if row.UserID.Valid {
		state.UserID = row.UserID.String
	}
	return state, nil
}

func (r *Repository) UpsertSessionState(ctx context.Context, state SessionState) error {
	updatedAt := state.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	return r.q.UpsertSessionState(ctx, sqlcdb.UpsertSessionStateParams{
		AccessToken:  state.AccessToken,
		RefreshToken: state.RefreshToken,
		UserID:       nullString(state.UserID),
		UpdatedAt:    updatedAt.Format(time.RFC3339Nano),
	})
}

func (r *Repository) ClearSessionState(ctx context.Context) error {
	return r.q.ClearSessionState(ctx)
}

func (r *Repository) GetPreferences(ctx context.Context, userID string) (*PreferencesCache, error) {
	row, err := r.q.GetUserPreferencesCache(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	overrides := map[string]string{}
	if row.ThemeOverridesJson != "" {
		_ = json.Unmarshal([]byte(row.ThemeOverridesJson), &overrides)
	}

	updatedAt, _ := time.Parse(time.RFC3339Nano, row.UpdatedAt)
	return &PreferencesCache{
		UserID:            row.UserID,
		ReadingMode:       row.ReadingMode,
		ZenRestoreOnOpen:  row.ZenRestoreOnOpen > 0,
		ThemeMode:         row.ThemeMode,
		ThemeOverrides:    overrides,
		TypographyProfile: row.TypographyProfile,
		RowVersion:        row.RowVersion,
		UpdatedAt:         updatedAt,
	}, nil
}

func (r *Repository) UpsertPreferences(ctx context.Context, prefs PreferencesCache) error {
	updatedAt := prefs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	overridesBytes, err := json.Marshal(prefs.ThemeOverrides)
	if err != nil {
		return err
	}

	zen := 0
	if prefs.ZenRestoreOnOpen {
		zen = 1
	}

	return r.q.UpsertUserPreferencesCache(ctx, sqlcdb.UpsertUserPreferencesCacheParams{
		UserID:             prefs.UserID,
		ReadingMode:        prefs.ReadingMode,
		ZenRestoreOnOpen:   int64(zen),
		ThemeMode:          prefs.ThemeMode,
		ThemeOverridesJson: string(overridesBytes),
		TypographyProfile:  prefs.TypographyProfile,
		RowVersion:         prefs.RowVersion,
		UpdatedAt:          updatedAt.Format(time.RFC3339Nano),
	})
}

func (r *Repository) GetReaderState(ctx context.Context, userID string) (*ReaderStateCache, error) {
	row, err := r.q.GetUserReaderStateCache(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	updatedAt, _ := time.Parse(time.RFC3339Nano, row.UpdatedAt)
	result := &ReaderStateCache{
		UserID:      row.UserID,
		ReadingMode: row.ReadingMode,
		RowVersion:  row.RowVersion,
		UpdatedAt:   updatedAt,
	}
	if row.CurrentEbookID.Valid {
		result.CurrentEbookID = row.CurrentEbookID.String
	}
	if row.CurrentLocation.Valid {
		result.CurrentLocation = row.CurrentLocation.String
	}
	if row.LastOpenedAt.Valid {
		if ts, err := time.Parse(time.RFC3339Nano, row.LastOpenedAt.String); err == nil {
			result.LastOpenedAt = &ts
		}
	}
	return result, nil
}

func (r *Repository) UpsertReaderState(ctx context.Context, state ReaderStateCache) error {
	updatedAt := state.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	var lastOpened sql.NullString
	if state.LastOpenedAt != nil {
		lastOpened = sql.NullString{String: state.LastOpenedAt.Format(time.RFC3339Nano), Valid: true}
	}

	return r.q.UpsertUserReaderStateCache(ctx, sqlcdb.UpsertUserReaderStateCacheParams{
		UserID:          state.UserID,
		CurrentEbookID:  nullString(state.CurrentEbookID),
		CurrentLocation: nullString(state.CurrentLocation),
		ReadingMode:     state.ReadingMode,
		RowVersion:      state.RowVersion,
		LastOpenedAt:    lastOpened,
		UpdatedAt:       updatedAt.Format(time.RFC3339Nano),
	})
}

func (r *Repository) UpsertEbooksFromRemote(ctx context.Context, ebooks []api.Ebook) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, item := range ebooks {
		if err := r.q.UpsertEbookCache(ctx, sqlcdb.UpsertEbookCacheParams{
			ID:         item.ID,
			Title:      item.Title,
			Author:     sql.NullString{},
			Format:     nullString(item.Format),
			FilePath:   sql.NullString{String: item.StorageKey, Valid: item.StorageKey != ""},
			RowVersion: 1,
			DeletedAt:  sql.NullString{},
			UpdatedAt:  now,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ListEbooks(ctx context.Context, query string) ([]EbookCache, error) {
	var rows []*sqlcdb.EbooksCache
	var err error
	if query == "" {
		rows, err = r.q.ListActiveEbooksCache(ctx)
	} else {
		rows, err = r.q.SearchActiveEbooksCache(ctx, query)
	}
	if err != nil {
		return nil, err
	}

	result := make([]EbookCache, 0, len(rows))
	for _, row := range rows {
		updatedAt, _ := time.Parse(time.RFC3339Nano, row.UpdatedAt)
		result = append(result, EbookCache{
			ID:         row.ID,
			Title:      row.Title,
			Author:     row.Author.String,
			Format:     row.Format.String,
			FilePath:   row.FilePath.String,
			RowVersion: row.RowVersion,
			UpdatedAt:  updatedAt,
		})
	}
	return result, nil
}

func (r *Repository) UpsertSharesFromRemote(ctx context.Context, shares []api.Share) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, item := range shares {
		if err := r.q.UpsertShareCache(ctx, sqlcdb.UpsertShareCacheParams{
			ID:          item.ID,
			EbookID:     item.EbookID,
			OwnerID:     item.OwnerUserID,
			Status:      item.Status,
			Title:       nullString(item.Title),
			BorrowUntil: sql.NullString{},
			RowVersion:  1,
			DeletedAt:   sql.NullString{},
			UpdatedAt:   now,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ListShares(ctx context.Context) ([]ShareCache, error) {
	rows, err := r.q.ListActiveSharesCache(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]ShareCache, 0, len(rows))
	for _, row := range rows {
		updatedAt, _ := time.Parse(time.RFC3339Nano, row.UpdatedAt)
		item := ShareCache{
			ID:         row.ID,
			EbookID:    row.EbookID,
			OwnerID:    row.OwnerID,
			Status:     row.Status,
			Title:      row.Title.String,
			RowVersion: row.RowVersion,
			UpdatedAt:  updatedAt,
		}
		if row.BorrowUntil.Valid {
			if ts, err := time.Parse(time.RFC3339Nano, row.BorrowUntil.String); err == nil {
				item.BorrowUntil = &ts
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func (r *Repository) EnqueueOutbox(ctx context.Context, event OutboxEvent) error {
	now := time.Now().UTC()
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	if event.IdempotencyKey == "" {
		event.IdempotencyKey = uuid.NewString()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	}
	if event.UpdatedAt.IsZero() {
		event.UpdatedAt = now
	}
	if event.NextAttemptAt.IsZero() {
		event.NextAttemptAt = now
	}

	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	params := sqlcdb.EnqueueSyncOutboxEventParams{
		ID:             event.ID,
		EntityType:     event.EntityType,
		EntityID:       event.EntityID,
		Operation:      event.Operation,
		PayloadJson:    sql.NullString{String: string(payloadBytes), Valid: len(payloadBytes) > 0 && string(payloadBytes) != "null"},
		IdempotencyKey: event.IdempotencyKey,
		NextAttemptAt:  event.NextAttemptAt.Format(time.RFC3339Nano),
		CreatedAt:      event.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:      event.UpdatedAt.Format(time.RFC3339Nano),
	}
	if event.BaseVersion != nil {
		params.BaseVersion = sql.NullInt64{Int64: *event.BaseVersion, Valid: true}
	}
	return r.q.EnqueueSyncOutboxEvent(ctx, params)
}

func (r *Repository) ListPendingOutbox(ctx context.Context, limit int) ([]OutboxEvent, error) {
	if limit <= 0 {
		limit = 25
	}
	rows, err := r.q.ListPendingSyncOutboxEvents(ctx, sqlcdb.ListPendingSyncOutboxEventsParams{
		Now:       time.Now().UTC().Format(time.RFC3339Nano),
		LimitRows: int64(limit),
	})
	if err != nil {
		return nil, err
	}

	items := make([]OutboxEvent, 0, len(rows))
	for _, row := range rows {
		createdAt, _ := time.Parse(time.RFC3339Nano, row.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339Nano, row.UpdatedAt)
		nextAttempt, _ := time.Parse(time.RFC3339Nano, row.NextAttemptAt)
		item := OutboxEvent{
			ID:             row.ID,
			EntityType:     row.EntityType,
			EntityID:       row.EntityID,
			Operation:      row.Operation,
			IdempotencyKey: row.IdempotencyKey,
			AttemptCount:   row.AttemptCount,
			NextAttemptAt:  nextAttempt,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		}
		if row.BaseVersion.Valid {
			v := row.BaseVersion.Int64
			item.BaseVersion = &v
		}
		if row.PayloadJson.Valid {
			var payload map[string]any
			if err := json.Unmarshal([]byte(row.PayloadJson.String), &payload); err == nil {
				item.Payload = payload
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) MarkOutboxDone(ctx context.Context, eventID string) error {
	return r.q.MarkSyncOutboxEventSucceeded(ctx, sqlcdb.MarkSyncOutboxEventSucceededParams{
		ID:        eventID,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func (r *Repository) MarkOutboxRetry(ctx context.Context, eventID string, nextAttempt time.Time, lastErr string) error {
	if nextAttempt.IsZero() {
		nextAttempt = time.Now().UTC().Add(10 * time.Second)
	}
	return r.q.MarkSyncOutboxEventRetry(ctx, sqlcdb.MarkSyncOutboxEventRetryParams{
		NextAttemptAt: nextAttempt.Format(time.RFC3339Nano),
		UpdatedAt:     time.Now().UTC().Format(time.RFC3339Nano),
		LastError:     nullString(lastErr),
		ID:            eventID,
	})
}

func (r *Repository) GetSyncCheckpoint(ctx context.Context) (*SyncCheckpoint, error) {
	row, err := r.q.GetSyncCheckpoint(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	updatedAt, _ := time.Parse(time.RFC3339Nano, row.UpdatedAt)
	item := &SyncCheckpoint{UpdatedAt: updatedAt}
	if row.LastEventID.Valid {
		item.LastEventID = row.LastEventID.String
	}
	if row.LastServerTimestamp.Valid {
		if ts, err := time.Parse(time.RFC3339Nano, row.LastServerTimestamp.String); err == nil {
			item.LastServerTimestamp = &ts
		}
	}
	return item, nil
}

func (r *Repository) UpsertSyncCheckpoint(ctx context.Context, checkpoint SyncCheckpoint) error {
	params := sqlcdb.UpsertSyncCheckpointParams{
		LastServerTimestamp: sql.NullString{},
		LastEventID:         nullString(checkpoint.LastEventID),
		UpdatedAt:           time.Now().UTC().Format(time.RFC3339Nano),
	}
	if checkpoint.LastServerTimestamp != nil {
		params.LastServerTimestamp = sql.NullString{String: checkpoint.LastServerTimestamp.Format(time.RFC3339Nano), Valid: true}
	}
	return r.q.UpsertSyncCheckpoint(ctx, params)
}

func nullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}
