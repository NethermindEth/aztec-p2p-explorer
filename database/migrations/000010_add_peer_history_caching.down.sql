BEGIN;

-- Drop daily peer history table
DROP TABLE IF EXISTS daily_peer_history;

-- Remove ccv_filtered_peer_count column from crawls table
ALTER TABLE crawls DROP COLUMN IF EXISTS ccv_filtered_peer_count;

COMMIT;
