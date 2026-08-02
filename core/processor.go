//nolint:misspell
package core

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/core/utils"
	"github.com/NethermindEth/aztec-p2p-explorer/maxmind"
	"github.com/NethermindEth/aztec-p2p-explorer/repo"
)

const (
	visitSuffix          = "visits.ndjson"
	neighborSuffix       = "neighbors.ndjson"
	crawlSuffix          = "crawl.json"
	crawlPropSuffix      = "crawl_properties.json"
	defaultNebulaFileLen = 4

	bufferSize = 1024 * 1024
)

// outcome ∈ {"synced", "not_synced", "unknown"}; reason is a SyncReason value.
var peerSyncClassificationTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "aztec_p2p_explorer_peer_sync_classification_total",
	Help: "Per-peer sync classifications, labelled by peer_id, outcome and reason.",
}, []string{"peer_id", "outcome", "reason"})

func classifyOutcome(synced, obtained bool) string {
	if !obtained {
		return "unknown"
	}
	if synced {
		return "synced"
	}
	return "not_synced"
}

type CrawlProperties struct {
	AgentVersion map[string]int `json:"agent_version"`
	Protocol     map[string]int `json:"protocol"`
}

type VisitProperties struct {
	ENR                 string `json:"enr,omitempty"`
	AztecStatusResponse string `json:"aztec_status_response,omitempty"`
}

type RawVisit struct {
	PeerID       string           `json:"PeerID"`
	MultiAddrs   []string         `json:"Maddrs"`
	Protocols    []string         `json:"Protocols"`
	AgentVersion string           `json:"AgentVersion"`
	Properties   *VisitProperties `json:"Properties,omitempty"`
	types.Visit  `json:",inline"`
}

type Processor interface {
	Process(ctx context.Context) (*types.Crawl, error)
}

type NebulaJSONProcessor struct {
	logger          *slog.Logger
	repo            *repo.PeerRepository
	maxmindClient   maxmind.Client
	processExisting bool
	dataDir         string
	aztecFeeder     *AztecFeederClient
	maxBlockDiff    uint64
}

func NewNebulaJSONProcessor(
	logger *slog.Logger,
	peerRepo *repo.PeerRepository,
	maxmindClient maxmind.Client,
	aztecFeeder *AztecFeederClient,
	maxBlockDiff uint64,
	options ...func(*NebulaJSONProcessor),
) *NebulaJSONProcessor {
	processor := &NebulaJSONProcessor{
		logger:          logger,
		repo:            peerRepo,
		maxmindClient:   maxmindClient,
		aztecFeeder:     aztecFeeder,
		processExisting: true,
		dataDir:         types.NebulaDataDirPath,
		maxBlockDiff:    30, // Default max block difference
	}

	for _, option := range options {
		option(processor)
	}

	return processor
}

func WithProcessExisting(processExisting bool) func(*NebulaJSONProcessor) {
	return func(n *NebulaJSONProcessor) {
		n.processExisting = processExisting
	}
}

func WithDataDir(dataDir string) func(*NebulaJSONProcessor) {
	return func(n *NebulaJSONProcessor) {
		n.dataDir = dataDir
	}
}

func (n *NebulaJSONProcessor) Process(ctx context.Context) (*types.Crawl, error) {
	fileSets, err := n.getFilesToProcess()
	if err != nil {
		return nil, fmt.Errorf("failed to get files to process: %w", err)
	}

	if len(fileSets) == 0 {
		return nil, fmt.Errorf("no valid file sets found to process")
	}

	var lastCrawl *types.Crawl
	for _, fileSet := range fileSets {
		crawl, err := n.processFileSet(ctx, fileSet)
		if err != nil {
			return nil, fmt.Errorf("failed to process file set: %w", err)
		}
		lastCrawl = crawl

		// Delete processed files
		for _, file := range fileSet {
			if err := os.Remove(file); err != nil {
				n.logger.Error("Failed to delete processed file", "file", file, "error", err)
			}
		}
	}

	return lastCrawl, nil
}

func (n *NebulaJSONProcessor) processCrawl(ctx context.Context, fileName, referenceCCV string) (*types.Crawl, error) {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var crawl *types.Crawl
	if err := json.Unmarshal(file, &crawl); err != nil {
		return nil, fmt.Errorf("unmarshal crawl: %w", err)
	}

	crawl.ReferenceCCV = referenceCCV
	crawlID, err := n.repo.CreateCrawl(ctx, crawl, types.CrawlStatusRunning)
	if err != nil {
		return nil, err
	}

	crawl.ID = crawlID
	return crawl, nil
}

func (n *NebulaJSONProcessor) processCrawlProperties(ctx context.Context, fileName string) error {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var crawlProperties *CrawlProperties
	if err := json.Unmarshal(file, &crawlProperties); err != nil {
		return fmt.Errorf("unmarshal crawl properties: %w", err)
	}

	// Remove empty string keys
	delete(crawlProperties.AgentVersion, "")
	delete(crawlProperties.Protocol, "")

	err = n.repo.CreateAgentVersions(ctx, crawlProperties.AgentVersion)
	if err != nil {
		n.logger.Error("Failed to create agent versions", "error", err)
	}

	err = n.repo.CreateProtocols(ctx, crawlProperties.Protocol)
	if err != nil {
		n.logger.Error("Failed to create protocols", "error", err)
	}

	return nil
}

func (n *NebulaJSONProcessor) processVisits(ctx context.Context, fileName string, crawlID int, referenceCCV string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	var referenceL2Tips *L2Tips
	if n.aztecFeeder != nil {
		referenceL2Tips = n.aztecFeeder.GetL2Tips()
		history, _ := json.Marshal(n.aztecFeeder.latestHistory)
		n.logger.Info("Reference L2 tips for visits", "tips", referenceL2Tips)
		n.logger.Info("Latest block history", "history", string(history))
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, bufferSize)
	scanner.Buffer(buf, bufferSize)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}

		var visit RawVisit
		if err := json.Unmarshal([]byte(line), &visit); err != nil {
			return fmt.Errorf("unmarshal visit: %w", err)
		}

		n.logger.Debug("Processing visit", "visit", visit)
		if err := n.processVisit(ctx, &visit, crawlID, referenceL2Tips, referenceCCV); err != nil {
			n.logger.Error("Failed to process visit", "error", err, "peerID", visit.PeerID)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

const (
	AgentVersionUnknown = "Unknown"
	AgentVersionOther   = "Other"
)

// isValidVersion checks if the version string matches semver-like pattern
// \d+\.\d+\.\d+ optionally followed by a -prerelease and/or +buildmetadata suffix
// (e.g. "4.1.3", "4.1.0-rc.2", "4.2.0-nightly.20260408-1", "4.1.3+build.5").
var versionRegexp = regexp.MustCompile(`^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

func isValidVersion(version string) bool {
	return versionRegexp.MatchString(version)
}

func (n *NebulaJSONProcessor) getAgentVersion(rawVisit *RawVisit) string {
	// If ENR is available in Properties, try to extract client version from "ver" field

	if rawVisit.Properties == nil || rawVisit.Properties.ENR == "" {
		return AgentVersionUnknown
	}

	enrInfo, err := utils.DecodeENR(rawVisit.Properties.ENR)
	if err != nil {
		n.logger.Error("Failed to decode ENR info",
			"peer_id", rawVisit.PeerID,
			"enr", rawVisit.Properties.ENR,
			"error", err)
		return AgentVersionOther
	}

	if enrInfo.Version == "" {
		return AgentVersionUnknown
	}

	// If doesn't match '\d+.\d+.\d+' pattern, return "Other"
	if !isValidVersion(enrInfo.Version) {
		n.logger.Warn("Client version does not match expected format",
			"peer_id", rawVisit.PeerID,
			"version", enrInfo.Version)
		return AgentVersionOther
	}

	return enrInfo.Version
}

func (n *NebulaJSONProcessor) processVisit(
	ctx context.Context, rawVisit *RawVisit, crawlID int, referenceL2Tips *L2Tips, referenceCCV string,
) error {
	maddrs, err := utils.ConvertMultiAddrs(rawVisit.MultiAddrs)
	if err != nil {
		return fmt.Errorf("convert multiaddrs: %w", err)
	}

	// Extract ENR info for both agent version and CCV
	enrInfo := n.extractENRInfo(rawVisit)

	peer := &types.Peer{
		PeerID:       rawVisit.PeerID,
		AgentVersion: n.getAgentVersion(rawVisit),
		Protocols:    rawVisit.Protocols,
		MultiAddrs:   maddrs,
	}

	visit := &types.Visit{
		ConnectDuration: rawVisit.ConnectDuration,
		CrawlDuration:   rawVisit.CrawlDuration,
		VisitStartedAt:  rawVisit.VisitStartedAt,
		VisitEndedAt:    rawVisit.VisitEndedAt,
		ConnectErrorStr: rawVisit.ConnectErrorStr,
		CrawlErrorStr:   rawVisit.CrawlErrorStr,
	}

	if visit.ConnectErrorStr != "" || visit.CrawlErrorStr != "" {
		return nil
	}

	// Process IP addresses first to get geo data
	firstGeoInfo := n.processIPAddresses(ctx, peer.MultiAddrs)

	// Process sync status and get peer state
	sync := n.processSyncStatus(rawVisit, peer.PeerID, referenceL2Tips)

	// Get spec version (CCV) from ENR, status message, or fall back to sync status details
	specVersion := n.getSpecVersion(enrInfo, sync.CCV, sync.Details)

	// Filter peers by CCV - only store peers that match the reference node's CCV
	if !n.shouldStorePeer(peer.PeerID, specVersion, referenceCCV) {
		return nil
	}

	peerState := &types.PeerChainInfo{
		BlockHeight:         sync.BlockHeight,
		BlockHeightObtained: sync.BlockHeightObtained,
		SpecVersion:         specVersion,
		SpecVersionObtained: true,
		SyncStatus:          sync.Synced,
		SyncStatusObtained:  sync.Obtained,
	}

	// Now create everything including index in one transaction
	err = n.repo.CreatePeerVisitWithIndex(ctx, crawlID, visit, peer, peerState, firstGeoInfo)
	if err != nil {
		return fmt.Errorf("create peer visit with index: %w", err)
	}

	return nil
}

func (n *NebulaJSONProcessor) extractENRInfo(rawVisit *RawVisit) *utils.ENRInfo {
	if rawVisit.Properties != nil && rawVisit.Properties.ENR != "" {
		enrInfo, _ := utils.DecodeENR(rawVisit.Properties.ENR)
		return enrInfo
	}
	return nil
}

// syncStatusResult holds the outcome of decoding a peer's status blob and
// validating it against the reference tips. Synced/Obtained/Details map to
// the ValidateNodeSync return values; CCV and block height come from the
// peer's own status message.
type syncStatusResult struct {
	Synced              bool
	Obtained            bool
	Details             string
	CCV                 string
	BlockHeight         uint64
	BlockHeightObtained bool
}

func (n *NebulaJSONProcessor) processSyncStatus(
	rawVisit *RawVisit, peerID string, referenceL2Tips *L2Tips,
) syncStatusResult {
	res := syncStatusResult{Details: "no status response"}

	if rawVisit.Properties == nil || rawVisit.Properties.AztecStatusResponse == "" {
		return res
	}

	// Decode base64 encoded binary status message
	statusBytes, err := base64.RawStdEncoding.DecodeString(rawVisit.Properties.AztecStatusResponse)
	if err != nil {
		n.logger.Error("Failed to decode base64 status response", "peer_id", peerID, "error", err)
		res.Details = fmt.Sprintf("failed to decode base64: %v", err)
		return res
	}

	// Decode binary status message
	statusMsg, err := DecodeStatusMessage(statusBytes)
	if err != nil {
		n.logger.Error("Failed to decode binary status message", "peer_id", peerID, "error", err)
		res.Details = fmt.Sprintf("failed to decode status: %v", err)
		return res
	}

	// Store block height and CCV from status message
	res.BlockHeight = uint64(statusMsg.LatestBlockNumber)
	res.BlockHeightObtained = true
	res.CCV = statusMsg.CompressedComponentsVersion

	// Validate sync status
	if n.aztecFeeder != nil {
		aztecStatus := AztecNodeStatus{
			CompressedComponentsVersion: statusMsg.CompressedComponentsVersion,
			LatestBlockNumber:           uint64(statusMsg.LatestBlockNumber),
			LatestBlockHash:             statusMsg.LatestBlockHash,
		}
		var reason SyncReason
		res.Synced, res.Obtained, reason, res.Details = n.aztecFeeder.ValidateNodeSync(referenceL2Tips, &aztecStatus, n.maxBlockDiff)
		peerSyncClassificationTotal.WithLabelValues(peerID, classifyOutcome(res.Synced, res.Obtained), string(reason)).Inc()
		n.logger.Debug("Validated Aztec node sync status",
			"peer_id", peerID,
			"synced", res.Synced,
			"obtained", res.Obtained,
			"reason", reason,
			"details", res.Details,
			"latest_block", aztecStatus.LatestBlockNumber)
	} else {
		// Feeder isn't wired up — we can't judge the peer's sync state.
		// Leave Obtained = false so the DB stores NULL (unknown).
		// Skip the classification counter: missing feeder is a deployment
		// misconfiguration, not a per-peer outcome.
		res.Details = fmt.Sprintf("aztec feeder is absent, status: latest block %d, finalised %d",
			statusMsg.LatestBlockNumber, statusMsg.FinalisedBlockNumber)
	}

	return res
}

func (n *NebulaJSONProcessor) getSpecVersion(enrInfo *utils.ENRInfo, statusCCV, syncStatusDetails string) string {
	// Prefer CCV from ENR (most reliable)
	if enrInfo != nil && enrInfo.CompressedComponentsVersion != "" {
		return enrInfo.CompressedComponentsVersion
	}
	// Fall back to CCV from status message
	if statusCCV != "" {
		return statusCCV
	}
	// Last resort: sync status details (for backwards compatibility)
	return syncStatusDetails
}

func (n *NebulaJSONProcessor) shouldStorePeer(peerID, specVersion, referenceCCV string) bool {
	if referenceCCV == "" || specVersion == referenceCCV {
		return true
	}

	n.logger.Debug("Skipping peer with non-matching CCV",
		"peer_id", peerID,
		"peer_ccv", specVersion,
		"reference_ccv", referenceCCV)
	return false
}

func (n *NebulaJSONProcessor) processIPAddresses(ctx context.Context, maddrs []*types.MultiAddr) *types.GeoInfo {
	var firstGeoInfo *types.GeoInfo
	for _, maddr := range maddrs {
		for _, ipInfo := range maddr.IPList {
			geoInfo, err := n.maxmindClient.GetAddrInfo(ipInfo.IPAddress)
			if err != nil {
				n.logger.Error("Failed to get geo info", "error", err, "ip", ipInfo.IPAddress)
				continue
			}
			ipInfo.GeoInfo = geoInfo

			// Capture first geo info for index
			if firstGeoInfo == nil && geoInfo != nil {
				firstGeoInfo = geoInfo
			}

			err = n.repo.CreateIPAddress(ctx, ipInfo)
			if err != nil {
				n.logger.Error("Failed to create IP address", "error", err, "ip", ipInfo.IPAddress)
			}
		}
	}
	return firstGeoInfo
}

func (n *NebulaJSONProcessor) processFileSet(ctx context.Context, fileSet map[string]string) (*types.Crawl, error) {
	var referenceCCV string
	if n.aztecFeeder != nil {
		referenceCCV = n.aztecFeeder.GetReferenceCCV()
	}

	crawl, crawlErr := n.processCrawl(ctx, fileSet[crawlSuffix], referenceCCV)
	if crawlErr != nil {
		return nil, fmt.Errorf("failed to process crawl: %w", crawlErr)
	}
	n.logger.Info("Processed crawl", "crawlID", crawl.ID)

	crawlPropertiesErr := n.processCrawlProperties(ctx, fileSet[crawlPropSuffix])
	n.logger.Info("Processed crawl properties")

	if crawlPropertiesErr != nil {
		return nil, fmt.Errorf("failed to process crawl properties: %w", crawlPropertiesErr)
	}

	// Process visits
	if err := n.processVisits(ctx, fileSet[visitSuffix], crawl.ID, referenceCCV); err != nil {
		return nil, fmt.Errorf("failed to process visits: %w", err)
	}
	n.logger.Info("Processed visits")

	// Process neighbors
	n.logger.Info("Skipping processing neighbors")

	err := n.repo.UpdateCrawlStatus(ctx, crawl.ID, types.CrawlStatusFinished)
	if err != nil {
		return nil, fmt.Errorf("failed to update crawl status: %w", err)
	}
	n.logger.Info("Updated crawl status to finished", "crawlID", crawl.ID)

	// Calculate and store sync status counts for this crawl
	n.logger.Info("Calculating sync status counts", "crawlID", crawl.ID)
	if err := n.repo.CalculateAndStoreSyncStatusCount(ctx, crawl.ID); err != nil {
		return nil, fmt.Errorf("failed to calculate sync status counts: %w", err)
	} else {
		n.logger.Info("Successfully calculated and stored sync status counts", "crawlID", crawl.ID)
	}

	// Calculate and store peer history for this crawl
	n.logger.Info("Calculating peer history", "crawlID", crawl.ID)
	if err := n.repo.CalculateAndStorePeerHistory(ctx, crawl.ID); err != nil {
		n.logger.Error("Failed to calculate peer history", "error", err)
		// Don't fail the entire crawl processing if history calculation fails
	} else {
		n.logger.Info("Successfully calculated and stored peer history", "crawlID", crawl.ID)
	}

	err = n.repo.UpdateCrawlStatus(ctx, crawl.ID, types.CrawlStatusFinished)
	if err != nil {
		return nil, fmt.Errorf("failed to update crawl status: %w", err)
	}
	n.logger.Info("Updated crawl status to finished", "crawlID", crawl.ID)

	return crawl, nil
}

func (n *NebulaJSONProcessor) getFilesToProcess() ([]map[string]string, error) {
	files, err := filepath.Glob(filepath.Join(n.dataDir, "*"))
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	fileGroups := make(map[string]map[string]string)
	for _, file := range files {
		baseName := filepath.Base(file)
		parts := strings.SplitN(baseName, "_", 2)
		if len(parts) != 2 {
			continue
		}
		timestamp, suffix := parts[0], parts[1]

		// Validate timestamp format
		if _, err := time.Parse("2006-01-02T15:04", timestamp); err != nil {
			continue
		}

		if _, exists := fileGroups[timestamp]; !exists {
			fileGroups[timestamp] = make(map[string]string)
		}

		// Check if the file has one of the required suffixes
		switch suffix {
		case crawlSuffix, crawlPropSuffix, visitSuffix, neighborSuffix:
			fileGroups[timestamp][suffix] = file
		default: // this shouldn't happen, but just in case
			if err := os.Remove(file); err != nil {
				return nil, fmt.Errorf("failed to delete invalid file: %w", err)
			}
		}
	}

	var validFileSets []map[string]string
	for _, group := range fileGroups {
		if len(group) == defaultNebulaFileLen {
			validFileSets = append(validFileSets, group)
		} else { // we only want to process complete file sets
			for _, file := range group {
				if err := os.Remove(file); err != nil {
					return nil, fmt.Errorf("failed to delete incomplete file: %w", err)
				}
			}
		}
	}

	// Sort valid file sets by timestamp in descending order (most recent first)
	sort.Slice(validFileSets, func(i, j int) bool {
		timestampI := strings.SplitN(filepath.Base(validFileSets[i][crawlSuffix]), "_", 2)[0]
		timestampJ := strings.SplitN(filepath.Base(validFileSets[j][crawlSuffix]), "_", 2)[0]
		return timestampI > timestampJ
	})

	if !n.processExisting { // only process the most recent file set
		return validFileSets[:1], nil
	}

	return validFileSets, nil
}

// func (n *NebulaJSONProcessor) processNeighbors(ctx context.Context, fileName string) error {
// 	file, err := os.Open(fileName)
// 	if err != nil {
// 		return fmt.Errorf("open file: %w", err)
// 	}
// 	defer file.Close()

// 	scanner := bufio.NewScanner(file)
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		if strings.TrimSpace(line) == "" {
// 			continue // Skip empty lines
// 		}

// 		var neighbour types.Neighbours
// 		if err := json.Unmarshal([]byte(line), &neighbour); err != nil {
// 			return fmt.Errorf("unmarshal neighbour: %w", err)
// 		}

// 		if err := n.repo.CreateNeighbours(ctx, &neighbour); err != nil {
// 			return fmt.Errorf("create neighbour: %w", err)
// 		}
// 	}

// 	if err := scanner.Err(); err != nil {
// 		return fmt.Errorf("scanner error: %w", err)
// 	}

// 	return nil
// }
