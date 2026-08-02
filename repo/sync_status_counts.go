package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

// syncSmoothingWindow is the number of most-recent peers_states rows per peer
// considered when computing the smoothed sync_status_total gauge. Excluding
// `unknown` (NULL) from the denominator (see GetSyncStatusCountByCrawlID),
// transient glitches that flip a peer to NULL for 1-2 crawls are absorbed.
// Real `synced -> not_synced` transitions still affect the metric proportionally.
//
// Tuning: bump to absorb longer glitches; reduce for faster reaction to real
// degradations. Width is per-peer rows, not crawls — peers visited
// inconsistently smooth over their actual sample history.
const syncSmoothingWindow = 10

// CreateCrawlSyncStatusCount creates a new sync status count record for a crawl
func (r *PeerRepository) CreateCrawlSyncStatusCount(
	ctx context.Context,
	crawlID, syncedCount, notSyncedCount, unknownCount, totalCount int,
) error {
	count := &models.CrawlSyncStatusCount{
		CrawlID:        crawlID,
		SyncedCount:    syncedCount,
		NotSyncedCount: notSyncedCount,
		UnknownCount:   unknownCount,
		TotalCount:     totalCount,
	}

	err := count.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("failed to insert crawl sync status count: %w", err)
	}

	return nil
}

// GetCrawlSyncStatusCount retrieves sync status counts for a specific crawl
func (r *PeerRepository) GetCrawlSyncStatusCount(ctx context.Context, crawlID int) (*models.CrawlSyncStatusCount, error) {
	count, err := models.CrawlSyncStatusCounts(
		qm.Where("crawl_id = ?", crawlID),
	).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get crawl sync status count: %w", err)
	}

	return count, nil
}

// GetSyncStatusCountByCrawlID returns smoothed per-bucket peer counts for the
// given crawl, suitable for the sync_status_total Prometheus gauge.
//
// For each peer visited in the crawl, the function looks at the last
// syncSmoothingWindow rows of peers_states. Each peer contributes a weighted
// fraction to the synced / not_synced / unknown buckets, using
// (synced_count / known_count) where known_count excludes NULL rows. A peer
// with zero known rows in the window contributes 1.0 to `unknown` (data
// deficit). Per-peer contributions sum to 1.0; the three returned buckets sum
// to the in-crawl peer count (also returned as `"total"`).
//
// Bypasses the precomputed crawl_sync_status_counts table — that table still
// holds the raw int counts written by CalculateAndStoreSyncStatusCount, used
// by other repo readers and the public HTTP API.
func (r *PeerRepository) GetSyncStatusCountByCrawlID(ctx context.Context, crawlID int) (map[string]float64, error) {
	var rows []struct {
		Name  string  `boil:"name"`
		Count float64 `boil:"count"`
	}

	query := fmt.Sprintf(`
		WITH peers_in_crawl AS (
			SELECT DISTINCT peer_id FROM visits WHERE crawl_id = %d
		),
		recent_states AS (
			SELECT
				ps.peer_id,
				ps.is_synced,
				ROW_NUMBER() OVER (PARTITION BY ps.peer_id ORDER BY ps.created_at DESC) AS rn
			FROM peers_states ps
			INNER JOIN peers_in_crawl pic ON ps.peer_id = pic.peer_id
		),
		windowed AS (
			SELECT peer_id, is_synced FROM recent_states WHERE rn <= %d
		),
		per_peer AS (
			SELECT
				pic.peer_id,
				SUM(CASE WHEN w.is_synced = true      THEN 1 ELSE 0 END) AS n_synced,
				SUM(CASE WHEN w.is_synced = false     THEN 1 ELSE 0 END) AS n_not_synced,
				SUM(CASE WHEN w.is_synced IS NOT NULL THEN 1 ELSE 0 END) AS n_known
			FROM peers_in_crawl pic
			LEFT JOIN windowed w ON w.peer_id = pic.peer_id
			GROUP BY pic.peer_id
		),
		peer_contribs AS (
			SELECT
				CASE WHEN n_known > 0 THEN n_synced::float8     / n_known ELSE 0 END AS w_synced,
				CASE WHEN n_known > 0 THEN n_not_synced::float8 / n_known ELSE 0 END AS w_not_synced,
				CASE WHEN n_known = 0 THEN 1.0                            ELSE 0 END AS w_unknown
			FROM per_peer
		)
		SELECT 'synced'     AS name, COALESCE(SUM(w_synced),     0) AS count FROM peer_contribs UNION ALL
		SELECT 'not_synced' AS name, COALESCE(SUM(w_not_synced), 0) AS count FROM peer_contribs UNION ALL
		SELECT 'unknown'    AS name, COALESCE(SUM(w_unknown),    0) AS count FROM peer_contribs
	`, crawlID, syncSmoothingWindow)

	if err := models.NewQuery(qm.SQL(query)).Bind(ctx, r.db, &rows); err != nil {
		return nil, fmt.Errorf("failed to compute smoothed sync status counts: %w", err)
	}

	if len(rows) == 0 {
		return nil, nil
	}

	result := make(map[string]float64, len(rows)+1)
	for _, row := range rows {
		result[row.Name] = row.Count
	}
	result["total"] = result["synced"] + result["not_synced"] + result["unknown"]
	return result, nil
}

// GetTotalCountByCrawlID retrieves the total count for a specific crawl ID
func (r *PeerRepository) GetTotalCountByCrawlID(ctx context.Context, crawlID int) (int64, error) {
	count, err := r.GetCrawlSyncStatusCount(ctx, crawlID)
	if err != nil {
		return 0, err
	}
	if count == nil {
		return 0, nil
	}

	return int64(count.TotalCount), nil
}

// GetLatestFinishedCrawlID retrieves the ID of the latest finished crawl
func (r *PeerRepository) GetLatestFinishedCrawlID(ctx context.Context) (int, error) {
	latestCrawl, err := models.Crawls(
		qm.Where("status = ?", "finished"),
		qm.OrderBy("id DESC"),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get latest crawl: %w", err)
	}

	return latestCrawl.ID, nil
}

// GetLatestCrawlSyncStatusCount retrieves sync status counts from the latest finished crawl
func (r *PeerRepository) GetLatestCrawlSyncStatusCount(ctx context.Context) (map[string]int, error) {
	// First, get the latest finished crawl
	latestCrawl, err := models.Crawls(
		qm.Where("status = ?", "finished"),
		qm.OrderBy("id DESC"),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest crawl: %w", err)
	}

	// Get the sync status count for this crawl
	count, err := r.GetCrawlSyncStatusCount(ctx, latestCrawl.ID)
	if err != nil {
		return nil, err
	}
	if count == nil {
		return nil, nil
	}

	// Convert to the expected map format
	result := make(map[string]int)
	result["synced"] = count.SyncedCount
	result["not_synced"] = count.NotSyncedCount
	result["unknown"] = count.UnknownCount
	result["total"] = count.TotalCount

	return result, nil
}

// GetLatestCrawlTotalCount retrieves the total count from the latest finished crawl
func (r *PeerRepository) GetLatestCrawlTotalCount(ctx context.Context) (int64, error) {
	// First, get the latest finished crawl
	latestCrawl, err := models.Crawls(
		qm.Where("status = ?", "finished"),
		qm.OrderBy("id DESC"),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get latest crawl: %w", err)
	}

	// Get the sync status count for this crawl
	count, err := r.GetCrawlSyncStatusCount(ctx, latestCrawl.ID)
	if err != nil {
		return 0, err
	}
	if count == nil {
		return 0, nil
	}

	return int64(count.TotalCount), nil
}

// CalculateAndStoreSyncStatusCount calculates sync status counts for a crawl and stores them
func (r *PeerRepository) CalculateAndStoreSyncStatusCount(ctx context.Context, crawlID int) error {
	// Use the optimised query to calculate counts
	var results []struct {
		Name  string `boil:"name"`
		Count int    `boil:"count"`
	}

	// Query to calculate sync status counts for the specific crawl
	err := models.NewQuery(
		qm.SQL(fmt.Sprintf(`
			WITH peers_in_crawl AS (
				SELECT DISTINCT peer_id
				FROM visits
				WHERE crawl_id = %d
			),
			latest_states AS (
				SELECT DISTINCT ON (ps.peer_id)
					ps.peer_id,
					ps.is_synced
				FROM peers_states ps
				INNER JOIN peers_in_crawl pic ON ps.peer_id = pic.peer_id
				ORDER BY ps.peer_id, ps.created_at DESC
			)
			SELECT 
				CASE 
					WHEN ls.is_synced IS NULL THEN 'unknown'
					WHEN ls.is_synced = true THEN 'synced'
					ELSE 'not_synced'
				END AS name,
				COUNT(*) AS count
			FROM peers_in_crawl pic
			LEFT JOIN latest_states ls ON pic.peer_id = ls.peer_id
			GROUP BY name
		`, crawlID)),
	).Bind(ctx, r.db, &results)
	if err != nil {
		return fmt.Errorf("failed to calculate sync status counts: %w", err)
	}

	// Initialise counts
	syncedCount := 0
	notSyncedCount := 0
	unknownCount := 0

	// Parse results
	for _, result := range results {
		switch result.Name {
		case "synced":
			syncedCount = result.Count
		case "not_synced":
			notSyncedCount = result.Count
		case "unknown":
			unknownCount = result.Count
		}
	}

	// Calculate total count (sum of all categories)
	totalCount := syncedCount + notSyncedCount + unknownCount

	// Store the counts
	return r.CreateCrawlSyncStatusCount(ctx, crawlID, syncedCount, notSyncedCount, unknownCount, totalCount)
}
