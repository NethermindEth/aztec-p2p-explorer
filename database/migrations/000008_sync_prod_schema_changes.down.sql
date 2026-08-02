BEGIN;

-- Revert the changes made in 000008_sync_prod_schema_changes.up.sql
-- WARNING: This rollback will fail if there are peer IDs > 2147483647 (integer max)

-- 1. Drop added indexes
DROP INDEX IF EXISTS idx_peer_visits_index_crawl_id;
DROP INDEX IF EXISTS idx_peer_visits_index_crawl_peer;
DROP INDEX IF EXISTS idx_peer_visits_index_latest_crawl;
DROP INDEX IF EXISTS idx_peer_visits_index_peer_id;
DROP INDEX IF EXISTS idx_peer_visits_index_visit_id;
DROP INDEX IF EXISTS idx_peers_last_seen;
DROP INDEX IF EXISTS idx_peers_states_peer_created_desc;

-- 2. Replace new_id_unique with peers_pkey PRIMARY KEY
DROP INDEX IF EXISTS new_id_unique;

-- Ensure peers.id is NOT NULL before creating PRIMARY KEY
ALTER TABLE peers ALTER COLUMN id SET NOT NULL;

-- Create PRIMARY KEY
ALTER TABLE peers ADD CONSTRAINT peers_pkey PRIMARY KEY (id);

-- 3. Change bigint columns back to integer
-- Drop foreign key constraints first
ALTER TABLE neighbors DROP CONSTRAINT IF EXISTS fk_neighbors_peer_id;
ALTER TABLE peer_visits_index DROP CONSTRAINT IF EXISTS peer_visits_index_peer_id_fkey;
ALTER TABLE peers_states DROP CONSTRAINT IF EXISTS fk_peers_states_peer_id;
ALTER TABLE peers_x_multi_addresses DROP CONSTRAINT IF EXISTS fk_peers_x_multi_addresses_peer_id;
ALTER TABLE visits DROP CONSTRAINT IF EXISTS fk_visits_peer_id;

-- Change columns back to integer
ALTER TABLE peers ALTER COLUMN id TYPE integer;
ALTER TABLE neighbors ALTER COLUMN peer_id TYPE integer;
ALTER TABLE peer_visits_index ALTER COLUMN peer_id TYPE integer;
-- Note: crawl_id stays as integer to match crawls.id
ALTER TABLE peer_visits_index ALTER COLUMN visit_id TYPE integer;
ALTER TABLE visits ALTER COLUMN id TYPE integer;
ALTER TABLE peers_states ALTER COLUMN peer_id TYPE integer;
ALTER TABLE peers_x_multi_addresses ALTER COLUMN peer_id TYPE integer;
ALTER TABLE visits ALTER COLUMN peer_id TYPE integer;

-- Recreate foreign key constraints with ON DELETE SET NULL
ALTER TABLE neighbors ADD CONSTRAINT fk_neighbors_peer_id
    FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE;

ALTER TABLE peer_visits_index ADD CONSTRAINT peer_visits_index_peer_id_fkey
    FOREIGN KEY (peer_id) REFERENCES peers (id);

ALTER TABLE peers_states ADD CONSTRAINT fk_peers_states_peer_id
    FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE;

ALTER TABLE peers_x_multi_addresses ADD CONSTRAINT fk_peers_x_multi_addresses_peer_id
    FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE;

ALTER TABLE visits ADD CONSTRAINT fk_visits_peer_id
    FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE SET NULL;

COMMIT;
