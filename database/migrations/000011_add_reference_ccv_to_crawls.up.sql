ALTER TABLE crawls ADD COLUMN reference_ccv TEXT;
COMMENT ON COLUMN crawls.reference_ccv IS 'The CCV used as reference for peer filtering during this crawl';
