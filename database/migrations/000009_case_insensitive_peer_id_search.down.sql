BEGIN;

-- Drop case-insensitive indexes
DROP INDEX IF EXISTS idx_pvi_peer_hash_ci;
DROP INDEX IF EXISTS idx_peers_multi_hash_ci;

COMMIT;
