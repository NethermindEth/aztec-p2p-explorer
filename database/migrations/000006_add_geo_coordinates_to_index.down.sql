BEGIN;

-- Remove geo coordinate columns
ALTER TABLE peer_visits_index 
DROP COLUMN IF EXISTS latitude,
DROP COLUMN IF EXISTS longitude;

-- Index will be automatically dropped when columns are dropped

COMMIT;