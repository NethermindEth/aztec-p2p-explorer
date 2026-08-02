-- Add performance indexes for map endpoint optimization
-- This migration adds critical indexes to improve query performance for the /peers/map endpoint

-- Index for latest crawl lookup optimization
-- Used in queries that find the most recent finished crawl
CREATE INDEX IF NOT EXISTS idx_crawls_status_id 
ON crawls(status, id DESC) 
WHERE status = 'finished';

-- Index for visits by crawl optimization
-- Used to efficiently find visits for a specific crawl
CREATE INDEX IF NOT EXISTS idx_visits_crawl_id_peer_id 
ON visits(crawl_id, peer_id);

-- Index for latest peer states optimization
-- Used to efficiently find the most recent state for each peer
CREATE INDEX IF NOT EXISTS idx_peers_states_peer_id_created_at 
ON peers_states(peer_id, created_at DESC);

-- Index for geographic filtering optimization
-- Used for filtering peers by geographic location in map queries
CREATE INDEX IF NOT EXISTS idx_ip_addresses_geo 
ON ip_addresses(continent_id, country_id, city_id) 
WHERE latitude IS NOT NULL AND longitude IS NOT NULL;

-- Index for multi-address relationship optimization
-- Used to efficiently join multi_addresses with IP addresses
CREATE INDEX IF NOT EXISTS idx_multi_addresses_x_ip_addresses_ip_id 
ON multi_addresses_x_ip_addresses(ip_address_id);

-- Index for peer multi-address relationships
-- Used to efficiently find multi-addresses for peers
CREATE INDEX IF NOT EXISTS idx_peers_x_multi_addresses_peer_id 
ON peers_x_multi_addresses(peer_id, multi_address_id);

-- Index for multi-address reverse lookup
-- Used when querying from multi_addresses back to peers
CREATE INDEX IF NOT EXISTS idx_peers_x_multi_addresses_multi_addr_id 
ON peers_x_multi_addresses(multi_address_id);
