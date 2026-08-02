BEGIN;

-- Add latitude and longitude columns to peer_visits_index for map visualization
ALTER TABLE peer_visits_index 
ADD COLUMN latitude DOUBLE PRECISION,
ADD COLUMN longitude DOUBLE PRECISION;

COMMENT ON COLUMN peer_visits_index.latitude IS 'Latitude coordinate from first IP''s geo info for map visualization';
COMMENT ON COLUMN peer_visits_index.longitude IS 'Longitude coordinate from first IP''s geo info for map visualization';

-- Create index for geo queries (useful for bounding box queries in the future)
CREATE INDEX idx_pvi_crawl_geo ON peer_visits_index(crawl_id, latitude, longitude) 
    WHERE latitude IS NOT NULL AND longitude IS NOT NULL;

COMMIT;