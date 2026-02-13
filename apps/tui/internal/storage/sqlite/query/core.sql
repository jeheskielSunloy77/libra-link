-- name: UpsertSessionState :exec
INSERT INTO session_state (
  id,
  access_token,
  refresh_token,
  user_id,
  updated_at
) VALUES (
  1,
  sqlc.arg(access_token),
  sqlc.arg(refresh_token),
  sqlc.narg(user_id),
  sqlc.arg(updated_at)
)
ON CONFLICT(id) DO UPDATE SET
  access_token = excluded.access_token,
  refresh_token = excluded.refresh_token,
  user_id = excluded.user_id,
  updated_at = excluded.updated_at;

-- name: GetSessionState :one
SELECT access_token, refresh_token, user_id, updated_at
FROM session_state
WHERE id = 1;

-- name: ClearSessionState :exec
DELETE FROM session_state WHERE id = 1;

-- name: UpsertUISettings :exec
INSERT INTO ui_settings (
  id,
  gutter_preset,
  updated_at
) VALUES (
  1,
  sqlc.arg(gutter_preset),
  sqlc.arg(updated_at)
)
ON CONFLICT(id) DO UPDATE SET
  gutter_preset = excluded.gutter_preset,
  updated_at = excluded.updated_at;

-- name: GetUISettings :one
SELECT gutter_preset, updated_at
FROM ui_settings
WHERE id = 1;

-- name: UpsertUserPreferencesCache :exec
INSERT INTO user_preferences_cache (
  user_id,
  reading_mode,
  zen_restore_on_open,
  theme_mode,
  theme_overrides_json,
  typography_profile,
  row_version,
  updated_at
) VALUES (
  sqlc.arg(user_id),
  sqlc.arg(reading_mode),
  sqlc.arg(zen_restore_on_open),
  sqlc.arg(theme_mode),
  sqlc.arg(theme_overrides_json),
  sqlc.arg(typography_profile),
  sqlc.arg(row_version),
  sqlc.arg(updated_at)
)
ON CONFLICT(user_id) DO UPDATE SET
  reading_mode = excluded.reading_mode,
  zen_restore_on_open = excluded.zen_restore_on_open,
  theme_mode = excluded.theme_mode,
  theme_overrides_json = excluded.theme_overrides_json,
  typography_profile = excluded.typography_profile,
  row_version = excluded.row_version,
  updated_at = excluded.updated_at;

-- name: GetUserPreferencesCache :one
SELECT user_id, reading_mode, zen_restore_on_open, theme_mode, theme_overrides_json, typography_profile, row_version, updated_at
FROM user_preferences_cache
WHERE user_id = sqlc.arg(user_id)
LIMIT 1;

-- name: UpsertUserReaderStateCache :exec
INSERT INTO user_reader_state_cache (
  user_id,
  current_ebook_id,
  current_location,
  reading_mode,
  row_version,
  last_opened_at,
  updated_at
) VALUES (
  sqlc.arg(user_id),
  sqlc.narg(current_ebook_id),
  sqlc.narg(current_location),
  sqlc.arg(reading_mode),
  sqlc.arg(row_version),
  sqlc.narg(last_opened_at),
  sqlc.arg(updated_at)
)
ON CONFLICT(user_id) DO UPDATE SET
  current_ebook_id = excluded.current_ebook_id,
  current_location = excluded.current_location,
  reading_mode = excluded.reading_mode,
  row_version = excluded.row_version,
  last_opened_at = excluded.last_opened_at,
  updated_at = excluded.updated_at;

-- name: GetUserReaderStateCache :one
SELECT user_id, current_ebook_id, current_location, reading_mode, row_version, last_opened_at, updated_at
FROM user_reader_state_cache
WHERE user_id = sqlc.arg(user_id)
LIMIT 1;

-- name: UpsertSyncCheckpoint :exec
INSERT INTO sync_checkpoint (
  id,
  last_server_timestamp,
  last_event_id,
  updated_at
) VALUES (
  1,
  sqlc.narg(last_server_timestamp),
  sqlc.narg(last_event_id),
  sqlc.arg(updated_at)
)
ON CONFLICT(id) DO UPDATE SET
  last_server_timestamp = excluded.last_server_timestamp,
  last_event_id = excluded.last_event_id,
  updated_at = excluded.updated_at;

-- name: GetSyncCheckpoint :one
SELECT last_server_timestamp, last_event_id, updated_at
FROM sync_checkpoint
WHERE id = 1;

-- name: EnqueueSyncOutboxEvent :exec
INSERT INTO sync_events_outbox (
  id,
  entity_type,
  entity_id,
  operation,
  payload_json,
  base_version,
  idempotency_key,
  status,
  attempt_count,
  next_attempt_at,
  last_error,
  created_at,
  updated_at
) VALUES (
  sqlc.arg(id),
  sqlc.arg(entity_type),
  sqlc.arg(entity_id),
  sqlc.arg(operation),
  sqlc.narg(payload_json),
  sqlc.narg(base_version),
  sqlc.arg(idempotency_key),
  'pending',
  0,
  sqlc.arg(next_attempt_at),
  NULL,
  sqlc.arg(created_at),
  sqlc.arg(updated_at)
)
ON CONFLICT(idempotency_key) DO NOTHING;

-- name: ListPendingSyncOutboxEvents :many
SELECT id,
  entity_type,
  entity_id,
  operation,
  payload_json,
  base_version,
  idempotency_key,
  status,
  attempt_count,
  next_attempt_at,
  last_error,
  created_at,
  updated_at
FROM sync_events_outbox
WHERE status = 'pending'
  AND next_attempt_at <= sqlc.arg(now)
ORDER BY created_at ASC
LIMIT sqlc.arg(limit_rows);

-- name: MarkSyncOutboxEventSucceeded :exec
UPDATE sync_events_outbox
SET status = 'done',
  updated_at = sqlc.arg(updated_at),
  last_error = NULL
WHERE id = sqlc.arg(id);

-- name: MarkSyncOutboxEventRetry :exec
UPDATE sync_events_outbox
SET attempt_count = attempt_count + 1,
  next_attempt_at = sqlc.arg(next_attempt_at),
  updated_at = sqlc.arg(updated_at),
  last_error = sqlc.arg(last_error)
WHERE id = sqlc.arg(id);

-- name: UpsertEbookCache :exec
INSERT INTO ebooks_cache (
  id,
  title,
  author,
  format,
  file_path,
  row_version,
  deleted_at,
  updated_at
) VALUES (
  sqlc.arg(id),
  sqlc.arg(title),
  sqlc.narg(author),
  sqlc.narg(format),
  sqlc.narg(file_path),
  sqlc.arg(row_version),
  sqlc.narg(deleted_at),
  sqlc.arg(updated_at)
)
ON CONFLICT(id) DO UPDATE SET
  title = excluded.title,
  author = excluded.author,
  format = excluded.format,
  file_path = excluded.file_path,
  row_version = excluded.row_version,
  deleted_at = excluded.deleted_at,
  updated_at = excluded.updated_at;

-- name: ListActiveEbooksCache :many
SELECT id, title, author, format, file_path, row_version, deleted_at, updated_at
FROM ebooks_cache
WHERE deleted_at IS NULL
ORDER BY updated_at DESC;

-- name: SearchActiveEbooksCache :many
SELECT id, title, author, format, file_path, row_version, deleted_at, updated_at
FROM ebooks_cache
WHERE deleted_at IS NULL
  AND (
    lower(title) LIKE '%' || lower(sqlc.arg(search_term)) || '%'
    OR lower(COALESCE(author, '')) LIKE '%' || lower(sqlc.arg(search_term)) || '%'
  )
ORDER BY updated_at DESC;

-- name: UpsertShareCache :exec
INSERT INTO shares_cache (
  id,
  ebook_id,
  owner_id,
  status,
  title,
  borrow_until,
  row_version,
  deleted_at,
  updated_at
) VALUES (
  sqlc.arg(id),
  sqlc.arg(ebook_id),
  sqlc.arg(owner_id),
  sqlc.arg(status),
  sqlc.narg(title),
  sqlc.narg(borrow_until),
  sqlc.arg(row_version),
  sqlc.narg(deleted_at),
  sqlc.arg(updated_at)
)
ON CONFLICT(id) DO UPDATE SET
  ebook_id = excluded.ebook_id,
  owner_id = excluded.owner_id,
  status = excluded.status,
  title = excluded.title,
  borrow_until = excluded.borrow_until,
  row_version = excluded.row_version,
  deleted_at = excluded.deleted_at,
  updated_at = excluded.updated_at;

-- name: ListActiveSharesCache :many
SELECT id, ebook_id, owner_id, status, title, borrow_until, row_version, deleted_at, updated_at
FROM shares_cache
WHERE deleted_at IS NULL
ORDER BY updated_at DESC;
