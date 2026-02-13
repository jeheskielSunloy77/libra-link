PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS session_state (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  access_token TEXT NOT NULL,
  refresh_token TEXT NOT NULL,
  user_id TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS ui_settings (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  gutter_preset TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_preferences_cache (
  user_id TEXT PRIMARY KEY,
  reading_mode TEXT NOT NULL,
  zen_restore_on_open INTEGER NOT NULL DEFAULT 1,
  theme_mode TEXT NOT NULL,
  theme_overrides_json TEXT NOT NULL DEFAULT '{}',
  typography_profile TEXT NOT NULL,
  row_version INTEGER NOT NULL DEFAULT 1,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_reader_state_cache (
  user_id TEXT PRIMARY KEY,
  current_ebook_id TEXT,
  current_location TEXT,
  reading_mode TEXT NOT NULL,
  row_version INTEGER NOT NULL DEFAULT 1,
  last_opened_at TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS ebooks_cache (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  author TEXT,
  format TEXT,
  file_path TEXT,
  row_version INTEGER NOT NULL DEFAULT 1,
  deleted_at TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS shares_cache (
  id TEXT PRIMARY KEY,
  ebook_id TEXT NOT NULL,
  owner_id TEXT NOT NULL,
  status TEXT NOT NULL,
  title TEXT,
  borrow_until TEXT,
  row_version INTEGER NOT NULL DEFAULT 1,
  deleted_at TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS borrows_cache (
  id TEXT PRIMARY KEY,
  share_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  status TEXT NOT NULL,
  expires_at TEXT,
  row_version INTEGER NOT NULL DEFAULT 1,
  deleted_at TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS share_reviews_cache (
  id TEXT PRIMARY KEY,
  share_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  rating INTEGER,
  review TEXT,
  row_version INTEGER NOT NULL DEFAULT 1,
  deleted_at TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS share_reports_cache (
  id TEXT PRIMARY KEY,
  share_id TEXT NOT NULL,
  reporter_id TEXT NOT NULL,
  reason TEXT NOT NULL,
  details TEXT,
  row_version INTEGER NOT NULL DEFAULT 1,
  deleted_at TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS reading_progress_cache (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  ebook_id TEXT NOT NULL,
  location TEXT NOT NULL,
  progress_percent REAL,
  reading_mode TEXT NOT NULL,
  row_version INTEGER NOT NULL DEFAULT 1,
  deleted_at TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS bookmarks_cache (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  ebook_id TEXT NOT NULL,
  location TEXT NOT NULL,
  label TEXT,
  row_version INTEGER NOT NULL DEFAULT 1,
  deleted_at TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS annotations_cache (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  ebook_id TEXT NOT NULL,
  location_start TEXT NOT NULL,
  location_end TEXT NOT NULL,
  highlight_text TEXT,
  note TEXT,
  color TEXT,
  row_version INTEGER NOT NULL DEFAULT 1,
  deleted_at TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sync_events_outbox (
  id TEXT PRIMARY KEY,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  operation TEXT NOT NULL,
  payload_json TEXT,
  base_version INTEGER,
  idempotency_key TEXT NOT NULL UNIQUE,
  status TEXT NOT NULL DEFAULT 'pending',
  attempt_count INTEGER NOT NULL DEFAULT 0,
  next_attempt_at TEXT NOT NULL,
  last_error TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sync_events_outbox_status_next
  ON sync_events_outbox(status, next_attempt_at);

CREATE TABLE IF NOT EXISTS sync_checkpoint (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  last_server_timestamp TEXT,
  last_event_id TEXT,
  updated_at TEXT NOT NULL
);
