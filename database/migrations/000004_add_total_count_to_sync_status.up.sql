BEGIN;

-- Add total_count column to crawl_sync_status_counts table
ALTER TABLE crawl_sync_status_counts
ADD COLUMN total_count INT NOT NULL DEFAULT 0;

COMMENT ON COLUMN crawl_sync_status_counts.total_count IS 'Total number of unique peers in the crawl';

COMMIT;