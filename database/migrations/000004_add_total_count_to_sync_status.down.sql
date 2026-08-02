BEGIN;

-- Remove total_count column from crawl_sync_status_counts table
ALTER TABLE crawl_sync_status_counts
DROP COLUMN total_count;

COMMIT;