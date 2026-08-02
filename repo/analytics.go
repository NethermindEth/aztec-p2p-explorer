//nolint:dupl
package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

// GetTotalPeerCount returns the total number of peers in the database.
func (r *PeerRepository) GetTotalPeerCount(ctx context.Context) (int64, error) {
	return models.Peers().Count(ctx, r.db)
}

// GetLatestPeerCount returns the number of peers in the database in the last crawl.
func (r *PeerRepository) GetLatestPeerCount(ctx context.Context) (int64, error) {
	latestCrawl, err := models.Crawls(
		qm.OrderBy("id DESC"),
		qm.Where("status = 'finished'"),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		return 0, fmt.Errorf("get latest crawl: %w", err)
	}

	count, err := models.Visits(
		qm.Select("COUNT(DISTINCT peer_id)"),
		qm.Where("crawl_id = ?", latestCrawl.ID),
	).Count(ctx, r.db)
	if err != nil {
		return 0, fmt.Errorf("count distinct peers in latest crawl: %w", err)
	}

	return count, nil
}

// GetChurnInPeerCount returns the number of new peers that
// were seen in the latest crawl but were not seen in the previous crawls.
func (r *PeerRepository) GetChurnInPeerCount(ctx context.Context) (int64, error) {
	var result struct {
		Count int64 `boil:"count"`
	}
	err := models.NewQuery(
		qm.With(buildCTE("peers_in_latest_crawl", peersInLatestCrawl())),
		qm.With(buildCTE("all_peers_before_last_crawl", allPeersBeforeLastCrawl())),
		qm.Select("COUNT(DISTINCT peer_id)"),
		qm.From("peers_in_latest_crawl"),
		qm.Where("peer_id NOT IN (SELECT peer_id FROM all_peers_before_last_crawl)"),
	).Bind(ctx, r.db, &result)
	if err != nil {
		return 0, fmt.Errorf("get no longer seen peers: %w", err)
	}

	return result.Count, nil
}

// GetChurnOutPeerCount returns the number of peers that were previously
// seen but are no longer seen in the latest crawl.
func (r *PeerRepository) GetChurnOutPeerCount(ctx context.Context) (int64, error) {
	var result struct {
		Count int64 `boil:"count"`
	}
	err := models.NewQuery(
		qm.With(buildCTE("peers_in_latest_crawl", peersInLatestCrawl())),
		qm.With(buildCTE("all_peers_before_last_crawl", allPeersBeforeLastCrawl())),
		qm.Select("COUNT(peer_id)"),
		qm.From("all_peers_before_last_crawl"),
		qm.Where("peer_id NOT IN (SELECT peer_id FROM peers_in_latest_crawl)"),
	).Bind(ctx, r.db, &result)
	if err != nil {
		return 0, fmt.Errorf("get no longer seen peers: %w", err)
	}

	return result.Count, nil
}

type AgentCount struct {
	Client string `json:"client"`
	Count  int    `json:"count"`
}

func (r *PeerRepository) GetPeerCountByAgent(ctx context.Context) ([]*AgentCount, error) {
	var results []*AgentCount
	err := models.NewQuery(
		qm.With(buildCTE("peers_in_latest_crawl", peersInLatestCrawl())),
		qm.Select(
			`av.agent_version as client,
			COUNT(DISTINCT p.id) as count`,
		),
		qm.From("peers_in_latest_crawl"),
		qm.InnerJoin("peers p on peers_in_latest_crawl.peer_id = p.id"),
		qm.InnerJoin("agent_versions av on p.agent_version_id = av.id"),
		qm.GroupBy("av.agent_version"),
		qm.OrderBy("count DESC"),
	).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, fmt.Errorf("get agent versions: %w", err)
	}

	return results, nil
}

type ProtocolCount struct {
	Protocol string `json:"protocol"`
	Count    int    `json:"count"`
}

func (r *PeerRepository) GetPeerCountByProtocol(ctx context.Context) ([]*ProtocolCount, error) {
	var results []*ProtocolCount
	err := models.NewQuery(
		qm.With(buildCTE("peers_in_latest_crawl", peersInLatestCrawl())),
		qm.Select(
			`p.protocol,
			COUNT(DISTINCT peers_in_latest_crawl.peer_id) as count`,
		),
		qm.From("peers_in_latest_crawl"),
		qm.InnerJoin("peers pe on peers_in_latest_crawl.peer_id = pe.id"),
		qm.InnerJoin("protocols_sets ps on pe.protocols_set_id = ps.id"),
		qm.InnerJoin("protocols p on p.id = ANY(ps.protocol_ids)"),
		qm.GroupBy("p.protocol"),
		qm.OrderBy("count DESC"),
	).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, fmt.Errorf("get protocols: %w", err)
	}

	return results, nil
}

type ContinentCount struct {
	ContinentName string `json:"continent_name"`
	ContinentCode string `json:"continent_code"`
	Count         int    `json:"count"`
}

func (r *PeerRepository) GetPeerCountByContinent(ctx context.Context) ([]*ContinentCount, error) {
	var results []*ContinentCount
	err := models.Peers(
		qm.With(buildCTE("peers_in_latest_crawl", peersInLatestCrawl())),
		qm.Select(
			`COALESCE(continents.continent_name, 'Unknown') as continent_name,
			COALESCE(continents.code, 'Unknown') as continent_code,
			COUNT(DISTINCT peers.id) as count`,
		),
		qm.LeftOuterJoin("peers_x_multi_addresses on peers.id = peers_x_multi_addresses.peer_id"),
		qm.LeftOuterJoin("multi_addresses on peers_x_multi_addresses.multi_address_id = multi_addresses.id"),
		qm.LeftOuterJoin("multi_addresses_x_ip_addresses on multi_addresses.id = multi_addresses_x_ip_addresses.multi_address_id"),
		qm.LeftOuterJoin("ip_addresses on multi_addresses_x_ip_addresses.ip_address_id = ip_addresses.id"),
		qm.LeftOuterJoin("continents on ip_addresses.continent_id = continents.id"),
		qm.InnerJoin("peers_in_latest_crawl on peers_in_latest_crawl.peer_id = peers.id"),
		qm.GroupBy("continent_name, continent_code"),
		qm.OrderBy("count DESC"),
	).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

type CountryCount struct {
	CountryName string `json:"country_name"`
	CountryCode string `json:"country_code"`
	Count       int    `json:"count"`
}

func (r *PeerRepository) GetPeerCountByCountry(ctx context.Context) ([]*CountryCount, error) {
	var results []*CountryCount
	err := models.Peers(
		qm.With(buildCTE("peers_in_latest_crawl", peersInLatestCrawl())),
		qm.Select(
			`COALESCE(countries.country_name, 'Unknown') as country_name, 
			TRIM(COALESCE(countries.iso_code, 'Unknown')) as country_code,
			COUNT(DISTINCT peers.id) as count`),
		qm.LeftOuterJoin("peers_x_multi_addresses on peers.id = peers_x_multi_addresses.peer_id"),
		qm.LeftOuterJoin("multi_addresses on peers_x_multi_addresses.multi_address_id = multi_addresses.id"),
		qm.LeftOuterJoin("multi_addresses_x_ip_addresses on multi_addresses.id = multi_addresses_x_ip_addresses.multi_address_id"),
		qm.LeftOuterJoin("ip_addresses on multi_addresses_x_ip_addresses.ip_address_id = ip_addresses.id"),
		qm.LeftOuterJoin("countries on ip_addresses.country_id = countries.id"),
		qm.InnerJoin("peers_in_latest_crawl on peers_in_latest_crawl.peer_id = peers.id"),
		qm.GroupBy("country_name, iso_code"),
		qm.OrderBy("count DESC"),
	).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

type CityCount struct {
	CityName string `json:"city_name"`
	Count    int    `json:"count"`
}

func (r *PeerRepository) GetPeerCountByCity(ctx context.Context) ([]*CityCount, error) {
	var results []*CityCount
	err := models.Peers(
		qm.With(buildCTE("peers_in_latest_crawl", peersInLatestCrawl())),
		qm.Select("COALESCE(cities.city_name, 'Unknown') as city_name, COUNT(DISTINCT peers.id) as count"),
		qm.LeftOuterJoin("peers_x_multi_addresses on peers.id = peers_x_multi_addresses.peer_id"),
		qm.LeftOuterJoin("multi_addresses on peers_x_multi_addresses.multi_address_id = multi_addresses.id"),
		qm.LeftOuterJoin("multi_addresses_x_ip_addresses on multi_addresses.id = multi_addresses_x_ip_addresses.multi_address_id"),
		qm.LeftOuterJoin("ip_addresses on multi_addresses_x_ip_addresses.ip_address_id = ip_addresses.id"),
		qm.LeftOuterJoin("cities on ip_addresses.city_id = cities.id"),
		qm.InnerJoin("peers_in_latest_crawl on peers_in_latest_crawl.peer_id = peers.id"),
		qm.GroupBy("city_name"),
		qm.OrderBy("count DESC"),
	).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

type ASCount struct {
	ASName   string `json:"as_name"`
	ASNumber int    `json:"as_number"`
	Count    int    `json:"count"`
}

func (r *PeerRepository) GetPeerCountByASO(ctx context.Context) ([]*ASCount, error) {
	var results []*ASCount
	err := models.Peers(
		qm.With(buildCTE("peers_in_latest_crawl", peersInLatestCrawl())),
		qm.Select(
			`COALESCE(autonomous_systems.as_name, 'Unknown') as as_name,
			COALESCE(autonomous_systems.as_number, 0) as as_number, 
			COUNT(DISTINCT peers.id) as count`,
		),
		qm.LeftOuterJoin("peers_x_multi_addresses on peers.id = peers_x_multi_addresses.peer_id"),
		qm.LeftOuterJoin("multi_addresses on peers_x_multi_addresses.multi_address_id = multi_addresses.id"),
		qm.LeftOuterJoin("multi_addresses_x_ip_addresses on multi_addresses.id = multi_addresses_x_ip_addresses.multi_address_id"),
		qm.LeftOuterJoin("ip_addresses on multi_addresses_x_ip_addresses.ip_address_id = ip_addresses.id"),
		qm.LeftOuterJoin("autonomous_systems on ip_addresses.as_id = autonomous_systems.id"),
		qm.InnerJoin("peers_in_latest_crawl on peers_in_latest_crawl.peer_id = peers.id"),
		qm.GroupBy("as_name, as_number"),
		qm.OrderBy("count DESC"),
	).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *PeerRepository) GetSyncStatusCount(ctx context.Context) (map[string]int, error) {
	var results []struct {
		Name  string `boil:"name"`
		Count int    `boil:"count"`
	}

	err := models.Peers(
		qm.With(buildCTE("peers_in_latest_crawl", peersInLatestCrawl())),
		qm.With(buildCTE("latest_peer_states", latestPeerStates())),
		qm.Select(`
            CASE 
                WHEN latest_peer_states.is_synced IS NULL THEN 'unknown'
                WHEN latest_peer_states.is_synced = true THEN 'synced'
                ELSE 'not_synced'
            END AS name,
            COUNT(DISTINCT peers.id) AS count
        `),
		qm.InnerJoin("peers_in_latest_crawl on peers_in_latest_crawl.peer_id = peers.id"),
		qm.LeftOuterJoin("latest_peer_states ON peers.id = latest_peer_states.peer_id"),
		qm.GroupBy("name"),
		qm.OrderBy("count DESC"),
	).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, fmt.Errorf("get sync status count: %w", err)
	}

	countMap := make(map[string]int)
	for _, result := range results {
		countMap[result.Name] = result.Count
	}

	return countMap, nil
}

type PeerHistoryPoint struct {
	Date      time.Time `json:"date" boil:"date"`
	PeerCount int       `json:"peer_count" boil:"peer_count"`
}

func (r *PeerRepository) GetPeerHistory(ctx context.Context, start, end time.Time) ([]*PeerHistoryPoint, error) {
	if end.IsZero() {
		end = time.Now()
	}

	if start.After(end) {
		return nil, fmt.Errorf("start date must be before end date")
	}

	var results []*PeerHistoryPoint
	err := models.NewQuery(
		qm.With(buildCTE("max_daily_peer_counts", maxDailyPeerCount(start, end))),
		qm.Select("date, peer_count"),
		qm.From("max_daily_peer_counts"),
	).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch peer history: %w", err)
	}

	return results, nil
}

func maxDailyPeerCount(start, end time.Time) string {
	format := "2006-01-02 15:04:05"
	return fmt.Sprintf(`
		SELECT finished_at::date as date, MAX(dialable_peers) peer_count FROM crawls 
		WHERE finished_at >= '%s' AND finished_at < '%s' AND status = 'finished'
		GROUP BY date ORDER BY date
	`, start.Format(format), end.Format(format))
}

func peersInLatestCrawl() string {
	return `
		SELECT DISTINCT peer_id
		FROM visits
		WHERE crawl_id = (SELECT id FROM crawls WHERE status = 'finished' ORDER BY id DESC LIMIT 1)
	`
}

func allPeersBeforeLastCrawl() string {
	return `
		SELECT peer_id
		FROM visits
		WHERE crawl_id < (SELECT id FROM crawls WHERE status = 'finished' ORDER BY id DESC LIMIT 1)
	`
}
