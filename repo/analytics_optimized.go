package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

// GetPeerCountByCountryOptimized returns peer count grouped by country using the optimised index table
// with optional CCV filtering
func (r *PeerRepository) GetPeerCountByCountryOptimized(ctx context.Context, filterCCV string) ([]*CountryCount, error) {
	var results []*CountryCount

	baseQuery := `
		WITH latest_crawl AS (
			SELECT MAX(id) as crawl_id FROM crawls WHERE status = 'finished'
		)
		SELECT
			COALESCE(country_name, 'Unknown') as country_name,
			COALESCE(country_iso, 'Unknown') as country_code,
			COUNT(DISTINCT peer_multi_hash) as count
		FROM peer_visits_index, latest_crawl
		WHERE peer_visits_index.crawl_id = latest_crawl.crawl_id`

	// Add CCV filter if provided
	if filterCCV != "" {
		baseQuery += fmt.Sprintf(" AND spec_version = '%s'", filterCCV)
	}

	baseQuery += `
		GROUP BY country_name, country_iso
		ORDER BY count DESC`

	err := models.NewQuery(qm.SQL(baseQuery)).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, fmt.Errorf("get optimised peer count by country: %w", err)
	}

	return results, nil
}

// GetPeerCountByAgentOptimized returns peer count grouped by agent version using the optimised index table
// with optional CCV filtering
func (r *PeerRepository) GetPeerCountByAgentOptimized(ctx context.Context, filterCCV string) ([]*AgentCount, error) {
	var results []*AgentCount

	baseQuery := `
		WITH latest_crawl AS (
			SELECT MAX(id) as crawl_id FROM crawls WHERE status = 'finished'
		)
		SELECT
			agent_version as client,
			COUNT(DISTINCT peer_multi_hash) as count
		FROM peer_visits_index, latest_crawl
		WHERE peer_visits_index.crawl_id = latest_crawl.crawl_id`

	// Add CCV filter if provided
	if filterCCV != "" {
		baseQuery += fmt.Sprintf(" AND spec_version = '%s'", filterCCV)
	}

	baseQuery += `
		GROUP BY agent_version
		ORDER BY count DESC`

	err := models.NewQuery(qm.SQL(baseQuery)).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, fmt.Errorf("get optimised peer count by agent: %w", err)
	}

	return results, nil
}

// GetSyncStatusCountOptimized returns sync status counts using the optimised index table
// with optional CCV filtering
func (r *PeerRepository) GetSyncStatusCountOptimized(ctx context.Context, filterCCV string) (map[string]int, error) {
	var results []struct {
		Name  string `boil:"name"`
		Count int    `boil:"count"`
	}

	baseQuery := `
		WITH latest_crawl AS (
			SELECT MAX(id) as crawl_id FROM crawls WHERE status = 'finished'
		)
		SELECT
			CASE
				WHEN is_synced IS NULL THEN 'unknown'
				WHEN is_synced = true THEN 'synced'
				ELSE 'not_synced'
			END AS name,
			COUNT(DISTINCT peer_multi_hash) AS count
		FROM peer_visits_index, latest_crawl
		WHERE peer_visits_index.crawl_id = latest_crawl.crawl_id`

	// Add CCV filter if provided
	if filterCCV != "" {
		baseQuery += fmt.Sprintf(" AND spec_version = '%s'", filterCCV)
	}

	baseQuery += `
		GROUP BY name
		ORDER BY count DESC`

	err := models.NewQuery(qm.SQL(baseQuery)).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, fmt.Errorf("get optimised sync status count: %w", err)
	}

	countMap := make(map[string]int)
	for _, result := range results {
		countMap[result.Name] = result.Count
	}

	return countMap, nil
}

// GetLatestCrawlTotalCountOptimized returns the total peer count from the latest crawl
// using the optimised index table with optional CCV filtering
func (r *PeerRepository) GetLatestCrawlTotalCountOptimized(ctx context.Context, filterCCV string) (int64, error) {
	var result struct {
		Count int64 `boil:"count"`
	}

	baseQuery := `
		WITH latest_crawl AS (
			SELECT MAX(id) as crawl_id FROM crawls WHERE status = 'finished'
		)
		SELECT COUNT(DISTINCT peer_multi_hash) as count
		FROM peer_visits_index, latest_crawl
		WHERE peer_visits_index.crawl_id = latest_crawl.crawl_id`

	// Add CCV filter if provided
	if filterCCV != "" {
		baseQuery += fmt.Sprintf(" AND spec_version = '%s'", filterCCV)
	}

	err := models.NewQuery(qm.SQL(baseQuery)).Bind(ctx, r.db, &result)
	if err != nil {
		return 0, fmt.Errorf("get optimised latest crawl total count: %w", err)
	}

	return result.Count, nil
}

// GetPeerCountByContinentOptimized returns peer count grouped by continent using the optimised index table
// with optional CCV filtering
func (r *PeerRepository) GetPeerCountByContinentOptimized(ctx context.Context, filterCCV string) ([]*ContinentCount, error) {
	var results []*ContinentCount

	baseQuery := `
		WITH latest_crawl AS (
			SELECT MAX(id) as crawl_id FROM crawls WHERE status = 'finished'
		)
		SELECT
			COALESCE(continent_name, 'Unknown') as continent_name,
			COALESCE(continent_code, 'Unknown') as continent_code,
			COUNT(DISTINCT peer_multi_hash) as count
		FROM peer_visits_index, latest_crawl
		WHERE peer_visits_index.crawl_id = latest_crawl.crawl_id`

	// Add CCV filter if provided
	if filterCCV != "" {
		baseQuery += fmt.Sprintf(" AND spec_version = '%s'", filterCCV)
	}

	baseQuery += `
		GROUP BY continent_name, continent_code
		ORDER BY count DESC`

	err := models.NewQuery(qm.SQL(baseQuery)).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, fmt.Errorf("get optimised peer count by continent: %w", err)
	}

	return results, nil
}

// GetPeerHistoryOptimized returns peer count history using the optimised index table
// with optional CCV filtering. Returns the maximum peer count per day (matching the old behaviour).
// Since CCV filtering is done at insertion time, all data in peer_visits_index is already filtered.
func (r *PeerRepository) GetPeerHistoryOptimized(
	ctx context.Context, start, end time.Time, filterCCV string,
) ([]*PeerHistoryPoint, error) {
	if end.IsZero() {
		end = time.Now()
	}

	if start.After(end) {
		return nil, fmt.Errorf("start date must be before end date")
	}

	format := "2006-01-02 15:04:05"

	baseQuery := fmt.Sprintf(`
		WITH crawl_peer_counts AS (
			SELECT
				c.id as crawl_id,
				c.finished_at::date as date,
				COUNT(DISTINCT pvi.peer_multi_hash) as peer_count
			FROM peer_visits_index pvi
			INNER JOIN crawls c ON pvi.crawl_id = c.id
			WHERE c.status = 'finished'
				AND c.finished_at >= '%s'
				AND c.finished_at < '%s'`, start.Format(format), end.Format(format))

	// Add CCV filter if provided (though data is already filtered at insertion time)
	if filterCCV != "" {
		baseQuery += fmt.Sprintf(" AND pvi.spec_version = '%s'", filterCCV)
	}

	baseQuery += `
			GROUP BY c.id, date
		)
		SELECT 
			date,
			MAX(peer_count) as peer_count
		FROM crawl_peer_counts
		GROUP BY date
		ORDER BY date`

	var results []*PeerHistoryPoint
	err := models.NewQuery(qm.SQL(baseQuery)).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, fmt.Errorf("get optimised peer history: %w", err)
	}

	return results, nil
}
