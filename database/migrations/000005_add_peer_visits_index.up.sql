BEGIN;

-- Create the denormalized index table for fast peer queries
CREATE TABLE peer_visits_index (
    -- Primary identifiers
    crawl_id INTEGER NOT NULL,
    visit_id INTEGER NOT NULL,
    peer_id INTEGER NOT NULL,
    peer_multi_hash TEXT NOT NULL,
    
    -- Pre-computed filter fields
    agent_version TEXT,
    client_name TEXT,  -- Extracted from agent_version (e.g., "aztec-node" from "aztec-node/1.2.3")
    
    -- Geographic fields (denormalized from first IP)
    continent_name TEXT,
    continent_code CHAR(2),
    country_name TEXT,
    country_iso CHAR(3),
    city_name TEXT,
    
    -- Autonomous System fields
    as_name TEXT,
    as_number INTEGER,
    
    -- Sync status fields
    is_synced BOOLEAN,
    block_height BIGINT,
    spec_version TEXT,
    
    -- Sort fields
    created_at TIMESTAMPTZ NOT NULL,
    last_seen TIMESTAMPTZ NOT NULL,
    
    -- For efficient lookups and deduplication
    multiaddr_hash TEXT,  -- Hash of all multiaddrs for deduplication
    
    PRIMARY KEY (crawl_id, visit_id),
    FOREIGN KEY (crawl_id) REFERENCES crawls(id) ON DELETE CASCADE,
    FOREIGN KEY (peer_id) REFERENCES peers(id) ON DELETE CASCADE
);

COMMENT ON TABLE peer_visits_index IS 'Denormalized index table for fast peer queries. Populated after each crawl completes.';
COMMENT ON COLUMN peer_visits_index.client_name IS 'Extracted client name from agent_version (e.g., "aztec-node" from "aztec-node/1.2.3")';
COMMENT ON COLUMN peer_visits_index.multiaddr_hash IS 'Hash of all multiaddrs for the peer, used for deduplication';

-- Create indexes for filtering
CREATE INDEX idx_pvi_crawl_peer ON peer_visits_index(crawl_id, peer_id);
CREATE INDEX idx_pvi_crawl_client ON peer_visits_index(crawl_id, client_name) WHERE client_name IS NOT NULL;
CREATE INDEX idx_pvi_crawl_continent ON peer_visits_index(crawl_id, continent_name) WHERE continent_name IS NOT NULL;
CREATE INDEX idx_pvi_crawl_continent_code ON peer_visits_index(crawl_id, continent_code) WHERE continent_code IS NOT NULL;
CREATE INDEX idx_pvi_crawl_country ON peer_visits_index(crawl_id, country_name) WHERE country_name IS NOT NULL;
CREATE INDEX idx_pvi_crawl_country_iso ON peer_visits_index(crawl_id, country_iso) WHERE country_iso IS NOT NULL;
CREATE INDEX idx_pvi_crawl_city ON peer_visits_index(crawl_id, city_name) WHERE city_name IS NOT NULL;
CREATE INDEX idx_pvi_crawl_as_name ON peer_visits_index(crawl_id, as_name) WHERE as_name IS NOT NULL;
CREATE INDEX idx_pvi_crawl_as_number ON peer_visits_index(crawl_id, as_number) WHERE as_number IS NOT NULL;
CREATE INDEX idx_pvi_crawl_synced ON peer_visits_index(crawl_id, is_synced) WHERE is_synced IS NOT NULL;

-- Create indexes for sorting
CREATE INDEX idx_pvi_crawl_last_seen ON peer_visits_index(crawl_id, last_seen DESC);
CREATE INDEX idx_pvi_crawl_block_height ON peer_visits_index(crawl_id, block_height DESC NULLS LAST);
CREATE INDEX idx_pvi_crawl_created_at ON peer_visits_index(crawl_id, created_at DESC);

-- Composite indexes for common filter combinations
CREATE INDEX idx_pvi_crawl_country_synced ON peer_visits_index(crawl_id, country_name, is_synced) 
    WHERE country_name IS NOT NULL AND is_synced IS NOT NULL;
CREATE INDEX idx_pvi_crawl_client_synced ON peer_visits_index(crawl_id, client_name, is_synced)
    WHERE client_name IS NOT NULL AND is_synced IS NOT NULL;

-- Index for peer multi_hash partial matching
CREATE INDEX idx_pvi_peer_hash ON peer_visits_index(peer_multi_hash text_pattern_ops);

COMMIT;