package repo

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/lib/pq"

	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

const (
	sortColumnBlockHeight = "block_height"
	sortColumnCreatedAt   = "created_at"
	sortColumnLastSeen    = "last_seen"
)

// PeerIndexEntry represents a denormalized entry in the peer_visits_index table
type PeerIndexEntry struct {
	CrawlID       int
	VisitID       int64
	PeerID        int64
	PeerMultiHash string
	AgentVersion  null.String
	ClientName    null.String
	ContinentName null.String
	ContinentCode null.String
	CountryName   null.String
	CountryISO    null.String
	CityName      null.String
	Latitude      null.Float64
	Longitude     null.Float64
	ASName        null.String
	ASNumber      null.Int
	IsSynced      null.Bool
	BlockHeight   null.Int64
	SpecVersion   null.String
	CreatedAt     time.Time
	LastSeen      time.Time
}

// CreatePeerVisitWithIndex creates a peer visit and its index entry in a single transaction
func (r *PeerRepository) CreatePeerVisitWithIndex(
	ctx context.Context,
	crawlID int,
	visit *coretypes.Visit,
	peer *coretypes.Peer,
	peerState *coretypes.PeerChainInfo,
	firstGeoInfo *coretypes.GeoInfo,
) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		// Prepare peer data
		r.preparePeerData(peer)

		// Process peer components
		avID, psID, multiAddrs, maIDs, err := r.processPeerComponents(ctx, tx, peer)
		if err != nil {
			return err
		}

		// Create or update peer
		peerID, err := r.findOrUpsertPeer(ctx, tx, peer.PeerID,
			null.NewInt(avID, avID != 0),
			null.NewInt(psID, psID != 0),
			multiAddrs)
		if err != nil {
			return fmt.Errorf("find or upsert peer: %w", err)
		}

		// Create visit record
		visitID, err := r.insertVisit(ctx, tx,
			peerID,
			null.NewInt(avID, avID != 0),
			null.NewInt(psID, psID != 0),
			visit,
			maIDs,
			null.IntFrom(crawlID),
		)
		if err != nil {
			return fmt.Errorf("insert visit: %w", err)
		}

		// Create peer chain info if available
		if peerState != nil {
			err = r.createPeerChainInfoTx(ctx, tx, peerID, peerState)
			if err != nil {
				return fmt.Errorf("create peer chain info: %w", err)
			}
		}

		// Build and insert index entry
		return r.buildAndInsertIndexEntry(ctx, tx, crawlID, visitID, peerID, peer, peerState, firstGeoInfo)
	})
}

// preparePeerData ensures peer data has required fields
func (r *PeerRepository) preparePeerData(peer *coretypes.Peer) {
	if peer.AgentVersion == "" {
		peer.AgentVersion = "aztec"
	}
}

// processPeerComponents processes agent version, protocols, and multiaddresses
func (r *PeerRepository) processPeerComponents(
	ctx context.Context,
	tx *sql.Tx,
	peer *coretypes.Peer,
) (int, int, []*models.MultiAddress, []int64, error) {
	// Process agent version
	avID, err := r.upsertAgentVersion(ctx, tx, peer.AgentVersion, 0)
	if err != nil {
		return 0, 0, nil, nil, fmt.Errorf("upsert agent version: %w", err)
	}

	// Process protocols
	protocolMap := make(map[string]int)
	for _, p := range peer.Protocols {
		protocolMap[p] = 0
	}
	psID, err := r.upsertProtocolsWithSet(ctx, tx, protocolMap)
	if err != nil {
		return 0, 0, nil, nil, fmt.Errorf("upsert protocols: %w", err)
	}

	// Process multiaddresses
	multiAddrs, maIDs, err := r.insertMultiAddrs(ctx, tx, peer.MultiAddrs)
	if err != nil {
		return 0, 0, nil, nil, fmt.Errorf("insert multiaddrs: %w", err)
	}

	return avID, psID, multiAddrs, maIDs, nil
}

// buildAndInsertIndexEntry builds the index entry and inserts it
func (r *PeerRepository) buildAndInsertIndexEntry(
	ctx context.Context,
	tx *sql.Tx,
	crawlID int,
	visitID int64,
	peerID int64,
	peer *coretypes.Peer,
	peerState *coretypes.PeerChainInfo,
	firstGeoInfo *coretypes.GeoInfo,
) error {
	// Get peer timestamps
	peerModel, err := models.FindPeer(ctx, tx, peerID)
	if err != nil {
		return fmt.Errorf("find peer for timestamps: %w", err)
	}

	// Build base index entry
	indexEntry := r.buildBaseIndexEntry(crawlID, visitID, peerID, peer, peerModel)

	// Add geo info
	r.addGeoInfoToIndexEntry(&indexEntry, firstGeoInfo)

	// Add peer state
	r.addPeerStateToIndexEntry(&indexEntry, peerState)

	// Insert index entry
	err = insertIndexEntry(ctx, tx, &indexEntry)
	if err != nil {
		return fmt.Errorf("insert index entry: %w", err)
	}

	return nil
}

// buildBaseIndexEntry creates the base index entry
func (r *PeerRepository) buildBaseIndexEntry(
	crawlID int,
	visitID int64,
	peerID int64,
	peer *coretypes.Peer,
	peerModel *models.Peer,
) PeerIndexEntry {
	return PeerIndexEntry{
		CrawlID:       crawlID,
		VisitID:       visitID,
		PeerID:        peerID,
		PeerMultiHash: peer.PeerID,
		AgentVersion:  null.StringFrom(peer.AgentVersion),
		ClientName:    null.StringFrom(extractClientName(peer.AgentVersion)),
		CreatedAt:     peerModel.CreatedAt,
		LastSeen:      peerModel.LastSeen,
	}
}

// addGeoInfoToIndexEntry adds geographic information to the index entry
func (r *PeerRepository) addGeoInfoToIndexEntry(entry *PeerIndexEntry, geoInfo *coretypes.GeoInfo) {
	if geoInfo == nil {
		return
	}

	entry.ContinentName = null.StringFrom(geoInfo.Continent)
	entry.ContinentCode = null.StringFrom(geoInfo.ContinentCode)
	entry.CountryName = null.StringFrom(geoInfo.Country)
	entry.CountryISO = null.StringFrom(geoInfo.CountryISO)
	entry.CityName = null.StringFrom(geoInfo.City)
	entry.Latitude = null.Float64From(geoInfo.Latitude)
	entry.Longitude = null.Float64From(geoInfo.Longitude)
	entry.ASName = null.StringFrom(geoInfo.ASOrganization)

	if geoInfo.ASNumber > 0 {
		entry.ASNumber = null.IntFrom(int(min(geoInfo.ASNumber, math.MaxInt32))) //nolint:gosec // bounded by min()
	}
}

// addPeerStateToIndexEntry adds peer state information to the index entry
func (r *PeerRepository) addPeerStateToIndexEntry(entry *PeerIndexEntry, peerState *coretypes.PeerChainInfo) {
	if peerState == nil {
		return
	}

	if peerState.SyncStatusObtained {
		entry.IsSynced = null.BoolFrom(peerState.SyncStatus)
	}

	if peerState.BlockHeightObtained {
		// BlockHeight is uint64, need safe conversion to int64
		const maxInt64 = 1<<63 - 1
		if peerState.BlockHeight <= maxInt64 {
			entry.BlockHeight = null.Int64From(int64(peerState.BlockHeight))
		}
	}

	if peerState.SpecVersionObtained {
		entry.SpecVersion = null.StringFrom(peerState.SpecVersion)
	}
}

// createPeerChainInfoTx creates peer chain info within a transaction
func (r *PeerRepository) createPeerChainInfoTx(
	ctx context.Context,
	tx *sql.Tx,
	peerID int64,
	info *coretypes.PeerChainInfo,
) error {
	peerState := &models.PeersState{
		PeerID:    peerID,
		CreatedAt: time.Now(),
	}

	if info.BlockHeightObtained {
		// BlockHeight is uint64, need safe conversion to int64
		const maxInt64 = 1<<63 - 1
		if info.BlockHeight <= maxInt64 {
			peerState.BlockHeight = null.Int64From(int64(info.BlockHeight))
		}
	}

	if info.SpecVersionObtained {
		peerState.SpecVersion = null.StringFrom(info.SpecVersion)
	}

	if info.SyncStatusObtained {
		peerState.IsSynced = null.BoolFrom(info.SyncStatus)
	}

	err := peerState.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return fmt.Errorf("insert peer state: %w", err)
	}

	return nil
}

// extractClientName extracts the client name from an agent version string
func extractClientName(agentVersion string) string {
	if agentVersion == "" {
		return ""
	}
	// Extract the part before the first '/' or return the whole string if no '/' found
	parts := strings.Split(agentVersion, "/")
	return strings.ToLower(parts[0])
}

// insertIndexEntry inserts a single index entry
func insertIndexEntry(ctx context.Context, tx *sql.Tx, entry *PeerIndexEntry) error {
	query := `
		INSERT INTO peer_visits_index (
			crawl_id, visit_id, peer_id, peer_multi_hash,
			agent_version, client_name,
			continent_name, continent_code, country_name, country_iso, city_name,
			latitude, longitude,
			as_name, as_number,
			is_synced, block_height, spec_version,
			created_at, last_seen
		) VALUES (
			$1, $2, $3, $4,
			$5, $6,
			$7, $8, $9, $10, $11,
			$12, $13,
			$14, $15,
			$16, $17, $18,
			$19, $20
		)`

	_, err := tx.ExecContext(ctx, query,
		entry.CrawlID, entry.VisitID, entry.PeerID, entry.PeerMultiHash,
		entry.AgentVersion, entry.ClientName,
		entry.ContinentName, entry.ContinentCode, entry.CountryName, entry.CountryISO, entry.CityName,
		entry.Latitude, entry.Longitude,
		entry.ASName, entry.ASNumber,
		entry.IsSynced, entry.BlockHeight, entry.SpecVersion,
		entry.CreatedAt, entry.LastSeen,
	)
	return err
}

// GetPeersForMap retrieves peer geographic data directly from the index for map visualisation
func (r *PeerRepository) GetPeersForMap(
	ctx context.Context,
	options *PeerQueryOptions,
) ([]*PeerIndexEntry, error) {
	// Build query to get geographic data from index
	query := `
		WITH latest_crawl AS (
			SELECT MAX(id) as crawl_id FROM crawls WHERE status = 'finished'
		)
		SELECT DISTINCT ON (peer_id)
			peer_id,
			peer_multi_hash,
			city_name,
			country_name,
			continent_name,
			latitude,
			longitude
		FROM peer_visits_index, latest_crawl
		WHERE peer_visits_index.crawl_id = latest_crawl.crawl_id
			AND latitude IS NOT NULL 
			AND longitude IS NOT NULL
			AND city_name IS NOT NULL
	`

	args := []any{}

	// Add filter conditions if provided
	if options.Filter != nil {
		conditions := r.buildIndexFilterConditions(options.Filter, &args)
		if len(conditions) > 0 {
			query += " AND " + strings.Join(conditions, " AND ")
		}
	}

	query += " ORDER BY peer_id"

	// Use a large limit for map data
	if options.PageSize <= 0 || options.PageSize > 50000 {
		options.PageSize = 50000
	}
	query += fmt.Sprintf(" LIMIT %d", options.PageSize)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query map data from index: %w", err)
	}
	defer rows.Close()

	var entries []*PeerIndexEntry
	for rows.Next() {
		var entry PeerIndexEntry
		err := rows.Scan(
			&entry.PeerID,
			&entry.PeerMultiHash,
			&entry.CityName,
			&entry.CountryName,
			&entry.ContinentName,
			&entry.Latitude,
			&entry.Longitude,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan map entry: %w", err)
		}
		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating map results: %w", err)
	}

	return entries, nil
}

// GetPeersOptimized retrieves peers using the optimised two-step query approach
func (r *PeerRepository) GetPeersOptimized(
	ctx context.Context,
	opt *PeerQueryOptions,
) ([]*PeerInfo, string, int64, error) {
	// Step 1: Get total count from index
	totalCount, err := r.getPeersOptimizedTotalCount(ctx, opt.Filter)
	if err != nil {
		return nil, "", 0, fmt.Errorf("error getting total peer count: %w", err)
	}

	// Step 2: Get filtered peer IDs and basic info from index
	indexEntries, err := r.queryPeerIndex(ctx, opt)
	if err != nil {
		return nil, "", 0, fmt.Errorf("error querying peer index: %w", err)
	}

	if len(indexEntries) == 0 {
		return []*PeerInfo{}, "", totalCount, nil
	}

	// Step 3: Enrich with full peer data
	peers, err := r.enrichPeerData(ctx, indexEntries)
	if err != nil {
		return nil, "", 0, fmt.Errorf("error enriching peer data: %w", err)
	}

	// Step 4: Create pagination token
	pageSize := opt.PageSize
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	nextPageToken := ""
	if len(peers) > pageSize {
		peers = peers[:pageSize]
		lastPeer := peers[pageSize-1]
		lastEntry := indexEntries[pageSize-1] // Use the index entry for sort value
		token := PaginationToken{
			LastPeerID:    lastPeer.PeerID,
			LastSortValue: r.getIndexLastSortValue(lastEntry, opt.Sort),
			SortColumn:    opt.Sort,
			IsAsc:         opt.IsAsc,
			Filter:        opt.Filter,
		}
		nextPageToken, _ = token.EncodeToString()
	}

	return peers, nextPageToken, totalCount, nil
}

// getPeersOptimizedTotalCount gets the total count using the index table
func (r *PeerRepository) getPeersOptimizedTotalCount(
	ctx context.Context,
	filter *coretypes.PeerFilter,
) (int64, error) {
	query := `
		WITH latest_crawl AS (
			SELECT MAX(id) as crawl_id FROM crawls WHERE status = 'finished'
		)
		SELECT COUNT(DISTINCT peer_id)
		FROM peer_visits_index, latest_crawl
		WHERE peer_visits_index.crawl_id = latest_crawl.crawl_id
	`

	args := []any{}
	conditions := r.buildIndexFilterConditions(filter, &args)
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get peer count from index: %w", err)
	}

	return count, nil
}

// queryPeerIndex queries the denormalized index table
func (r *PeerRepository) queryPeerIndex(ctx context.Context, opt *PeerQueryOptions) ([]*PeerIndexEntry, error) {
	pageSize := opt.PageSize
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	// Build the query
	query := `
		WITH latest_crawl AS (
			SELECT MAX(id) as crawl_id FROM crawls WHERE status = 'finished'
		)
		SELECT 
			visit_id,
			peer_id,
			peer_multi_hash,
			agent_version,
			client_name,
			last_seen,
			block_height,
			created_at,
			is_synced,
			spec_version
		FROM peer_visits_index, latest_crawl
		WHERE peer_visits_index.crawl_id = latest_crawl.crawl_id
	`

	args := []any{}

	// Add filter conditions
	conditions := r.buildIndexFilterConditions(opt.Filter, &args)
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	// Add pagination condition if token provided
	if opt.PaginationToken != "" {
		token, err := decodePaginationToken(opt.PaginationToken)
		if err != nil {
			return nil, fmt.Errorf("invalid pagination token: %w", err)
		}

		sortCol := r.getIndexSortColumn(opt.Sort)
		argIdx := len(args) + 1

		if opt.IsAsc {
			query += fmt.Sprintf(" AND (%s > $%d OR (%s = $%d AND peer_multi_hash > $%d))",
				sortCol, argIdx, sortCol, argIdx, argIdx+1)
		} else {
			query += fmt.Sprintf(" AND (%s < $%d OR (%s = $%d AND peer_multi_hash < $%d))",
				sortCol, argIdx, sortCol, argIdx, argIdx+1)
		}
		args = append(args, token.LastSortValue, token.LastPeerID)
	}

	// Add ORDER BY
	sortCol := r.getIndexSortColumn(opt.Sort)
	order := "DESC"
	if opt.IsAsc {
		order = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s, peer_multi_hash %s", sortCol, order, order)

	// Add LIMIT (fetch one extra for pagination)
	query += fmt.Sprintf(" LIMIT %d", pageSize+1)

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query peer index: %w", err)
	}
	defer rows.Close()

	entries := []*PeerIndexEntry{}
	for rows.Next() {
		entry := &PeerIndexEntry{}
		err := rows.Scan(
			&entry.VisitID,
			&entry.PeerID,
			&entry.PeerMultiHash,
			&entry.AgentVersion,
			&entry.ClientName,
			&entry.LastSeen,
			&entry.BlockHeight,
			&entry.CreatedAt,
			&entry.IsSynced,
			&entry.SpecVersion,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan index entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// buildIndexFilterConditions builds WHERE conditions for the index table
func (r *PeerRepository) buildIndexFilterConditions(filter *coretypes.PeerFilter, args *[]any) []string {
	conditions := []string{}

	if filter == nil {
		return conditions
	}

	// Peer ID filter (partial match, case-insensitive)
	if filter.PeerID != "" {
		*args = append(*args, strings.ToLower(filter.PeerID)+"%")
		conditions = append(conditions, fmt.Sprintf("LOWER(peer_multi_hash) LIKE $%d", len(*args)))
	}

	// Client filters
	if len(filter.ClientNames) > 0 {
		*args = append(*args, pq.Array(filter.ClientNames))
		conditions = append(conditions, fmt.Sprintf("client_name = ANY($%d)", len(*args)))
	}

	// Geographic filters
	if len(filter.Continents) > 0 {
		*args = append(*args, pq.Array(filter.Continents))
		conditions = append(conditions, fmt.Sprintf("continent_name = ANY($%d)", len(*args)))
	}

	if len(filter.ContinentsCodes) > 0 {
		*args = append(*args, pq.Array(filter.ContinentsCodes))
		conditions = append(conditions, fmt.Sprintf("continent_code = ANY($%d)", len(*args)))
	}

	if len(filter.Countries) > 0 {
		*args = append(*args, pq.Array(filter.Countries))
		conditions = append(conditions, fmt.Sprintf("country_name = ANY($%d)", len(*args)))
	}

	if len(filter.CountriesISOCodes) > 0 {
		*args = append(*args, pq.Array(filter.CountriesISOCodes))
		conditions = append(conditions, fmt.Sprintf("country_iso = ANY($%d)", len(*args)))
	}

	if len(filter.Cities) > 0 {
		*args = append(*args, pq.Array(filter.Cities))
		conditions = append(conditions, fmt.Sprintf("city_name = ANY($%d)", len(*args)))
	}

	// AS filters
	if len(filter.ASOrganizations) > 0 {
		*args = append(*args, pq.Array(filter.ASOrganizations))
		conditions = append(conditions, fmt.Sprintf("as_name = ANY($%d)", len(*args)))
	}

	if len(filter.ASNumbers) > 0 {
		*args = append(*args, pq.Array(filter.ASNumbers))
		conditions = append(conditions, fmt.Sprintf("as_number = ANY($%d)", len(*args)))
	}

	// Sync status filter
	if filter.SyncStatus != nil {
		*args = append(*args, *filter.SyncStatus)
		conditions = append(conditions, fmt.Sprintf("is_synced = $%d", len(*args)))
	}

	// Spec version (CCV) filter
	if filter.SpecVersion != "" {
		*args = append(*args, filter.SpecVersion)
		conditions = append(conditions, fmt.Sprintf("spec_version = $%d", len(*args)))
	}

	return conditions
}

// getIndexSortColumn returns the column name in the index table for sorting
func (r *PeerRepository) getIndexSortColumn(sort string) string {
	switch sort {
	case sortColumnBlockHeight:
		return sortColumnBlockHeight
	case sortColumnCreatedAt:
		return sortColumnCreatedAt
	case sortColumnLastSeen, "":
		return sortColumnLastSeen
	default:
		return sortColumnLastSeen
	}
}

// getIndexLastSortValue gets the sort value from an index entry for pagination
func (r *PeerRepository) getIndexLastSortValue(entry *PeerIndexEntry, sortColumn string) any {
	switch sortColumn {
	case sortColumnBlockHeight:
		if entry.BlockHeight.Valid {
			return entry.BlockHeight.Int64
		}
		return int64(0)
	case sortColumnCreatedAt:
		return entry.CreatedAt
	case sortColumnLastSeen, "":
		return entry.LastSeen
	default:
		return entry.LastSeen
	}
}

// enrichPeerData fetches complete peer information for the given index entries
func (r *PeerRepository) enrichPeerData(ctx context.Context, entries []*PeerIndexEntry) ([]*PeerInfo, error) {
	if len(entries) == 0 {
		return []*PeerInfo{}, nil
	}

	peerIDs, peerIDMap := r.collectPeerIDs(entries)
	if len(peerIDs) == 0 {
		return []*PeerInfo{}, nil
	}

	peers, err := r.fetchPeersWithRelations(ctx, peerIDs)
	if err != nil {
		return nil, err
	}

	result, err := r.convertToPeerInfos(ctx, peers, peerIDMap)
	if err != nil {
		return nil, err
	}

	return r.maintainOriginalOrder(entries, result), nil
}

// collectPeerIDs collects peer IDs from index entries
func (r *PeerRepository) collectPeerIDs(entries []*PeerIndexEntry) ([]int64, map[string]*PeerIndexEntry) {
	peerIDs := make([]int64, 0, len(entries))
	peerIDMap := make(map[string]*PeerIndexEntry)

	for _, entry := range entries {
		peerIDs = append(peerIDs, entry.PeerID)
		peerIDMap[entry.PeerMultiHash] = entry
	}

	return peerIDs, peerIDMap
}

// fetchPeersWithRelations fetches peers with their relations
func (r *PeerRepository) fetchPeersWithRelations(ctx context.Context, peerIDs []int64) ([]*models.Peer, error) {
	peers, err := models.Peers(
		qm.WhereIn("id IN ?", getInterfaceSlice64(peerIDs)...),
		qm.Load(models.PeerRels.AgentVersion),
		qm.Load(models.PeerRels.ProtocolsSet),
		qm.Load(models.PeerRels.MultiAddresses),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch peers: %w", err)
	}
	return peers, nil
}

// convertToPeerInfos converts database models to PeerInfo structs
func (r *PeerRepository) convertToPeerInfos(
	ctx context.Context,
	peers []*models.Peer,
	peerIDMap map[string]*PeerIndexEntry,
) ([]*PeerInfo, error) {
	result := make([]*PeerInfo, 0, len(peers))

	for _, peer := range peers {
		peerInfo := &PeerInfo{
			PeerID:    peer.MultiHash,
			CreatedAt: peer.CreatedAt,
			LastSeen:  peer.LastSeen,
		}

		if err := r.enrichPeerInfo(ctx, peerInfo, peer, peerIDMap); err != nil {
			return nil, err
		}

		result = append(result, peerInfo)
	}

	return result, nil
}

// enrichPeerInfo adds additional data to a PeerInfo struct
func (r *PeerRepository) enrichPeerInfo(
	ctx context.Context,
	peerInfo *PeerInfo,
	peer *models.Peer,
	peerIDMap map[string]*PeerIndexEntry,
) error {
	if peer.R == nil {
		return r.addIndexData(peerInfo, peer.MultiHash, peerIDMap)
	}

	// Add agent version
	if peer.R.AgentVersion != nil {
		peerInfo.AgentVersion = null.StringFrom(peer.R.AgentVersion.AgentVersion)
	}

	// Add protocols
	if err := r.addProtocols(ctx, peerInfo, peer.R.ProtocolsSet); err != nil {
		return err
	}

	// Add multiaddresses
	if err := r.addMultiAddresses(ctx, peerInfo, peer.R.MultiAddresses); err != nil {
		return err
	}

	return r.addIndexData(peerInfo, peer.MultiHash, peerIDMap)
}

// addProtocols adds protocol information to PeerInfo
func (r *PeerRepository) addProtocols(ctx context.Context, peerInfo *PeerInfo, protocolsSet *models.ProtocolsSet) error {
	if protocolsSet == nil || len(protocolsSet.ProtocolIds) == 0 {
		return nil
	}

	protocols, err := models.Protocols(
		qm.WhereIn("id IN ?", convertInt64ArrayToInterface(protocolsSet.ProtocolIds)...),
	).All(ctx, r.db)
	if err != nil {
		return nil // Non-critical error, continue without protocols
	}

	protocolNames := make([]string, len(protocols))
	for i, p := range protocols {
		protocolNames[i] = p.Protocol
	}
	peerInfo.Protocols = protocolNames
	return nil
}

// addMultiAddresses adds multiaddress information to PeerInfo
func (r *PeerRepository) addMultiAddresses(ctx context.Context, peerInfo *PeerInfo, multiAddresses []*models.MultiAddress) error {
	if len(multiAddresses) == 0 {
		return nil
	}

	multiAddrs := make([]*coretypes.MultiAddr, 0, len(multiAddresses))
	for _, maddr := range multiAddresses {
		ma := &coretypes.MultiAddr{
			Address: maddr.Maddr,
			IPList:  []*coretypes.IPInfo{},
		}

		// Fetch IP info for this multiaddress
		ipInfos, err := r.fetchIPInfoForMultiAddr(ctx, maddr.ID)
		if err == nil {
			ma.IPList = ipInfos
		}

		multiAddrs = append(multiAddrs, ma)
	}
	peerInfo.MultiAddresses = multiAddrs
	return nil
}

// addIndexData adds index-specific data to PeerInfo
func (r *PeerRepository) addIndexData(peerInfo *PeerInfo, multiHash string, peerIDMap map[string]*PeerIndexEntry) error {
	if entry, ok := peerIDMap[multiHash]; ok {
		peerInfo.BlockHeight = entry.BlockHeight
		peerInfo.IsSynced = entry.IsSynced
		peerInfo.SpecVersion = entry.SpecVersion
	}
	return nil
}

// maintainOriginalOrder maintains the original order from index query
func (r *PeerRepository) maintainOriginalOrder(entries []*PeerIndexEntry, peerInfos []*PeerInfo) []*PeerInfo {
	peerMap := make(map[string]*PeerInfo)
	for _, peer := range peerInfos {
		peerMap[peer.PeerID] = peer
	}

	orderedPeers := make([]*PeerInfo, 0, len(entries))
	for _, entry := range entries {
		if peer, ok := peerMap[entry.PeerMultiHash]; ok {
			orderedPeers = append(orderedPeers, peer)
		}
	}

	return orderedPeers
}

// fetchIPInfoForMultiAddr fetches IP information for a multiaddress
func (r *PeerRepository) fetchIPInfoForMultiAddr(ctx context.Context, multiAddrID int) ([]*coretypes.IPInfo, error) {
	query := `
		SELECT 
			ip.ip_address,
			ip.port,
			ip.latitude,
			ip.longitude,
			asys.as_name,
			asys.as_number,
			cont.continent_name,
			cont.code as continent_code,
			coun.country_name,
			TRIM(coun.iso_code) as country_iso,
			cit.city_name
		FROM multi_addresses_x_ip_addresses mxip
		LEFT JOIN ip_addresses ip ON mxip.ip_address_id = ip.id
		LEFT JOIN autonomous_systems asys ON ip.as_id = asys.id
		LEFT JOIN continents cont ON ip.continent_id = cont.id
		LEFT JOIN countries coun ON ip.country_id = coun.id
		LEFT JOIN cities cit ON ip.city_id = cit.id
		WHERE mxip.multi_address_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, multiAddrID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ipInfos []*coretypes.IPInfo
	for rows.Next() {
		var (
			ipAddress     string
			port          int
			latitude      sql.NullFloat64
			longitude     sql.NullFloat64
			asName        sql.NullString
			asNumber      sql.NullInt64
			continentName sql.NullString
			continentCode sql.NullString
			countryName   sql.NullString
			countryISO    sql.NullString
			cityName      sql.NullString
		)

		err := rows.Scan(
			&ipAddress,
			&port,
			&latitude,
			&longitude,
			&asName,
			&asNumber,
			&continentName,
			&continentCode,
			&countryName,
			&countryISO,
			&cityName,
		)
		if err != nil {
			continue
		}

		ipInfo := &coretypes.IPInfo{
			IPAddress: ipAddress,
			Port:      port,
			GeoInfo: &coretypes.GeoInfo{
				ASOrganization: asName.String,
				ASNumber:       uint(max(0, min(asNumber.Int64, math.MaxUint32))), //nolint:gosec // bounded by min/max
				City:           cityName.String,
				Country:        countryName.String,
				CountryISO:     countryISO.String,
				Continent:      continentName.String,
				ContinentCode:  continentCode.String,
				Latitude:       latitude.Float64,
				Longitude:      longitude.Float64,
			},
		}

		ipInfos = append(ipInfos, ipInfo)
	}

	return ipInfos, nil
}

// getInterfaceSlice64 converts []int64 to []interface{} for use with WhereIn
func getInterfaceSlice64(ids []int64) []any {
	result := make([]any, len(ids))
	for i, id := range ids {
		result[i] = id
	}
	return result
}

// convertInt64ArrayToInterface converts types.Int64Array to []interface{}
func convertInt64ArrayToInterface(ids types.Int64Array) []any {
	result := make([]any, len(ids))
	for i, id := range ids {
		result[i] = int(id)
	}
	return result
}
