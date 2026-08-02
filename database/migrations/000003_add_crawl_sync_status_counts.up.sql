BEGIN;

-- Table to store pre-calculated sync status counts per crawl
CREATE TABLE crawl_sync_status_counts
(
    -- A unique id that identifies this record
    id INT GENERATED ALWAYS AS IDENTITY,
    -- Reference to the crawl this count is for
    crawl_id INT NOT NULL,
    -- Number of peers that are synced
    synced_count INT NOT NULL DEFAULT 0,
    -- Number of peers that are not synced
    not_synced_count INT NOT NULL DEFAULT 0,
    -- Number of peers with unknown sync status
    unknown_count INT NOT NULL DEFAULT 0,
    -- Timestamp when this record was created
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id),
    -- Ensure one record per crawl
    CONSTRAINT uq_crawl_sync_status_crawl_id UNIQUE (crawl_id),
    -- Foreign key to crawls table
    CONSTRAINT fk_crawl_sync_status_crawl_id FOREIGN KEY (crawl_id) REFERENCES crawls(id) ON DELETE CASCADE
);

COMMENT ON TABLE crawl_sync_status_counts IS 'Stores pre-calculated sync status counts for each crawl';
COMMENT ON COLUMN crawl_sync_status_counts.id IS 'A unique id that identifies this record';
COMMENT ON COLUMN crawl_sync_status_counts.crawl_id IS 'Reference to the crawl this count is for';
COMMENT ON COLUMN crawl_sync_status_counts.synced_count IS 'Number of peers that are synced';
COMMENT ON COLUMN crawl_sync_status_counts.not_synced_count IS 'Number of peers that are not synced';
COMMENT ON COLUMN crawl_sync_status_counts.unknown_count IS 'Number of peers with unknown sync status';
COMMENT ON COLUMN crawl_sync_status_counts.created_at IS 'Timestamp when this record was created';

-- Index for fast lookup by crawl_id
CREATE INDEX idx_crawl_sync_status_counts_crawl_id ON crawl_sync_status_counts(crawl_id);

-- Index for getting the latest counts
CREATE INDEX idx_crawl_sync_status_counts_created_at ON crawl_sync_status_counts(created_at DESC);

COMMIT;