BEGIN;

-- Add index for spec_version filtering (used for CCV filtering)
CREATE INDEX idx_pvi_crawl_spec_version ON peer_visits_index(crawl_id, spec_version)
    WHERE spec_version IS NOT NULL;

COMMENT ON COLUMN peer_visits_index.spec_version IS 'Compressed Component Version (CCV) extracted from ENR aztec field, used for filtering specific protocol versions';

COMMIT;