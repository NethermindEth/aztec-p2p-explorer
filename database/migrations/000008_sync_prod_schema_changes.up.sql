BEGIN;

-- This migration brings staging/fresh schemas in sync with production
-- where manual schema changes were made. This migration is idempotent
-- and will be a no-op in production.

-- IMPORTANT: This migration requires table locks and may hang if the application is running.
-- Recommended approach: Stop the application (scale to 0 replicas) before running this migration.

-- Set a lock timeout to prevent hanging indefinitely
-- If the lock cannot be acquired within 10 seconds, the statement will fail
SET LOCAL lock_timeout = '10s';

-- 1. Change peer_id columns from integer to bigint across multiple tables
-- This accommodates larger peer ID values

-- Drop foreign key constraints that reference peers.id before changing type
ALTER TABLE neighbors DROP CONSTRAINT IF EXISTS fk_neighbors_peer_id;
ALTER TABLE peer_visits_index DROP CONSTRAINT IF EXISTS peer_visits_index_peer_id_fkey;
ALTER TABLE peers_states DROP CONSTRAINT IF EXISTS fk_peers_states_peer_id;
ALTER TABLE peers_x_multi_addresses DROP CONSTRAINT IF EXISTS fk_peers_x_multi_addresses_peer_id;
ALTER TABLE visits DROP CONSTRAINT IF EXISTS fk_visits_peer_id;

-- Change peers.id to bigint (this is the referenced column)
DO $$
BEGIN
    -- Check if column is not already bigint
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'peers'
        AND column_name = 'id'
        AND data_type = 'integer'
    ) THEN
        ALTER TABLE peers ALTER COLUMN id TYPE bigint;
    END IF;
END $$;

-- Change peer_id columns to bigint in referencing tables
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'neighbors'
        AND column_name = 'peer_id'
        AND data_type = 'integer'
    ) THEN
        ALTER TABLE neighbors ALTER COLUMN peer_id TYPE bigint;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'peer_visits_index'
        AND column_name = 'peer_id'
        AND data_type = 'integer'
    ) THEN
        ALTER TABLE peer_visits_index ALTER COLUMN peer_id TYPE bigint;
    END IF;
END $$;

-- Note: crawl_id stays as integer to match crawls.id

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'peer_visits_index'
        AND column_name = 'visit_id'
        AND data_type = 'integer'
    ) THEN
        ALTER TABLE peer_visits_index ALTER COLUMN visit_id TYPE bigint;
    END IF;
END $$;

-- Change visits.id to bigint to match peer_visits_index.visit_id
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'visits'
        AND column_name = 'id'
        AND data_type = 'integer'
    ) THEN
        ALTER TABLE visits ALTER COLUMN id TYPE bigint;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'peers_states'
        AND column_name = 'peer_id'
        AND data_type = 'integer'
    ) THEN
        ALTER TABLE peers_states ALTER COLUMN peer_id TYPE bigint;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'peers_x_multi_addresses'
        AND column_name = 'peer_id'
        AND data_type = 'integer'
    ) THEN
        ALTER TABLE peers_x_multi_addresses ALTER COLUMN peer_id TYPE bigint;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'visits'
        AND column_name = 'peer_id'
        AND data_type = 'integer'
    ) THEN
        ALTER TABLE visits ALTER COLUMN peer_id TYPE bigint;
    END IF;
END $$;

-- 2. Ensure peers table has a PRIMARY KEY on id
-- Note: The original migration replaced PRIMARY KEY with a UNIQUE index,
-- but sqlboiler requires a PRIMARY KEY for model generation.
DO $$
BEGIN
    -- If only UNIQUE index exists (no PRIMARY KEY), convert to PRIMARY KEY
    IF EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE indexname = 'new_id_unique'
        AND tablename = 'peers'
    ) AND NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'peers_pkey'
        AND conrelid = 'peers'::regclass
    ) THEN
        DROP INDEX new_id_unique;
        ALTER TABLE peers ADD PRIMARY KEY (id);
    END IF;

    -- If neither exists, add PRIMARY KEY
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'peers_pkey'
        AND conrelid = 'peers'::regclass
    ) AND NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE indexname = 'new_id_unique'
        AND tablename = 'peers'
    ) THEN
        ALTER TABLE peers ADD PRIMARY KEY (id);
    END IF;
END $$;

-- 3. Recreate foreign key constraints (after PRIMARY KEY is replaced)
-- These will now reference the new_id_unique index instead of peers_pkey
ALTER TABLE neighbors ADD CONSTRAINT fk_neighbors_peer_id
    FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE;

ALTER TABLE peer_visits_index ADD CONSTRAINT peer_visits_index_peer_id_fkey
    FOREIGN KEY (peer_id) REFERENCES peers (id);

ALTER TABLE peers_states ADD CONSTRAINT fk_peers_states_peer_id
    FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE;

ALTER TABLE peers_x_multi_addresses ADD CONSTRAINT fk_peers_x_multi_addresses_peer_id
    FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE;

ALTER TABLE visits ADD CONSTRAINT fk_visits_peer_id
    FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE;

-- 4. Add missing indexes on peer_visits_index
CREATE INDEX IF NOT EXISTS idx_peer_visits_index_crawl_id
    ON peer_visits_index (crawl_id);

CREATE INDEX IF NOT EXISTS idx_peer_visits_index_crawl_peer
    ON peer_visits_index (crawl_id, peer_id);

CREATE INDEX IF NOT EXISTS idx_peer_visits_index_latest_crawl
    ON peer_visits_index (crawl_id, last_seen DESC);

CREATE INDEX IF NOT EXISTS idx_peer_visits_index_peer_id
    ON peer_visits_index (peer_id);

CREATE INDEX IF NOT EXISTS idx_peer_visits_index_visit_id
    ON peer_visits_index (visit_id);

-- 5. Add missing index on peers
CREATE INDEX IF NOT EXISTS idx_peers_last_seen
    ON peers (last_seen DESC) WHERE (id IS NOT NULL);

-- 6. Add missing index on peers_states
CREATE INDEX IF NOT EXISTS idx_peers_states_peer_created_desc
    ON peers_states (peer_id, created_at DESC);

-- 7. Add missing index on visits
-- Note: idx_visits_crawl_peer is equivalent to idx_visits_crawl_id_peer_id
-- So we create it only if the other doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE indexname IN ('idx_visits_crawl_peer', 'idx_visits_crawl_id_peer_id')
        AND tablename = 'visits'
    ) THEN
        CREATE INDEX idx_visits_crawl_peer ON visits (crawl_id, peer_id);
    END IF;
END $$;

COMMIT;
