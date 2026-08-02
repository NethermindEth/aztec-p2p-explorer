//nolint:misspell
package types

import (
	"time"
)

type CrawlStatus string

const (
	CrawlStatusRunning  CrawlStatus = "running"
	CrawlStatusFinished CrawlStatus = "finished"
	CrawlStatusFailed   CrawlStatus = "failed"
)

type Crawl struct {
	ID              int       `json:"-"` // Database ID, set after insert
	State           string    `json:"state"`
	StartedAt       time.Time `json:"started_at"`
	FinishedAt      time.Time `json:"finished_at"`
	CrawledPeers    int       `json:"crawled_peers"`
	DialablePeers   int       `json:"dialable_peers"`
	UndialablePeers int       `json:"undialable_peers"`
	RemainingPeers  int       `json:"remaining_peers"`
	Version         string    `json:"version"`
	ReferenceCCV    string    `json:"reference_ccv,omitempty"`
}

type Visit struct {
	ConnectDuration string
	CrawlDuration   string
	VisitStartedAt  time.Time
	VisitEndedAt    time.Time
	ConnectErrorStr string
	CrawlErrorStr   string
}

type Peer struct {
	PeerID       string
	AgentVersion string
	Protocols    []string
	MultiAddrs   []*MultiAddr
}

type Neighbors struct {
	PeerID    string   `json:"PeerID"`
	Neighbors []string `json:"NeighborIDs"`
}

type MultiAddr struct {
	Address string    `json:"maddr"`
	IPList  []*IPInfo `json:"ip_info"`
}

type IPInfo struct {
	IPAddress string `json:"ip_address"`
	// Refers to the port number of the peer when connected via p2p.
	// Not the HTTP port or any other port.
	Port     int `json:"port"`
	*GeoInfo `json:",inline"`
}

type GeoInfo struct {
	ASOrganization string  `json:"as_name"`
	ASNumber       uint    `json:"as_number"`
	City           string  `json:"city_name"`
	Country        string  `json:"country_name"`
	CountryISO     string  `json:"country_iso"`
	Continent      string  `json:"continent_name"`
	ContinentCode  string  `json:"continent_code"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
}

type PeerChainInfo struct {
	BlockHeight         uint64
	BlockHeightObtained bool

	SpecVersion         string
	SpecVersionObtained bool

	SyncStatus         bool
	SyncStatusObtained bool
}
