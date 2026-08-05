//nolint:misspell,dupl
package server

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mitchellh/hashstructure/v2"

	"github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/repo"
)

const (
	MaxPageSize        = 200
	MaxPrivatePageSize = 1000
)

//nolint:unparam
func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(DefaultPrecision, float64(precision))
	return math.Round(val*ratio) / ratio
}

func cleanPeer(peer *repo.PeerInfo) *repo.PeerInfo {
	newMultiAddrs := make([]*types.MultiAddr, 0, len(peer.MultiAddresses))
	for _, multiAddr := range peer.MultiAddresses {
		ipList := make([]*types.IPInfo, 0, len(multiAddr.IPList))
		for _, ipInfo := range multiAddr.IPList {
			ipList = append(ipList, &types.IPInfo{
				IPAddress: "",
				Port:      0,
				GeoInfo: &types.GeoInfo{
					ASOrganization: "",
					ASNumber:       0,
					City:           ipInfo.City,
					Country:        ipInfo.Country,
					CountryISO:     ipInfo.CountryISO,
					Continent:      ipInfo.Continent,
					ContinentCode:  ipInfo.ContinentCode,
					Latitude:       roundFloat(ipInfo.Latitude, 1),
					Longitude:      roundFloat(ipInfo.Longitude, 1),
				},
			})
		}
		newMultiAddrs = append(newMultiAddrs, &types.MultiAddr{
			Address: "",
			IPList:  ipList,
		})
	}
	peer.MultiAddresses = newMultiAddrs
	return peer
}

func clearPeersMultiAddrs(peers []*repo.PeerInfo) []*repo.PeerInfo {
	for _, peer := range peers {
		cleanPeer(peer)
	}
	return peers
}

// GetPeers returns a list of peers without IP addresses based on the provided query options.
//
//	@Summary		Get a list of peers
//	@Description	Get a list of peers based on the provided query options
//	@Tags			Peers
//	@Accept			json
//	@Produce		json
//	@Param			id				query		string		false	"Peer ID (e.g. 12...)"
//	@Param			continents		query		[]string	false	"Continent names"
//	@Param 			continents_code query 		[]string 	false 	"Continent codes"
//	@Param			countries		query		[]string	false	"Country names"
//	@Param 			countries_codes query 		[]string 	false 	"Country iso codes"
//	@Param			cities			query		[]string	false	"City names"
//	@Param			as_names		query		[]string	false	"Names of Autonomous System Organizations"
//	@Param			as_numbers		query		[]string	false	"Numbers of Autonomous System Organizations"
//	@Param			clients			query		[]string	false	"Client names (e.g. aztec-node)"
//	@Param			synced			query		bool		false	"Sync status"
//	@Param			sort			query		string		false	"Sort column (created_at, last_seen, block_height)"
//	@Param			order			query		string		false	"Sort order (asc or desc)"
//	@Param			latest			query		bool		false	"Include peers from latest crawl only (default: false, all crawls)"
//	@Param			page_size		query		int			false	"Number of items per page"
//	@Param			pagination_token	query	string		false	"Pagination token"
//	@Success		200	{object}	PeersResponse
//	@Failure		400	{object}	APIError
//	@Failure		500	{object}	APIError
//	@Router			/peers [get]
func (h *Handler) GetPeers(c echo.Context) error {
	var filter types.PeerFilter
	if err := c.Bind(&filter); err != nil {
		return InvalidParameters(err)
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))

	// Cap page size to prevent oversized requests
	if pageSize > MaxPageSize {
		return InvalidParameters(fmt.Errorf("page_size cannot exceed %d", MaxPageSize))
	}

	isLatest := c.QueryParam("latest") == "true"

	// Apply CCV filter when latest=true and CCV is available from reference node
	if isLatest {
		if ccv := h.GetFilterCCV(); ccv != "" {
			filter.SpecVersion = ccv
		}
	}

	options := &repo.PeerQueryOptions{
		Filter:          &filter,
		Sort:            c.QueryParam("sort"),
		IsAsc:           c.QueryParam("order") == "asc",
		Latest:          isLatest,
		PageSize:        pageSize,
		PaginationToken: c.QueryParam("pagination_token"),
	}

	optionsHash, err := hashstructure.Hash(options, hashstructure.FormatV2, nil)
	if err == nil && h.PeersResponseCache != nil {
		if cachedPeer, ok := h.PeersResponseCache.Get(optionsHash); ok {
			return c.JSON(http.StatusOK, cachedPeer)
		}
	}

	// Use optimised index table if enabled, otherwise use the original implementation
	var peers []*repo.PeerInfo
	var nextToken string
	var totalCount int64

	if h.UseIndexTable {
		peers, nextToken, totalCount, err = h.Repo.GetPeersOptimized(c.Request().Context(), options)
	} else {
		peers, nextToken, totalCount, err = h.Repo.GetPeers(c.Request().Context(), options)
	}

	if err != nil {
		return ServerError(err)
	}

	peers = clearPeersMultiAddrs(peers)

	response := &PeersResponse{
		Peers:               peers,
		NextPaginationToken: nextToken,
		TotalPeers:          totalCount,
	}

	if h.PeersResponseCache != nil {
		h.PeersResponseCache.Add(optionsHash, *response)
	}

	return c.JSON(http.StatusOK, response)
}

// GetPeersPrivate returns a list of peers with IP addresses based on the provided query options.
//
//	@Summary		Get a list of peers
//	@Description	Get a list of peers based on the provided query options
//	@Tags			Peers
//	@Accept			json
//	@Produce		json
//	@Param			id				query		string		false	"Peer ID (e.g. 12...)"
//	@Param			continents		query		[]string	false	"Continent names"
//	@Param 			continents_code query 		[]string 	false 	"Continent codes"
//	@Param			countries		query		[]string	false	"Country names"
//	@Param 			countries_codes query 		[]string 	false 	"Country iso codes"
//	@Param			cities			query		[]string	false	"City names"
//	@Param			as_names		query		[]string	false	"Names of Autonomous System Organizations"
//	@Param			as_numbers		query		[]string	false	"Numbers of Autonomous System Organizations"
//	@Param			clients			query		[]string	false	"Client names (e.g. aztec-node)"
//	@Param			synced			query		bool		false	"Sync status"
//	@Param			sort			query		string		false	"Sort column (created_at, last_seen, block_height)"
//	@Param			order			query		string		false	"Sort order (asc or desc)"
//	@Param			latest			query		bool		false	"Include peers from latest crawl only (default: false, all crawls)"
//	@Param			page_size		query		int			false	"Number of items per page"
//	@Param			pagination_token	query	string		false	"Pagination token"
//	@Success		200	{object}	PeersResponse
//	@Failure		400	{object}	APIError
//	@Failure		500	{object}	APIError
//	@Router			/peers [get]
func (h *Handler) GetPeersPrivate(c echo.Context) error {
	var filter types.PeerFilter
	if err := c.Bind(&filter); err != nil {
		return InvalidParameters(err)
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))

	// Cap page size to prevent oversized requests
	if pageSize > MaxPrivatePageSize {
		return InvalidParameters(fmt.Errorf("page_size cannot exceed %d", MaxPrivatePageSize))
	}

	isLatest := c.QueryParam("latest") == "true"

	// Apply CCV filter when latest=true and CCV is available from reference node
	if isLatest {
		if ccv := h.GetFilterCCV(); ccv != "" {
			filter.SpecVersion = ccv
		}
	}

	options := &repo.PeerQueryOptions{
		Filter:          &filter,
		Sort:            c.QueryParam("sort"),
		IsAsc:           c.QueryParam("order") == "asc",
		Latest:          isLatest,
		PageSize:        pageSize,
		PaginationToken: c.QueryParam("pagination_token"),
	}

	// Use optimised index table if enabled, otherwise use the original implementation
	var peers []*repo.PeerInfo
	var nextToken string
	var totalCount int64
	var err error

	if h.UseIndexTable {
		peers, nextToken, totalCount, err = h.Repo.GetPeersOptimized(c.Request().Context(), options)
	} else {
		peers, nextToken, totalCount, err = h.Repo.GetPeers(c.Request().Context(), options)
	}

	if err != nil {
		return ServerError(err)
	}

	return c.JSON(http.StatusOK, PeersResponse{
		Peers:               peers,
		NextPaginationToken: nextToken,
		TotalPeers:          totalCount,
	})
}

// GetPeerByID returns a peer by its ID.
//
//	@Summary		Get a peer by ID
//	@Description	Get detailed information about a specific peer
//	@Tags			Peers
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Peer ID (e.g. 12...)"
//	@Success		200	{object}	PeerResponse
//	@Failure		404	{object}	APIError
//	@Failure		500	{object}	APIError
//	@Router			/peers/{id} [get]
func (h *Handler) GetPeerByID(c echo.Context) error {
	id := c.Param("id")
	peer, err := h.Repo.GetPeerByID(c.Request().Context(), id)
	if err != nil {
		return ServerError(err)
	}
	if peer == nil {
		return RecordWithStringIDNotFound(id)
	}
	peer = cleanPeer(peer)
	return c.JSON(http.StatusOK, PeerResponse{peer})
}

// GetPeersNeighbors returns the neighbors of all peers
//
//	@Summary		Get every peer's neighbors
//	@Description	Get the list of neighbors for all peers
//	@Tags			Peers
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	NeighborsResponse
//	@Failure		404	{object}	APIError
//	@Failure		500	{object}	APIError
//	@Router			/peers/neighbors [get]
func (h *Handler) GetPeersNeighbors(c echo.Context) error {
	neighbors, err := h.Repo.GetPeersNeighbors(c.Request().Context())
	if err != nil {
		return ServerError(err)
	}
	return c.JSON(http.StatusOK, NeighborsResponse{neighbors})
}

// GetPeerNeighbors returns the neighbors of a peer.
//
//	@Summary		Get peer neighbors
//	@Description	Get the list of neighbors for a specific peer
//	@Tags			Peers
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Peer ID (e.g. 12...)"
//	@Success		200	{object}	NeighborResponse
//	@Failure		404	{object}	APIError
//	@Failure		500	{object}	APIError
//	@Router			/peers/{id}/neighbors [get]
func (h *Handler) GetPeerNeighbors(c echo.Context) error {
	id := c.Param("id")
	neighbor, err := h.Repo.GetPeerNeighbors(c.Request().Context(), id)
	if err != nil {
		return ServerError(err)
	}
	return c.JSON(http.StatusOK, NeighborResponse{neighbor})
}

// GetAnalytics returns overall analytics data.
//
//	@Summary		Get overall analytics
//	@Description	Get comprehensive analytics data about the network
//	@Tags			Analytics
//	@Produce		json
//	@Success		200	{object}	AnalyticsResponse
//	@Failure		500	{object}	APIError
//	@Router			/analytics [get]
func (h *Handler) GetAnalytics(c echo.Context) error {
	var (
		analytics AnalyticsResponse
		err       error
	)

	var cacheKey uint64 = 0
	if h.AnalyticsResponseCache != nil {
		cached, ok := h.AnalyticsResponseCache.Get(cacheKey)
		if ok {
			return c.JSON(http.StatusOK, cached)
		}
	}

	ctx := c.Request().Context()

	// Always use optimized methods (pass CCV from reference node if available)
	ccv := h.GetFilterCCV()

	// Get latest peer count using optimized method
	analytics.PeersLatest, err = h.Repo.GetLatestCrawlTotalCountOptimized(ctx, ccv)
	if err != nil {
		return ServerError(err)
	}

	// Get sync status counts using optimized method
	analytics.SyncStatusCount, err = h.Repo.GetSyncStatusCountOptimized(ctx, ccv)
	if err != nil {
		return ServerError(err)
	}

	if h.AnalyticsResponseCache != nil {
		h.AnalyticsResponseCache.Add(cacheKey, analytics)
	}

	return c.JSON(http.StatusOK, analytics)
}

// GetPeerCount returns the total number of peers that we have seen in the network.
//
//	@Summary		Get total peer count and peers in the last crawl
//	@Description	Get the total number of peers in the network and the number of peers in the last crawl
//	@Tags			Analytics
//	@Produce		json
//	@Success		200	{object}	PeerCountResponse
//	@Failure		500	{object}	APIError
//	@Router			/analytics/peers [get]
func (h *Handler) GetPeerCount(c echo.Context) error {
	count, err := h.Repo.GetTotalPeerCount(c.Request().Context())
	if err != nil {
		return ServerError(err)
	}

	// Use the stored total count from the latest crawl sync status
	peersLatest, err := h.Repo.GetLatestCrawlTotalCount(c.Request().Context())
	if err != nil {
		return ServerError(err)
	}

	peersChurnIn, err := h.Repo.GetChurnInPeerCount(c.Request().Context())
	if err != nil {
		return ServerError(err)
	}

	peersChurnOut, err := h.Repo.GetChurnOutPeerCount(c.Request().Context())
	if err != nil {
		return ServerError(err)
	}

	return c.JSON(http.StatusOK, PeerCountResponse{
		TotalPeers:    count,
		PeersLatest:   peersLatest,
		PeersChurnIn:  peersChurnIn,
		PeersChurnOut: peersChurnOut,
	})
}

// GetPeerCountByAgent returns the peer count grouped by agent and version.
//
//	@Summary		Get peer count by client names and versions
//	@Description	Get the number of peers grouped by client names and versions
//	@Tags			Analytics
//	@Produce		json
//	@Success		200	{object}	AgentCountResponse
//	@Failure		500	{object}	APIError
//	@Router			/analytics/clients [get]
func (h *Handler) GetPeerCountByAgent(c echo.Context) error {
	var counts []*repo.AgentCount
	var err error

	// Always use optimized method (pass CCV from reference node if available)
	counts, err = h.Repo.GetPeerCountByAgentOptimized(c.Request().Context(), h.GetFilterCCV())
	if err != nil {
		return ServerError(err)
	}

	return c.JSON(http.StatusOK, AgentCountResponse{AgentCounts: ProcessAgentCounts(counts)})
}

// GetPeerCountByProtocol returns the peer count grouped by protocol.
//
//	@Summary		Get peer count by protocol
//	@Description	Get the number of peers grouped by protocol
//	@Tags			Analytics
//	@Produce		json
//	@Success		200	{object}	ProtocolCountResponse
//	@Failure		500	{object}	APIError
//	@Router			/analytics/protocols [get]
func (h *Handler) GetPeerCountByProtocol(c echo.Context) error {
	counts, err := h.Repo.GetPeerCountByProtocol(c.Request().Context())
	if err != nil {
		return ServerError(err)
	}
	return c.JSON(http.StatusOK, ProtocolCountResponse{counts})
}

// GetPeerCountByContinent returns the peer count grouped by continent.
//
//	@Summary		Get peer count by continent
//	@Description	Get the number of peers grouped by continent
//	@Tags			Analytics
//	@Produce		json
//	@Success		200	{object}	ContinentCountResponse
//	@Failure		500	{object}	APIError
//	@Router			/analytics/continents [get]
func (h *Handler) GetPeerCountByContinent(c echo.Context) error {
	var cacheKey uint64 = 0
	if h.ContinentsResponseCache != nil {
		cached, ok := h.ContinentsResponseCache.Get(cacheKey)
		if ok {
			return c.JSON(http.StatusOK, cached)
		}
	}

	var counts []*repo.ContinentCount
	var err error

	// Always use optimized method (pass CCV from reference node if available)
	counts, err = h.Repo.GetPeerCountByContinentOptimized(c.Request().Context(), h.GetFilterCCV())
	if err != nil {
		return ServerError(err)
	}

	response := ContinentCountResponse{counts}

	if h.ContinentsResponseCache != nil {
		h.ContinentsResponseCache.Add(cacheKey, response)
	}

	return c.JSON(http.StatusOK, response)
}

// GetPeerCountByCountry returns the peer count grouped by country.
//
//	@Summary		Get peer count by country
//	@Description	Get the number of peers grouped by country
//	@Tags			Analytics
//	@Produce		json
//	@Success		200	{object}	CountryCountResponse
//	@Failure		500	{object}	APIError
//	@Router			/analytics/countries [get]
func (h *Handler) GetPeerCountByCountry(c echo.Context) error {
	var cacheKey uint64 = 0
	if h.CountriesResponseCache != nil {
		cached, ok := h.CountriesResponseCache.Get(cacheKey)
		if ok {
			return c.JSON(http.StatusOK, cached)
		}
	}

	var counts []*repo.CountryCount
	var err error

	// Always use optimized method (pass CCV from reference node if available)
	counts, err = h.Repo.GetPeerCountByCountryOptimized(c.Request().Context(), h.GetFilterCCV())
	if err != nil {
		return ServerError(err)
	}

	response := CountryCountResponse{counts}

	if h.CountriesResponseCache != nil {
		h.CountriesResponseCache.Add(cacheKey, response)
	}

	return c.JSON(http.StatusOK, response)
}

// GetPeerCountByCity returns the peer count grouped by city.
//
//	@Summary		Get peer count by city
//	@Description	Get the number of peers grouped by city
//	@Tags			Analytics
//	@Produce		json
//	@Success		200	{object}	CityCountResponse
//	@Failure		500	{object}	APIError
//	@Router			/analytics/cities [get]
func (h *Handler) GetPeerCountByCity(c echo.Context) error {
	var cacheKey uint64 = 0
	if h.CitiesResponseCache != nil {
		cached, ok := h.CitiesResponseCache.Get(cacheKey)
		if ok {
			return c.JSON(http.StatusOK, cached)
		}
	}

	counts, err := h.Repo.GetPeerCountByCity(c.Request().Context())
	if err != nil {
		return ServerError(err)
	}

	response := CityCountResponse{counts}

	if h.CitiesResponseCache != nil {
		h.CitiesResponseCache.Add(cacheKey, response)
	}

	return c.JSON(http.StatusOK, response)
}

// GetPeerCountByASO returns the peer count grouped by Autonomous System Organization.
//
//	@Summary		Get peer count by ASO
//	@Description	Get the number of peers grouped by Autonomous System Organization
//	@Tags			Analytics
//	@Produce		json
//	@Success		200	{object}	ASCountResponse
//	@Failure		500	{object}	APIError
//	@Router			/analytics/networks [get]
func (h *Handler) GetPeerCountByASO(c echo.Context) error {
	counts, err := h.Repo.GetPeerCountByASO(c.Request().Context())
	if err != nil {
		return ServerError(err)
	}
	return c.JSON(http.StatusOK, ASCountResponse{counts})
}

// GetPeerHistory returns the peer count history.
//
//	@Summary		Get peer count history
//	@Description	Get the history of peer counts over time (daily)
//	@Tags			Analytics
//	@Produce		json
//	@Param			start	query	string	false	"Start date (RFC3339 format)"
//	@Param			end		query	string	false	"End date (RFC3339 format)"
//	@Success		200	{object}	PeerHistoryResponse
//	@Failure		400	{object}	APIError
//	@Failure		500	{object}	APIError
//	@Router			/analytics/peers/history [get]
func (h *Handler) GetPeerHistory(c echo.Context) error {
	start, err := time.Parse(time.RFC3339, c.QueryParam("start"))
	if err != nil && c.QueryParam("start") != "" {
		return InvalidParameters(err)
	}

	end, err := time.Parse(time.RFC3339, c.QueryParam("end"))
	if err != nil && c.QueryParam("end") != "" {
		return InvalidParameters(err)
	}

	// Use cached daily peer history from daily_peer_history table
	history, err := h.Repo.GetDailyPeerHistory(c.Request().Context(), start, end)
	if err != nil {
		return ServerError(err)
	}

	return c.JSON(http.StatusOK, PeerHistoryResponse{History: history})
}

// Response types

type PeersResponse struct {
	Peers               []*repo.PeerInfo `json:"peers"`
	NextPaginationToken string           `json:"next_pagination_token"`
	TotalPeers          int64            `json:"total_peers"`
}

type PeerResponse struct {
	*repo.PeerInfo
}

type NeighborsResponse struct {
	Neighbors []*repo.PeerNeighbors `json:"neighbors"`
}

type NeighborResponse struct {
	*repo.PeerNeighbors
}

type AnalyticsResponse struct {
	// Total number of peers that were seen in the network.
	TotalPeers int64 `json:"total_peers"`

	// Number of peers that were seen in the latest crawl.
	PeersLatest int64 `json:"peers_latest"`

	// Number of new peers that were seen in the latest crawl but were not seen in the previous crawls.
	PeersChurnIn int64 `json:"peers_churn_in"`

	// Number of peers that were previously seen but are no longer seen in the latest crawl.
	PeersChurnOut int64 `json:"peers_churn_out"`

	PeersByAgent     []*AgentVersionCounts  `json:"clients"`
	PeersByProtocol  []*repo.ProtocolCount  `json:"protocols"`
	PeersByContinent []*repo.ContinentCount `json:"continents"`
	PeersByCountry   []*repo.CountryCount   `json:"countries"`
	PeersByCity      []*repo.CityCount      `json:"cities"`
	PeersByASO       []*repo.ASCount        `json:"networks"`
	SyncStatusCount  map[string]int         `json:"sync_status"`
}

type PeerCountResponse struct {
	TotalPeers    int64 `json:"total_peers"`
	PeersLatest   int64 `json:"peers_latest"`
	PeersChurnIn  int64 `json:"peers_churn_in"`
	PeersChurnOut int64 `json:"peers_churn_out"`
}

type AgentCountResponse struct {
	AgentCounts []*AgentVersionCounts `json:"agents"`
}

type ProtocolCountResponse struct {
	ProtocolCounts []*repo.ProtocolCount `json:"protocols"`
}

type ContinentCountResponse struct {
	ContinentCounts []*repo.ContinentCount `json:"continents"`
}

type CountryCountResponse struct {
	CountryCounts []*repo.CountryCount `json:"countries"`
}

type CityCountResponse struct {
	CityCounts []*repo.CityCount `json:"cities"`
}

type ASCountResponse struct {
	ASCounts []*repo.ASCount `json:"networks"`
}

type PeerHistoryResponse struct {
	History []*repo.PeerHistoryPoint `json:"history"`
}
