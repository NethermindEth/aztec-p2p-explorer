package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mitchellh/hashstructure/v2"

	"github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/repo"
)

// Cluster represents a geographic cluster of peers
type Cluster struct {
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	Count     int      `json:"count"`
	Continent string   `json:"continent"`
	Country   string   `json:"country"`
	City      string   `json:"city"`
	PeerIDs   []string `json:"peer_ids"`
}

// PeersMapResponse represents the response for the peers/map endpoint
type PeersMapResponse struct {
	Clusters    []*Cluster `json:"clusters"`
	TotalPeers  int64      `json:"total_peers"`
	ActivePeers int64      `json:"active_peers"`
}

const useCache = false

// GetPeersMap returns a list of peer clusters for map visualisation
//
//	@Summary		Get peer clusters for map visualisation
//	@Description	Get geographical clusters of peers with their locations and counts
//	@Tags			Peers
//	@Accept			json
//	@Produce		json
//	@Param			continents		query		[]string	false	"Continent names"
//	@Param 			continents_code query 		[]string 	false 	"Continent codes"
//	@Param			countries		query		[]string	false	"Country names"
//	@Param 			countries_codes query 		[]string 	false 	"Country iso codes"
//	@Param			cities			query		[]string	false	"City names"
//	@Param			as_names		query		[]string	false	"Names of Autonomous System Organisations"
//	@Param			as_numbers		query		[]string	false	"Numbers of Autonomous System Organisations"
//	@Param			clients			query		[]string	false	"Client names (e.g. aztec-node)"
//	@Param			synced			query		bool		false	"Sync status"
//	@Param			latest			query		bool		false	"Include peers from latest crawl only (default: true)"
//	@Success		200	{object}	PeersMapResponse
//	@Failure		400	{object}	APIError
//	@Failure		500	{object}	APIError
//	@Router			/peers/map [get]
//
//nolint:gocyclo
func (h *Handler) GetPeersMap(c echo.Context) error {
	var filter types.PeerFilter
	if err := c.Bind(&filter); err != nil {
		return InvalidParameters(err)
	}

	// Default to latest crawl only for map visualisation
	latest := c.QueryParam("latest") != "false"

	// Apply CCV filter when latest=true and CCV is available from reference node
	if latest {
		if ccv := h.GetFilterCCV(); ccv != "" {
			filter.SpecVersion = ccv
		}
	}

	options := &repo.PeerQueryOptions{
		Filter: &filter,
		Latest: latest,
		// Use a large page size to get all peers
		PageSize: 50_000,
	}

	// Calculate cache key based on the options - using the same approach as the /peers endpoint
	// but storing in the dedicated PeersMapResponseCache since the response structure is different
	optionsHash, err := hashstructure.Hash(options, hashstructure.FormatV2, nil)
	if useCache && err == nil && h.PeersMapResponseCache != nil {
		if cachedResponse, ok := h.PeersMapResponseCache.Get(optionsHash); ok {
			// Use the cached response if available
			return c.JSON(http.StatusOK, cachedResponse)
		}
	}

	// Group peers by city name
	cityMap := make(map[string]*struct {
		PeerIDs   map[string]bool
		LatSum    float64
		LongSum   float64
		Count     int
		Continent string
		Country   string
		City      string
	})

	if h.UseIndexTable {
		// Use optimised index table for geographic data
		peerEntries, err := h.Repo.GetPeersForMap(c.Request().Context(), options)
		if err != nil {
			return ServerError(err)
		}

		for _, entry := range peerEntries {
			// Skip entries without valid geographic data (already filtered in query)
			if !entry.Latitude.Valid || !entry.Longitude.Valid || !entry.CityName.Valid {
				continue
			}

			// Create a key using city and country to distinguish cities with the same name
			cityKey := entry.CityName.String + "_" + entry.CountryName.String

			city, exists := cityMap[cityKey]
			if !exists {
				city = &struct {
					PeerIDs   map[string]bool
					LatSum    float64
					LongSum   float64
					Count     int
					Continent string
					Country   string
					City      string
				}{
					PeerIDs:   make(map[string]bool),
					Continent: entry.ContinentName.String,
					Country:   entry.CountryName.String,
					City:      entry.CityName.String,
				}
				cityMap[cityKey] = city
			}

			// Only count the peer once per city
			if !city.PeerIDs[entry.PeerMultiHash] {
				city.PeerIDs[entry.PeerMultiHash] = true
				city.Count++
				city.LatSum += entry.Latitude.Float64
				city.LongSum += entry.Longitude.Float64
			}
		}
	} else {
		// Use original implementation
		peers, _, _, err := h.Repo.GetPeers(c.Request().Context(), options)
		if err != nil {
			return ServerError(err)
		}

		for _, peer := range peers {
			for _, addr := range peer.MultiAddresses {
				for _, ip := range addr.IPList {
					if ip.GeoInfo == nil || ip.GeoInfo.Latitude == 0 && ip.GeoInfo.Longitude == 0 || ip.GeoInfo.City == "" {
						continue
					}

					// Create a key using city and country to distinguish cities with the same name
					cityKey := ip.GeoInfo.City + "_" + ip.GeoInfo.Country

					city, exists := cityMap[cityKey]
					if !exists {
						city = &struct {
							PeerIDs   map[string]bool
							LatSum    float64
							LongSum   float64
							Count     int
							Continent string
							Country   string
							City      string
						}{
							PeerIDs:   make(map[string]bool),
							Continent: ip.GeoInfo.Continent,
							Country:   ip.GeoInfo.Country,
							City:      ip.GeoInfo.City,
						}
						cityMap[cityKey] = city
					}

					// Only count the peer once per city
					if !city.PeerIDs[peer.PeerID] {
						city.PeerIDs[peer.PeerID] = true
						city.Count++
						city.LatSum += ip.GeoInfo.Latitude
						city.LongSum += ip.GeoInfo.Longitude
					}
				}
			}
		}
	}

	// Convert city map to clusters
	clusters := make([]*Cluster, 0, len(cityMap))
	for _, city := range cityMap {
		// Skip cities with no peers
		if city.Count == 0 {
			continue
		}

		// Calculate average lat/long for the city
		// We use the total number of IP entries for the average coordinates
		numCoordinates := float64(city.Count)

		// Convert peer IDs to slice
		peerIDs := make([]string, 0, len(city.PeerIDs))
		for id := range city.PeerIDs {
			peerIDs = append(peerIDs, id)
		}

		cluster := &Cluster{
			Latitude:  roundFloat(city.LatSum/numCoordinates, 1),
			Longitude: roundFloat(city.LongSum/numCoordinates, 1),
			Count:     city.Count,
			Continent: city.Continent,
			Country:   city.Country,
			City:      city.City,
			PeerIDs:   peerIDs,
		}

		clusters = append(clusters, cluster)
	}

	// Use the stored total count from the latest crawl sync status
	activePeers, err := h.Repo.GetLatestCrawlTotalCount(c.Request().Context())
	if err != nil {
		return ServerError(err)
	}

	response := PeersMapResponse{
		Clusters:    clusters,
		ActivePeers: activePeers,
	}

	// Cache the response
	if h.PeersMapResponseCache != nil {
		h.PeersMapResponseCache.Add(optionsHash, response)
	}

	return c.JSON(http.StatusOK, response)
}
