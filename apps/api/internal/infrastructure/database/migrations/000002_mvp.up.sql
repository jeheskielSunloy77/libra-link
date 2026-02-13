DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ebook_format') THEN
        CREATE TYPE ebook_format AS ENUM ('epub', 'pdf', 'txt');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'reading_mode') THEN
        CREATE TYPE reading_mode AS ENUM ('normal', 'zen');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'theme_mode') THEN
        CREATE TYPE theme_mode AS ENUM ('light', 'dark', 'sepia', 'high_contrast');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'typography_profile') THEN
        CREATE TYPE typography_profile AS ENUM ('compact', 'comfortable', 'large');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'share_visibility') THEN
        CREATE TYPE share_visibility AS ENUM ('public', 'unlisted');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'share_status') THEN
        CREATE TYPE share_status AS ENUM ('active', 'disabled', 'removed');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'borrow_status') THEN
        CREATE TYPE borrow_status AS ENUM ('active', 'returned', 'expired', 'revoked');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'report_reason') THEN
        CREATE TYPE report_reason AS ENUM ('copyright', 'abuse', 'spam', 'other');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'report_status') THEN
        CREATE TYPE report_status AS ENUM ('open', 'in_review', 'resolved', 'rejected');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'sync_entity_type') THEN
        CREATE TYPE sync_entity_type AS ENUM ('progress', 'annotation', 'bookmark', 'preference', 'reader_state');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'sync_operation') THEN
        CREATE TYPE sync_operation AS ENUM ('upsert', 'delete');
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS ebooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    format ebook_format NOT NULL,
    language_code VARCHAR(16),
    storage_key TEXT NOT NULL UNIQUE,
    file_size_bytes BIGINT NOT NULL,
    checksum_sha256 VARCHAR(64) NOT NULL,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_ebooks_owner_id UNIQUE (owner_user_id, id)
);

CREATE INDEX IF NOT EXISTS idx_ebooks_owner_created_at_desc ON ebooks (owner_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ebooks_format ON ebooks (format);

CREATE TABLE IF NOT EXISTS authors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_authors_name_active ON authors (name) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_name_active ON tags (name) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS ebook_authors (
    ebook_id UUID NOT NULL REFERENCES ebooks(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
    PRIMARY KEY (ebook_id, author_id)
);

CREATE INDEX IF NOT EXISTS idx_ebook_authors_author_ebook ON ebook_authors (author_id, ebook_id);

CREATE TABLE IF NOT EXISTS ebook_tags (
    ebook_id UUID NOT NULL REFERENCES ebooks(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (ebook_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_ebook_tags_tag_ebook ON ebook_tags (tag_id, ebook_id);

CREATE TABLE IF NOT EXISTS ebook_google_metadata (
    ebook_id UUID PRIMARY KEY REFERENCES ebooks(id) ON DELETE CASCADE,
    google_books_id TEXT NOT NULL,
    isbn_10 VARCHAR(10),
    isbn_13 VARCHAR(13),
    publisher TEXT,
    published_date TEXT,
    page_count INT,
    categories JSONB,
    thumbnail_url TEXT,
    info_link TEXT,
    raw_payload JSONB,
    attached_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    reading_mode reading_mode NOT NULL DEFAULT 'normal',
    zen_restore_on_open BOOLEAN NOT NULL DEFAULT TRUE,
    theme_mode theme_mode NOT NULL DEFAULT 'dark',
    theme_overrides JSONB NOT NULL DEFAULT '{}'::jsonb,
    typography_profile typography_profile NOT NULL DEFAULT 'comfortable',
    row_version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_reader_state (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    current_ebook_id UUID REFERENCES ebooks(id),
    current_location TEXT,
    reading_mode reading_mode NOT NULL DEFAULT 'normal',
    row_version BIGINT NOT NULL DEFAULT 1,
    last_opened_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reading_progress (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ebook_id UUID NOT NULL REFERENCES ebooks(id) ON DELETE CASCADE,
    location TEXT NOT NULL,
    progress_percent NUMERIC(5,2),
    reading_mode reading_mode NOT NULL,
    row_version BIGINT NOT NULL DEFAULT 1,
    last_read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_reading_progress_percent CHECK (progress_percent IS NULL OR (progress_percent >= 0 AND progress_percent <= 100))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_reading_progress_user_ebook_active ON reading_progress (user_id, ebook_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS bookmarks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ebook_id UUID NOT NULL REFERENCES ebooks(id) ON DELETE CASCADE,
    location TEXT NOT NULL,
    label TEXT,
    row_version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_bookmarks_user_ebook_location_active ON bookmarks (user_id, ebook_id, location) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS annotations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ebook_id UUID NOT NULL REFERENCES ebooks(id) ON DELETE CASCADE,
    location_start TEXT NOT NULL,
    location_end TEXT NOT NULL,
    highlight_text TEXT,
    note TEXT,
    color VARCHAR(32),
    row_version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS shares (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ebook_id UUID NOT NULL REFERENCES ebooks(id) ON DELETE CASCADE,
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title_override TEXT,
    description TEXT,
    visibility share_visibility NOT NULL DEFAULT 'public',
    status share_status NOT NULL DEFAULT 'active',
    borrow_duration_hours INT NOT NULL,
    max_concurrent_borrows INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_shares_owner_ebook FOREIGN KEY (owner_user_id, ebook_id) REFERENCES ebooks(owner_user_id, id) ON DELETE CASCADE,
    CONSTRAINT chk_shares_borrow_duration CHECK (borrow_duration_hours > 0),
    CONSTRAINT chk_shares_max_concurrent CHECK (max_concurrent_borrows > 0)
);

CREATE INDEX IF NOT EXISTS idx_shares_status_visibility_created_at_desc ON shares (status, visibility, created_at DESC);

CREATE TABLE IF NOT EXISTS borrows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    share_id UUID NOT NULL REFERENCES shares(id) ON DELETE CASCADE,
    borrower_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL,
    due_at TIMESTAMPTZ NOT NULL,
    returned_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    status borrow_status NOT NULL,
    legal_acknowledged_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_borrows_due_after_start CHECK (due_at >= started_at),
    CONSTRAINT chk_borrows_returned_after_start CHECK (returned_at IS NULL OR returned_at >= started_at),
    CONSTRAINT chk_borrows_expired_after_due CHECK (expired_at IS NULL OR expired_at >= due_at)
);

CREATE INDEX IF NOT EXISTS idx_borrows_borrower_status_due_at ON borrows (borrower_user_id, status, due_at);
CREATE UNIQUE INDEX IF NOT EXISTS uq_borrows_share_borrower_active ON borrows (share_id, borrower_user_id) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_borrows_active_due_at ON borrows (status, due_at) WHERE status = 'active';

CREATE OR REPLACE FUNCTION prevent_self_borrow()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM shares s
        WHERE s.id = NEW.share_id
          AND s.owner_user_id = NEW.borrower_user_id
    ) THEN
        RAISE EXCEPTION 'share owner cannot borrow own share' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_borrows_prevent_self_borrow ON borrows;
CREATE TRIGGER trg_borrows_prevent_self_borrow
BEFORE INSERT OR UPDATE ON borrows
FOR EACH ROW
EXECUTE FUNCTION prevent_self_borrow();

CREATE TABLE IF NOT EXISTS share_reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    share_id UUID NOT NULL REFERENCES shares(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating SMALLINT NOT NULL,
    review_text TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_share_reviews_rating CHECK (rating >= 1 AND rating <= 5)
);

CREATE INDEX IF NOT EXISTS idx_share_reviews_share_created_at_desc ON share_reviews (share_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS uq_share_reviews_share_user_active ON share_reviews (share_id, user_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS share_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    share_id UUID NOT NULL REFERENCES shares(id) ON DELETE CASCADE,
    reporter_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason report_reason NOT NULL,
    details TEXT,
    status report_status NOT NULL DEFAULT 'open',
    reviewed_by_user_id UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    resolution_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_share_reports_status_created_at_desc ON share_reports (status, created_at DESC);

CREATE TABLE IF NOT EXISTS sync_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entity_type sync_entity_type NOT NULL,
    entity_id UUID NOT NULL,
    operation sync_operation NOT NULL,
    payload JSONB,
    base_version BIGINT,
    client_timestamp TIMESTAMPTZ NOT NULL,
    server_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    idempotency_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_sync_events_user_id_event_id UNIQUE (user_id, id),
    CONSTRAINT uq_sync_events_user_idempotency_key UNIQUE (user_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_sync_events_user_server_timestamp_asc ON sync_events (user_id, server_timestamp ASC);

CREATE TABLE IF NOT EXISTS sync_checkpoints (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    last_server_timestamp TIMESTAMPTZ NOT NULL,
    last_event_id UUID,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_sync_checkpoints_event FOREIGN KEY (user_id, last_event_id) REFERENCES sync_events(user_id, id)
);
