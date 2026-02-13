package sync

import (
	"context"
	"math"
	"time"

	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/api"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/storage/repo"
)

type API interface {
	StoreSyncEvent(ctx context.Context, input api.SyncEvent) error
}

type Store interface {
	ListPendingOutbox(ctx context.Context, limit int) ([]repo.OutboxEvent, error)
	MarkOutboxDone(ctx context.Context, eventID string) error
	MarkOutboxRetry(ctx context.Context, eventID string, nextAttempt time.Time, lastErr string) error
}

type Worker struct {
	store     Store
	apiClient API
	batchSize int
}

func NewWorker(store Store, apiClient API, batchSize int) *Worker {
	if batchSize <= 0 {
		batchSize = 25
	}
	return &Worker{store: store, apiClient: apiClient, batchSize: batchSize}
}

func (w *Worker) FlushOnce(ctx context.Context) error {
	events, err := w.store.ListPendingOutbox(ctx, w.batchSize)
	if err != nil {
		return err
	}

	for _, event := range events {
		baseVersion := intPtrFromInt64(event.BaseVersion)
		payload := map[string]any{}
		if event.Payload != nil {
			payload = event.Payload
		}

		err := w.apiClient.StoreSyncEvent(ctx, api.SyncEvent{
			EntityType:     event.EntityType,
			EntityID:       event.EntityID,
			Operation:      event.Operation,
			Payload:        payload,
			BaseVersion:    baseVersion,
			ClientTS:       event.CreatedAt,
			IdempotencyKey: event.IdempotencyKey,
		})
		if err != nil {
			delay := backoff(event.AttemptCount + 1)
			_ = w.store.MarkOutboxRetry(ctx, event.ID, time.Now().UTC().Add(delay), err.Error())
			continue
		}
		_ = w.store.MarkOutboxDone(ctx, event.ID)
	}

	return nil
}

func (w *Worker) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 10 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	_ = w.FlushOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = w.FlushOnce(ctx)
		}
	}
}

func intPtrFromInt64(value *int64) *int {
	if value == nil {
		return nil
	}
	v := int(*value)
	return &v
}

func backoff(attempt int64) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 6 {
		attempt = 6
	}
	seconds := math.Pow(2, float64(attempt))
	return time.Duration(seconds) * time.Second
}
