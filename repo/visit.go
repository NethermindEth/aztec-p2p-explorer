package repo

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/lib/pq"

	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

const (
	defaultPageSize = 20
)

func (r *PeerRepository) CreatePeerVisit(ctx context.Context, crawlID int, visit *coretypes.Visit, peer *coretypes.Peer) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		if peer.AgentVersion == "" {
			peer.AgentVersion = "aztec"
		}
		avID, err := r.upsertAgentVersion(ctx, tx, peer.AgentVersion, 0)
		r.logger.Debug("Upserted agent version", "agentversion", peer.AgentVersion, "avID", avID, "error", err)
		if err != nil {
			return err
		}

		protocolMap := make(map[string]int)
		for _, p := range peer.Protocols {
			protocolMap[p] = 0
		}
		psID, err := r.upsertProtocolsWithSet(ctx, tx, protocolMap)
		if err != nil {
			return err
		}

		multiAddrs, maIDs, err := r.insertMultiAddrs(ctx, tx, peer.MultiAddrs)
		if err != nil {
			return err
		}

		peerID, err := r.findOrUpsertPeer(ctx, tx, peer.PeerID, null.NewInt(avID, avID != 0), null.NewInt(psID, psID != 0), multiAddrs)
		if err != nil {
			r.logger.Error("Failed to find or upsert peer", "peerID", peer.PeerID, "avID", avID, "error", err)
			return err
		}

		_, err = r.insertVisit(ctx, tx,
			peerID,
			null.NewInt(avID, avID != 0),
			null.NewInt(psID, psID != 0),
			visit,
			maIDs,
			null.IntFrom(crawlID),
		)

		return err
	})
}

// insertVisit inserts a visit record and returns the visit ID
func (r *PeerRepository) insertVisit(
	ctx context.Context,
	exec boil.ContextExecutor,
	peerID int64,
	agentVersionID null.Int,
	protocolsSetID null.Int,
	visit *coretypes.Visit,
	multiAddrIDs []int64,
	crawlID null.Int,
) (int64, error) {
	vs := &models.Visit{
		PeerID:          peerID,
		AgentVersionID:  agentVersionID,
		ProtocolsSetID:  protocolsSetID,
		ConnectError:    null.StringFrom(visit.ConnectErrorStr),
		CrawlError:      null.StringFrom(visit.CrawlErrorStr),
		VisitStartedAt:  visit.VisitStartedAt,
		VisitEndedAt:    visit.VisitEndedAt,
		ConnectDuration: null.StringFrom(visit.ConnectDuration),
		CrawlDuration:   null.StringFrom(visit.CrawlDuration),
		MultiAddressIds: types.Int64Array(multiAddrIDs),
		CrawlID:         crawlID,
	}

	err := vs.Insert(ctx, exec, boil.Infer())
	if err != nil {
		return 0, fmt.Errorf("insert visit: %w", err)
	}

	return vs.ID, nil
}

// findOrUpsertPeer upserts a peer record and returns the peer ID
func (r *PeerRepository) findOrUpsertPeer(
	ctx context.Context,
	exec boil.ContextExecutor,
	peerID string,
	agentVersionID null.Int,
	protocolsSetID null.Int,
	multiAddrs []*models.MultiAddress,
) (int64, error) {
	peer := &models.Peer{
		MultiHash:      peerID,
		AgentVersionID: agentVersionID,
		ProtocolsSetID: protocolsSetID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastSeen:       time.Now(),
	}

	// What the following does is that it only includes not null fields in the whitelist.
	// This way, we can avoid updating the record with null values, but at the same time
	// we can still retrieve the ID of the existing record.
	columns := []string{models.PeerColumns.MultiHash, models.PeerColumns.LastSeen}
	for _, valid := range []struct {
		valid  bool
		column string
	}{
		{agentVersionID.Valid, models.PeerColumns.AgentVersionID},
		{protocolsSetID.Valid, models.PeerColumns.ProtocolsSetID},
	} {
		if valid.valid {
			columns = append(columns, valid.column)
		}
	}

	if len(columns) > 2 {
		columns = append(columns, models.PeerColumns.UpdatedAt)
	}

	err := peer.Upsert(ctx, exec, true,
		[]string{models.PeerColumns.MultiHash},
		boil.Whitelist(columns...), boil.Infer())
	if err != nil {
		return 0, fmt.Errorf("upsert peer: %w", err)
	}

	if multiAddrs != nil { // avoid deleting multi address if empty
		err = peer.SetMultiAddresses(ctx, exec, false, multiAddrs...)
		if err != nil {
			return 0, fmt.Errorf("add multi addresses: %w", err)
		}
	}

	return peer.ID, nil
}

// insertMultiAddrs inserts multi address records if not exists and returns the multi address objects and IDs
func (r *PeerRepository) insertMultiAddrs(
	ctx context.Context,
	exec boil.ContextExecutor,
	maddrs []*coretypes.MultiAddr,
) ([]*models.MultiAddress, []int64, error) {
	var multiAddrs []*models.MultiAddress
	var maIDs []int64

	for _, maddr := range maddrs {
		ma, err := r.insertMultiAddr(ctx, exec, maddr)
		r.logger.Debug("Inserted multi address", "ma", ma, "err", err)
		if err != nil {
			return nil, nil, err
		}
		multiAddrs = append(multiAddrs, ma)
		maIDs = append(maIDs, int64(ma.ID))
	}

	return multiAddrs, maIDs, nil
}

// insertMultiAddr inserts a multi address record if not exists and returns the multi address object
func (r *PeerRepository) insertMultiAddr(
	ctx context.Context,
	exec boil.ContextExecutor,
	maddr *coretypes.MultiAddr,
) (*models.MultiAddress, error) {
	ma := &models.MultiAddress{
		Maddr:     maddr.Address,
		CreatedAt: time.Now(),
	}

	err := ma.Upsert(ctx, exec, true, []string{models.MultiAddressColumns.Maddr},
		boil.Whitelist(models.MultiAddressColumns.Maddr), boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("insert multi address: %w", err)
	}

	if maddr.IPList != nil {
		for _, ip := range maddr.IPList {
			ip, err := r.upsertIPAddress(ctx, exec, ip.IPAddress, ip.Port,
				null.Int{}, null.Int{}, null.Int{}, null.Int{}, null.Float64{}, null.Float64{})
			if err != nil {
				return nil, err
			}

			err = ma.SetIPAddresses(ctx, exec, false, ip)
			if err != nil {
				return nil, fmt.Errorf("add ip addresses: %w", err)
			}
		}
	}

	return ma, nil
}

type PeerInfo struct {
	PeerID         string                 `json:"id" boil:"multi_hash"`
	CreatedAt      time.Time              `json:"created_at" boil:"created_at"`
	LastSeen       time.Time              `json:"last_seen" boil:"last_seen"`
	AgentVersion   null.String            `json:"client" boil:"agent_version" swaggertype:"string,null"`
	MultiAddresses []*coretypes.MultiAddr `json:"multi_addresses" boil:"multi_addresses"`
	Protocols      pq.StringArray         `json:"protocols" boil:"protocols" swaggertype:"array,string"`
	BlockHeight    null.Int64             `json:"block_height" boil:"block_height" swaggertype:"integer,null"`
	SpecVersion    null.String            `json:"spec_version" boil:"spec_version" swaggertype:"string,null"`
	IsSynced       null.Bool              `json:"is_synced" boil:"is_synced" swaggertype:"boolean,null"`
}

type PeerQueryOptions struct {
	Filter          *coretypes.PeerFilter
	Sort            string
	IsAsc           bool
	Latest          bool
	PageSize        int
	PaginationToken string
}

// GetPeers returns a list of peer info with the given filter and pagination parameters
func (r *PeerRepository) GetPeers(
	ctx context.Context,
	opt *PeerQueryOptions,
) ([]*PeerInfo, string, int64, error) {
	sortColumn, err := getSortColumn(opt.Sort)
	if err != nil {
		return nil, "", 0, err
	}

	pageSize := opt.PageSize
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	// Get total count first
	totalCount, err := r.getPeersTotalCount(ctx, opt.Filter, opt.Latest)
	if err != nil {
		return nil, "", 0, fmt.Errorf("error getting total peer count: %w", err)
	}

	query, err := r.getPeersQuery(opt.Filter, sortColumn, opt.IsAsc, pageSize, opt.PaginationToken, opt.Latest)
	if err != nil {
		return nil, "", 0, err
	}

	rows, err := query.QueryContext(ctx, r.db)
	if err != nil {
		return nil, "", 0, fmt.Errorf("error executing peer info query: %w", err)
	}
	defer rows.Close()

	result, err := scanPeerRows(rows)
	if err != nil {
		return nil, "", 0, err
	}

	nextPageToken, err := createNextPageToken(result, pageSize, sortColumn, opt.IsAsc, opt.Filter)
	if err != nil {
		return nil, "", 0, err
	}

	if len(result) > pageSize {
		result = result[:pageSize] // remove the extra record used for pagination
	}

	return result, nextPageToken, totalCount, nil
}

// getPeersTotalCount returns the total count of peers matching the given filter
func (r *PeerRepository) getPeersTotalCount(
	ctx context.Context,
	filter *coretypes.PeerFilter,
	latest bool,
) (int64, error) {
	mods := []qm.QueryMod{
		qm.With(buildCTE("filtered_peers", filteredPeers(filter, latest))),
		qm.Select("COUNT(*)"),
		qm.From("filtered_peers"),
	}

	query := models.NewQuery(mods...)
	var count int64
	err := query.QueryRowContext(ctx, r.db).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error executing peer count query: %w", err)
	}

	return count, nil
}

func (r *PeerRepository) GetPeerByID(
	ctx context.Context,
	peerID string,
) (*PeerInfo, error) {
	sortColumn, _ := getSortColumn("")

	query, err := r.getPeersQuery(&coretypes.PeerFilter{PeerID: peerID}, sortColumn, false, 1, "", false)
	if err != nil {
		return nil, err
	}

	rows, err := query.QueryContext(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("error executing peer info query: %w", err)
	}
	defer rows.Close()

	peers, err := scanPeerRows(rows)
	if err != nil {
		return nil, err
	}

	if len(peers) == 0 {
		return nil, nil
	}

	return peers[0], nil
}

// getPeersQuery returns the query to fetch peer info with the given filter and pagination parameters
func (r *PeerRepository) getPeersQuery(
	filter *coretypes.PeerFilter,
	sortColumn string,
	isAsc bool,
	pageSize int,
	paginationToken string,
	latest bool,
) (*queries.Query, error) {
	mods := []qm.QueryMod{
		qm.With(buildCTE("latest_peer_states", latestPeerStates())),
		qm.With(buildCTE("peer_protocols", peerProtocols())),
		qm.With(buildCTE("filtered_peers", filteredPeers(filter, latest))),
		qm.With(buildCTE("peer_addresses", peerAddresses())),
		qm.Select(
			"peers.multi_hash",
			"peers.created_at",
			"peers.last_seen",
			"agent_versions.agent_version",
			"peer_addresses.multi_addresses",
			"peer_protocols.protocols",
			"latest_peer_states.block_height",
			"latest_peer_states.spec_version",
			"latest_peer_states.is_synced",
		),
		qm.From("peers"),
		qm.InnerJoin("peer_addresses ON peers.id = peer_addresses.peer_id"),
		qm.LeftOuterJoin("agent_versions ON peers.agent_version_id = agent_versions.id"),
		qm.LeftOuterJoin("peer_protocols ON peers.protocols_set_id = peer_protocols.protocols_set_id"),
		qm.LeftOuterJoin("latest_peer_states ON peers.id = latest_peer_states.peer_id"),
		qm.Where("peers.id IN (SELECT id FROM filtered_peers)"),
	}

	conditions := buildPeerFilterConditions(filter)
	mods = append(mods, conditions...)

	paginationCond, err := buildPaginationCondition(paginationToken, sortColumn, isAsc)
	if err != nil {
		return nil, err
	}
	if paginationCond != nil {
		mods = append(mods, paginationCond)
	}

	mods = append(mods,
		qm.OrderBy(fmt.Sprintf("%s %s", sortColumn, getOrder(isAsc))),
		qm.Limit(pageSize+1), // fetch one more to determine if there is a next page
	)

	return models.NewQuery(mods...), nil
}

func scanPeerRows(rows *sql.Rows) ([]*PeerInfo, error) {
	var result []*PeerInfo
	for rows.Next() {
		var pi PeerInfo
		var multiAddressesJSON []byte
		err := rows.Scan(
			&pi.PeerID,
			&pi.CreatedAt,
			&pi.LastSeen,
			&pi.AgentVersion,
			&multiAddressesJSON,
			&pi.Protocols,
			&pi.BlockHeight,
			&pi.SpecVersion,
			&pi.IsSynced,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		err = json.Unmarshal(multiAddressesJSON, &pi.MultiAddresses)
		if err != nil {
			return nil, fmt.Errorf("error parsing multi addresses: %w", err)
		}

		result = append(result, &pi)
	}
	return result, nil
}

// buildPeerFilterConditions builds the where conditions for the given peer filter
func buildPeerFilterConditions(filter *coretypes.PeerFilter) []qm.QueryMod {
	conditions := []qm.QueryMod{}

	if filter == nil {
		return conditions
	}

	if len(filter.ClientNames) > 0 {
		clientConditions := []string{}
		for _, client := range filter.ClientNames {
			clientConditions = append(clientConditions, fmt.Sprintf("agent_versions.agent_version ILIKE '%%%s%%'", client))
		}
		conditions = append(conditions, qm.Where(strings.Join(clientConditions, " OR ")))
	}

	if filter.SyncStatus != nil {
		conditions = append(conditions, qm.Where(
			"latest_peer_states.is_synced = ?", *filter.SyncStatus,
		))
	}

	if filter.PeerID != "" {
		conditions = append(conditions, qm.Where(
			"LOWER(peers.multi_hash) = LOWER(?)", filter.PeerID,
		))
	}

	return conditions
}

// buildPaginationCondition builds the pagination condition for the given pagination token
func buildPaginationCondition(paginationToken, sortColumn string, isAsc bool) (qm.QueryMod, error) {
	if paginationToken == "" {
		return nil, nil
	}

	token, err := decodePaginationToken(paginationToken)
	if err != nil {
		return nil, fmt.Errorf("invalid pagination token: %w", err)
	}

	if token.SortColumn != sortColumn || token.IsAsc != isAsc || !coretypes.CompareFilters(token.Filter, token.Filter) {
		return nil, fmt.Errorf("pagination token is not valid for the current query parameters")
	}

	if isAsc {
		return qm.Where(fmt.Sprintf("(%s > ? OR (%s = ? AND peers.multi_hash > ?))", sortColumn, sortColumn),
			token.LastSortValue, token.LastSortValue, token.LastPeerID), nil
	} else {
		return qm.Where(fmt.Sprintf("(%s < ? OR (%s = ? AND peers.multi_hash < ?))", sortColumn, sortColumn),
			token.LastSortValue, token.LastSortValue, token.LastPeerID), nil
	}
}

// createNextPageToken creates a base64 encoded pagination token for the next page
func createNextPageToken(result []*PeerInfo, pageSize int, sortColumn string, isAsc bool, filter *coretypes.PeerFilter) (string, error) {
	if len(result) <= pageSize {
		return "", nil
	}

	lastPeer := result[pageSize-1]
	lastSortValue := getLastSortValue(lastPeer, sortColumn)

	token := PaginationToken{
		LastPeerID:    lastPeer.PeerID,
		LastSortValue: lastSortValue,
		SortColumn:    sortColumn,
		IsAsc:         isAsc,
		Filter:        filter,
	}

	return token.EncodeToString()
}

func getLastSortValue(lastPeer *PeerInfo, sortColumn string) interface{} {
	switch sortColumn {
	case "peers.last_seen":
		return lastPeer.LastSeen
	case "latest_peer_states.block_height":
		return lastPeer.BlockHeight
	default:
		return lastPeer.CreatedAt
	}
}

// buildCTE builds a common table expression with the given name and query.
func buildCTE(name, query string) string {
	return fmt.Sprintf("%s AS (%s)", name, query)
}

// latestPeerStates is a raw SQL query that returns the latest peer state for each peer
func latestPeerStates() string {
	return `SELECT DISTINCT ON (peer_id)
				peer_id,
				block_height,
				spec_version,
				is_synced,
				created_at
				FROM peers_states
				ORDER BY peer_id, created_at DESC
			`
}

// peerProtocols is a raw SQL query that returns the string array form of the protocols for each protocols set
func peerProtocols() string {
	return `SELECT
				protocols_sets.id AS protocols_set_id,
				array_agg(protocols.protocol) AS protocols
			FROM protocols_sets
			JOIN LATERAL unnest(protocols_sets.protocol_ids) AS protocol_id ON true
			JOIN protocols ON protocols.id = protocol_id::integer
			GROUP BY protocols_sets.id
		`
}

// filteredPeers is a raw SQL query that returns the peer IDs that match the given filter
func filteredPeers(filter *coretypes.PeerFilter, latest bool) string {
	var conditions strings.Builder

	if filter != nil {
		if filter.Continents != nil {
			conditions.WriteString(AndIn("continents.continent_name", filter.Continents))
		}

		if filter.ContinentsCodes != nil {
			conditions.WriteString(AndIn("continents.code", filter.ContinentsCodes))
		}

		if filter.Countries != nil {
			conditions.WriteString(AndIn("countries.country_name", filter.Countries))
		}

		if filter.CountriesISOCodes != nil {
			conditions.WriteString(AndIn("countries.iso_code", filter.CountriesISOCodes))
		}

		if filter.Cities != nil {
			conditions.WriteString(AndIn("cities.city_name", filter.Cities))
		}

		if filter.ASOrganizations != nil {
			conditions.WriteString(AndIn("autonomous_systems.as_name", filter.ASOrganizations))
		}

		if filter.ASNumbers != nil {
			conditions.WriteString(AndIn("autonomous_systems.as_number", filter.ASNumbers))
		}
	}

	if latest {
		conditions.WriteString(" AND peers.id IN (SELECT peer_id FROM peers_in_latest_crawl)")
	}

	return fmt.Sprintf(
		`
		WITH peers_in_latest_crawl AS (
			%s
		)
		SELECT DISTINCT peers.id
			FROM peers
			JOIN peers_x_multi_addresses ON peers.id = peers_x_multi_addresses.peer_id
			JOIN multi_addresses ON peers_x_multi_addresses.multi_address_id = multi_addresses.id
			JOIN multi_addresses_x_ip_addresses ON multi_addresses.id = multi_addresses_x_ip_addresses.multi_address_id
			JOIN ip_addresses ON multi_addresses_x_ip_addresses.ip_address_id = ip_addresses.id
			LEFT JOIN autonomous_systems ON ip_addresses.as_id = autonomous_systems.id
			LEFT JOIN continents ON ip_addresses.continent_id = continents.id
			LEFT JOIN countries ON ip_addresses.country_id = countries.id
			LEFT JOIN cities ON ip_addresses.city_id = cities.id
			WHERE 1=1
			%s
		`,
		peersInLatestCrawl(),
		conditions.String(),
	)
}

// peerAddresses returns the multi addresses with embedded ip addresses and geo info
// for each peer
func peerAddresses() string {
	return `SELECT 
				peers.id AS peer_id,
				json_agg(
					json_build_object(
						'maddr', multi_addresses.maddr,
						'ip_info', (
							SELECT json_agg(
								json_build_object(
									'ip_address', ip_addresses.ip_address,
									'port', ip_addresses.port,
									'latitude', ip_addresses.latitude,
									'longitude', ip_addresses.longitude,
									'as_name', autonomous_systems.as_name,
									'as_number', autonomous_systems.as_number,
									'continent_name', continents.continent_name,
									'continent_code', continents.code,
									'country_name', countries.country_name,
									'country_iso', TRIM(countries.iso_code),
									'city_name', cities.city_name
								)
							)
							FROM multi_addresses_x_ip_addresses
							LEFT JOIN ip_addresses ON multi_addresses_x_ip_addresses.ip_address_id = ip_addresses.id
							LEFT JOIN autonomous_systems ON ip_addresses.as_id = autonomous_systems.id
							LEFT JOIN continents ON ip_addresses.continent_id = continents.id
							LEFT JOIN countries ON ip_addresses.country_id = countries.id
							LEFT JOIN cities ON ip_addresses.city_id = cities.id
							WHERE multi_addresses_x_ip_addresses.multi_address_id = multi_addresses.id
						)
					)
				) AS multi_addresses
			FROM peers
			JOIN peers_x_multi_addresses ON peers.id = peers_x_multi_addresses.peer_id
			JOIN multi_addresses ON peers_x_multi_addresses.multi_address_id = multi_addresses.id
			WHERE peers.id IN (SELECT peer_id FROM filtered_peers)
			GROUP BY peers.id
		`
}

// getSortColumn returns the exact table column name for the given sort column.
func getSortColumn(sort string) (string, error) {
	switch sort {
	case "", "last_seen":
		return "peers.last_seen", nil
	case sortColumnCreatedAt:
		return "peers.created_at", nil
	case sortColumnBlockHeight:
		return "latest_peer_states.block_height", nil
	default:
		return "", fmt.Errorf("invalid sort column: %s", sort)
	}
}

// getOrder returns the order string for the given IsAsc flag.
func getOrder(isAsc bool) string {
	if isAsc {
		return "ASC"
	}
	return "DESC"
}

type PaginationToken struct {
	LastPeerID    string
	LastSortValue interface{}
	SortColumn    string
	IsAsc         bool
	Filter        *coretypes.PeerFilter
}

func (p *PaginationToken) EncodeToString() (string, error) {
	tokenBytes, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(tokenBytes), nil
}

func decodePaginationToken(paginationToken string) (*PaginationToken, error) {
	decodedToken, err := base64.StdEncoding.DecodeString(paginationToken)
	if err != nil {
		return nil, fmt.Errorf("invalid pagination token: %w", err)
	}
	var token PaginationToken
	err = json.Unmarshal(decodedToken, &token)
	if err != nil {
		return nil, fmt.Errorf("invalid pagination token: %w", err)
	}
	return &token, nil
}
