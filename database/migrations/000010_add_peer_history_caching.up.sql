BEGIN;

-- Add CCV-filtered peer count column to crawls table
ALTER TABLE crawls
ADD COLUMN ccv_filtered_peer_count INT;

COMMENT ON COLUMN crawls.ccv_filtered_peer_count IS 'Number of distinct CCV-filtered peers in this crawl';

-- Create table for pre-aggregated daily maximum peer counts
CREATE TABLE daily_peer_history (
    -- A unique id that identifies this record
    id INT GENERATED ALWAYS AS IDENTITY,
    -- The date for this peer count record
    date DATE NOT NULL,
    -- Maximum peer count seen on this date (across all crawls on this date)
    peer_count INT NOT NULL,
    -- Timestamp when this record was created
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Timestamp when this record was last updated
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id),
    -- Ensure one record per date
    CONSTRAINT uq_daily_peer_history_date UNIQUE (date)
);

-- Index for fast date range queries (descending for recent-first queries)
CREATE INDEX idx_daily_peer_history_date ON daily_peer_history(date DESC);

COMMENT ON TABLE daily_peer_history IS 'Stores maximum CCV-filtered peer count per day for fast history queries';
COMMENT ON COLUMN daily_peer_history.id IS 'A unique id that identifies this record';
COMMENT ON COLUMN daily_peer_history.date IS 'The date for this peer count record';
COMMENT ON COLUMN daily_peer_history.peer_count IS 'Maximum CCV-filtered peer count seen on this date';
COMMENT ON COLUMN daily_peer_history.created_at IS 'Timestamp when this record was created';
COMMENT ON COLUMN daily_peer_history.updated_at IS 'Timestamp when this record was last updated';

COMMIT;
