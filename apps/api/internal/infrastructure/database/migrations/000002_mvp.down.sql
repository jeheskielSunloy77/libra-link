DROP TABLE IF EXISTS sync_checkpoints;

DROP INDEX IF EXISTS idx_sync_events_user_server_timestamp_asc;
DROP TABLE IF EXISTS sync_events;

DROP INDEX IF EXISTS idx_share_reports_status_created_at_desc;
DROP TABLE IF EXISTS share_reports;

DROP INDEX IF EXISTS uq_share_reviews_share_user_active;
DROP INDEX IF EXISTS idx_share_reviews_share_created_at_desc;
DROP TABLE IF EXISTS share_reviews;

DROP TRIGGER IF EXISTS trg_borrows_prevent_self_borrow ON borrows;
DROP FUNCTION IF EXISTS prevent_self_borrow;
DROP INDEX IF EXISTS idx_borrows_active_due_at;
DROP INDEX IF EXISTS uq_borrows_share_borrower_active;
DROP INDEX IF EXISTS idx_borrows_borrower_status_due_at;
DROP TABLE IF EXISTS borrows;

DROP INDEX IF EXISTS idx_shares_status_visibility_created_at_desc;
DROP TABLE IF EXISTS shares;

DROP TABLE IF EXISTS annotations;

DROP INDEX IF EXISTS uq_bookmarks_user_ebook_location_active;
DROP TABLE IF EXISTS bookmarks;

DROP INDEX IF EXISTS uq_reading_progress_user_ebook_active;
DROP TABLE IF EXISTS reading_progress;

DROP TABLE IF EXISTS user_reader_state;
DROP TABLE IF EXISTS user_preferences;

DROP TABLE IF EXISTS ebook_google_metadata;
DROP INDEX IF EXISTS idx_ebook_tags_tag_ebook;
DROP TABLE IF EXISTS ebook_tags;
DROP INDEX IF EXISTS idx_ebook_authors_author_ebook;
DROP TABLE IF EXISTS ebook_authors;

DROP INDEX IF EXISTS idx_tags_name_active;
DROP TABLE IF EXISTS tags;
DROP INDEX IF EXISTS idx_authors_name_active;
DROP TABLE IF EXISTS authors;

DROP INDEX IF EXISTS idx_ebooks_format;
DROP INDEX IF EXISTS idx_ebooks_owner_created_at_desc;
DROP TABLE IF EXISTS ebooks;

DROP TYPE IF EXISTS sync_operation;
DROP TYPE IF EXISTS sync_entity_type;
DROP TYPE IF EXISTS report_status;
DROP TYPE IF EXISTS report_reason;
DROP TYPE IF EXISTS borrow_status;
DROP TYPE IF EXISTS share_status;
DROP TYPE IF EXISTS share_visibility;
DROP TYPE IF EXISTS typography_profile;
DROP TYPE IF EXISTS theme_mode;
DROP TYPE IF EXISTS reading_mode;
DROP TYPE IF EXISTS ebook_format;
