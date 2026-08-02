BEGIN;

-- Add case-insensitive index on peers.multi_hash for efficient case-insensitive peer ID lookups
CREATE INDEX idx_peers_multi_hash_ci ON peers (LOWER(multi_hash));

-- Add case-insensitive index on peer_visits_index.peer_multi_hash for efficient case-insensitive searches
CREATE INDEX idx_pvi_peer_hash_ci ON peer_visits_index (LOWER(peer_multi_hash));

COMMENT ON INDEX idx_peers_multi_hash_ci IS 'Case-insensitive index for peer ID searches';
COMMENT ON INDEX idx_pvi_peer_hash_ci IS 'Case-insensitive index for peer ID searches in the visits index';

COMMIT;
