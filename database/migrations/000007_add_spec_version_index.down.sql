BEGIN;

-- Drop the spec_version index
DROP INDEX IF EXISTS idx_pvi_crawl_spec_version;

COMMIT;