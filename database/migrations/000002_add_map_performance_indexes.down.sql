-- Rollback migration: Remove performance indexes for map endpoint
-- This migration removes the indexes added for map endpoint optimization

-- Remove geographic filtering index
DROP INDEX CONCURRENTLY IF EXISTS idx_ip_addresses_geo;

-- Remove multi-address relationship indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_multi_addresses_x_ip_addresses_ip_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_peers_x_multi_addresses_peer_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_peers_x_multi_addresses_multi_addr_id;

-- Remove peer states index
DROP INDEX CONCURRENTLY IF EXISTS idx_peers_states_peer_id_created_at;

-- Remove visits index
DROP INDEX CONCURRENTLY IF EXISTS idx_visits_crawl_id_peer_id;

-- Remove crawls index
DROP INDEX CONCURRENTLY IF EXISTS idx_crawls_status_id;