package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/NethermindEth/aztec-p2p-explorer/core/utils"
)

var pollErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "aztec_p2p_explorer_aztec_feeder_poll_errors_total",
	Help: "Number of failed updateL2Tips polls, labelled by error class.",
}, []string{"reason"})

const (
	pollReasonHTTP           = "http"
	pollReasonRPCError       = "rpc_error"
	pollReasonDecode         = "decode"
	pollReasonEmptyResult    = "empty_result"
	pollReasonSchemaMismatch = "schema_mismatch"
)

func failPoll(reason string, err error) error {
	pollErrorsTotal.WithLabelValues(reason).Inc()
	return err
}

const (
	methodGetChainTips = "node_getChainTips" // v5+
	methodGetL2Tips    = "node_getL2Tips"    // v4 and earlier

	rpcCodeMethodNotFound = -32601
)

// Probe order: newest first.
var tipsMethods = []string{methodGetChainTips, methodGetL2Tips}

// rpcError keeps the JSON-RPC error code so the poller can branch on method-not-found.
type rpcError struct {
	Code    int
	Message string
}

func (e *rpcError) Error() string {
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

func isMethodNotFound(err error) bool {
	var rpcErr *rpcError
	return errors.As(err, &rpcErr) && rpcErr.Code == rpcCodeMethodNotFound
}

// AztecNodeStatus represents the status response from Nebula's Aztec node crawl
type AztecNodeStatus struct {
	CompressedComponentsVersion string `json:"compressedComponentsVersion"`
	LatestBlockNumber           uint64 `json:"latestBlockNumber"`
	LatestBlockHash             string `json:"latestBlockHash"`
}

// L2Tips is the chain-tips RPC response (node_getChainTips on v5+,
// node_getL2Tips earlier — same shape).
type L2Tips struct {
	Proposed     BlockInfo `json:"proposed"`
	Proven       NestedTip `json:"proven"`
	Finalised    NestedTip `json:"finalized"` //nolint:misspell
	Checkpointed NestedTip `json:"checkpointed"`
}

// NestedTip wraps a block and its checkpoint, as returned by the RPC for
// proven/finalised/checkpointed entries.
type NestedTip struct {
	Block      BlockInfo `json:"block"`
	Checkpoint BlockInfo `json:"checkpoint"`
}

// BlockInfo contains block number and hash information
type BlockInfo struct {
	Number uint64 `json:"number"`
	Hash   string `json:"hash"`
}

// NodeInfo represents the response from node_getNodeInfo RPC method
type NodeInfo struct {
	NodeVersion               string      `json:"nodeVersion"`
	L1ChainID                 int         `json:"l1ChainId"`
	ProtocolVersion           int         `json:"protocolVersion"`
	ENR                       string      `json:"enr"`
	L1ContractAddresses       interface{} `json:"l1ContractAddresses"`
	ProtocolContractAddresses interface{} `json:"protocolContractAddresses"`
}

// AztecFeederClient is a client for fetching chain information from Aztec's RPC
type AztecFeederClient struct {
	rpcURL            string
	httpClient        *http.Client
	latestTips        *L2Tips
	latestTipsMu      sync.RWMutex
	lastTipsSuccessAt time.Time // protected by latestTipsMu
	referenceENRInfo  *utils.ENRInfo
	referenceENRMu    sync.RWMutex
	tipsMethod        string // pinned tips RPC method; poll goroutine only
	stopChan          chan struct{}
	pollInterval      time.Duration
	logger            *slog.Logger
	latestHistory     map[uint64]string // block number -> hash for latest blocks
	finalizedHistory  map[uint64]string // block number -> hash for finalised blocks
	provenHistory     map[uint64]string // block number -> hash for proven blocks
	historyMu         sync.RWMutex
	maxHistorySize    int
}

// NewAztecFeederClient creates a new Aztec feeder client
func NewAztecFeederClient(logger *slog.Logger, rpcURL string, options ...func(*AztecFeederClient)) *AztecFeederClient {
	const defaultHistorySize = 20 // Default to storing 20 historical tips

	client := &AztecFeederClient{
		rpcURL:         rpcURL,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
		pollInterval:   5 * time.Second,
		logger:         logger,
		stopChan:       make(chan struct{}),
		maxHistorySize: defaultHistorySize,
	}

	// Initialise history maps
	client.latestHistory = make(map[uint64]string)
	client.finalizedHistory = make(map[uint64]string)
	client.provenHistory = make(map[uint64]string)

	for _, option := range options {
		option(client)
	}

	return client
}

// WithHistorySize configures the size of the tips history buffer
func WithHistorySize(size int) func(*AztecFeederClient) {
	return func(client *AztecFeederClient) {
		client.maxHistorySize = size
	}
}

// Start begins polling for the latest L2 tips and node info
func (r *AztecFeederClient) Start(ctx context.Context) error {
	// Populate the latest tips before starting the poller
	err := r.updateL2Tips(ctx)
	if err != nil {
		return err
	}

	// Populate the node info (including ENR) before starting the poller
	err = r.updateNodeInfo(ctx)
	if err != nil {
		r.logger.Warn("Failed to update node info on start", "err", err)
		// Don't fail startup if node info fetch fails
	}

	go r.pollL2Tips(ctx)
	return nil
}

// updateL2Tips fetches the latest tips, counting RPC errors once per update
// so a probe miss answered by a fallback is not a failed poll.
func (r *AztecFeederClient) updateL2Tips(ctx context.Context) error {
	err := r.fetchTipsPinnedOrProbe(ctx)
	if err == nil {
		return nil
	}
	var rpcErr *rpcError
	if errors.As(err, &rpcErr) {
		return failPoll(pollReasonRPCError, err)
	}
	// non-RPC failures were already counted with a more specific reason
	return err
}

// fetchTipsPinnedOrProbe uses the pinned method, or probes tipsMethods and
// pins the winner; method-not-found clears the pin, other errors abort.
func (r *AztecFeederClient) fetchTipsPinnedOrProbe(ctx context.Context) error {
	if r.tipsMethod != "" {
		err := r.fetchL2Tips(ctx, r.tipsMethod)
		if err == nil || !isMethodNotFound(err) {
			return err
		}
		r.logger.Warn("Tips RPC method no longer supported by node, re-probing",
			"method", r.tipsMethod)
		r.tipsMethod = ""
	}

	var lastErr error
	for _, method := range tipsMethods {
		err := r.fetchL2Tips(ctx, method)
		if err == nil {
			r.tipsMethod = method
			r.logger.Info("Pinned tips RPC method", "method", method)
			return nil
		}
		if !isMethodNotFound(err) {
			return err
		}
		lastErr = err
	}
	return lastErr
}

const jsonRPCVersionField = "jsonrpc"

func rpcRequestBody(method string) map[string]interface{} {
	return map[string]interface{}{
		jsonRPCVersionField: "2.0",
		"method":            method,
		"params":            []interface{}{},
		"id":                1,
	}
}

// fetchL2Tips performs a single tips request with the given RPC method.
func (r *AztecFeederClient) fetchL2Tips(ctx context.Context, method string) error {
	requestBody := rpcRequestBody(method)

	// Marshal / request-construction errors are programmer or config bugs,
	// not runtime poll failures — don't route them through failPoll.
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.rpcURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return failPoll(pollReasonHTTP, fmt.Errorf("failed to send request: %w", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return failPoll(pollReasonHTTP, fmt.Errorf("failed to read response body: %w", err))
	}

	var rpcResponse struct {
		JSONRPC string  `json:"jsonrpc"`
		ID      int     `json:"id"`
		Result  *L2Tips `json:"result"`
		Error   *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &rpcResponse); err != nil {
		return failPoll(pollReasonDecode, fmt.Errorf("failed to unmarshal response: %w", err))
	}

	if rpcResponse.Error != nil {
		return &rpcError{Code: rpcResponse.Error.Code, Message: rpcResponse.Error.Message}
	}

	if rpcResponse.Result == nil {
		return failPoll(pollReasonEmptyResult, fmt.Errorf("empty result from RPC"))
	}

	// Guard against silent schema drift: the RPC responded, but our struct
	// found no keys to populate. This is what happens if Aztec renames a
	// field (e.g. `latest` → `proposed`, as in the 4.x release). Without
	// this check the poller would keep running against a zero-valued tip,
	// and every peer would be classified as "syncing".
	if rpcResponse.Result.Proposed.Number == 0 && rpcResponse.Result.Proposed.Hash == "" {
		return failPoll(pollReasonSchemaMismatch, fmt.Errorf(
			"RPC returned tips with empty `proposed` — likely a schema mismatch, "+
				"check the Aztec %s response shape", method))
	}

	r.latestTipsMu.Lock()
	r.latestTips = rpcResponse.Result
	r.lastTipsSuccessAt = time.Now()
	r.latestTipsMu.Unlock()

	// Store tips in history for hash validation
	r.storeTipsInHistory(*rpcResponse.Result)

	r.logger.Debug("Updated L2 tips",
		"method", method,
		"proposed", r.latestTips.Proposed.Number,
		"proven", r.latestTips.Proven.Block.Number,
		"finalised", r.latestTips.Finalised.Block.Number,
		"checkpointed", r.latestTips.Checkpointed.Block.Number)

	return nil
}

// storeTipsInHistory stores the current tips in the history maps
func (r *AztecFeederClient) storeTipsInHistory(tips L2Tips) {
	r.historyMu.Lock()
	defer r.historyMu.Unlock()

	// Store in separate maps
	r.latestHistory[tips.Proposed.Number] = tips.Proposed.Hash
	r.finalizedHistory[tips.Finalised.Block.Number] = tips.Finalised.Block.Hash
	r.provenHistory[tips.Proven.Block.Number] = tips.Proven.Block.Hash

	// Clean up old entries if we exceed max history size
	r.cleanupOldHistory()

	r.logger.Debug("Stored tips in history",
		"proposed", tips.Proposed.Number,
		"proven", tips.Proven.Block.Number,
		"finalised", tips.Finalised.Block.Number)
}

// cleanupOldHistory removes old entries from history maps to maintain max size
func (r *AztecFeederClient) cleanupOldHistory() {
	// Helper function to clean a specific history map
	cleanMap := func(m map[uint64]string) {
		if len(m) <= r.maxHistorySize {
			return
		}

		// Find oldest entries to remove
		var blockNumbers []uint64
		for blockNum := range m {
			blockNumbers = append(blockNumbers, blockNum)
		}

		// Sort block numbers
		for i := 0; i < len(blockNumbers)-1; i++ {
			for j := i + 1; j < len(blockNumbers); j++ {
				if blockNumbers[i] > blockNumbers[j] {
					blockNumbers[i], blockNumbers[j] = blockNumbers[j], blockNumbers[i]
				}
			}
		}

		// Remove oldest entries
		toRemove := len(m) - r.maxHistorySize
		for i := 0; i < toRemove && i < len(blockNumbers); i++ {
			delete(m, blockNumbers[i])
		}
	}

	cleanMap(r.latestHistory)
	cleanMap(r.finalizedHistory)
	cleanMap(r.provenHistory)
}

// findBlockHashInHistory searches for a specific block number in the tips history
// and returns the hash if found, empty string if not found
func (r *AztecFeederClient) findLatestBlockHashInHistory(blockNumber uint64) string {
	r.historyMu.RLock()
	defer r.historyMu.RUnlock()

	// Check in all history maps
	if hash, exists := r.latestHistory[blockNumber]; exists {
		return hash
	}

	return ""
}

// pollL2Tips continuously polls for the latest L2 tips and node info
func (r *AztecFeederClient) pollL2Tips(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := r.updateL2Tips(ctx)
			if err != nil {
				r.logger.Error("Failed to update L2 tips", "err", err)
			}

			// Also update node info to get latest ENR/CCV
			err = r.updateNodeInfo(ctx)
			if err != nil {
				r.logger.Error("Failed to update node info", "err", err)
			}

			timer := time.NewTimer(r.pollInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				// Continue to the next iteration
			}
		}
	}
}

// GetLatestHeight returns the latest block height (implements Feeder interface)
func (r *AztecFeederClient) GetLatestHeight() uint64 {
	r.latestTipsMu.RLock()
	defer r.latestTipsMu.RUnlock()

	if r.latestTips == nil {
		return 0
	}

	r.logger.Debug("Latest Aztec height", "height", r.latestTips.Proposed.Number)
	return r.latestTips.Proposed.Number
}

// GetL2Tips returns the current L2 tips
func (r *AztecFeederClient) GetL2Tips() *L2Tips {
	r.latestTipsMu.RLock()
	defer r.latestTipsMu.RUnlock()

	if r.latestTips == nil {
		return nil
	}

	// Return a copy to avoid data races
	return &L2Tips{
		Proposed:     r.latestTips.Proposed,
		Proven:       r.latestTips.Proven,
		Finalised:    r.latestTips.Finalised,
		Checkpointed: r.latestTips.Checkpointed,
	}
}

// GetTipsAge returns the time elapsed since the last successful updateL2Tips,
// or -1 if updateL2Tips has never succeeded.
func (r *AztecFeederClient) GetTipsAge() time.Duration {
	r.latestTipsMu.RLock()
	defer r.latestTipsMu.RUnlock()

	if r.lastTipsSuccessAt.IsZero() {
		return time.Duration(-1)
	}
	return time.Since(r.lastTipsSuccessAt)
}

// SyncReason names the branch ValidateNodeSync took. Used as a stable
// Prometheus label value.
type SyncReason string

const (
	ReasonInWindow         SyncReason = "in_window"
	ReasonBehind           SyncReason = "behind"
	ReasonAheadOfReference SyncReason = "ahead_of_reference"
	ReasonHashNotInHistory SyncReason = "hash_not_in_history"
	ReasonNoReferenceTips  SyncReason = "no_reference_tips"
)

// ValidateNodeSync validates if a node is synced based on its status and the
// reference L2 tips. Returns:
//   - synced:   whether the node is synced (only meaningful when obtained is true)
//   - obtained: whether we could make a determination at all. False means we
//     genuinely don't know (no reference tips, or the peer is ahead of our
//     reference by more than maxBlockDiff, implying our poller is stale).
//   - reason:   the branch that produced the verdict.
//   - details:  human-readable explanation for logs.
func (r *AztecFeederClient) ValidateNodeSync(
	reference *L2Tips, status *AztecNodeStatus, maxBlockDiff uint64,
) (synced, obtained bool, reason SyncReason, details string) {
	tips := reference
	if tips == nil {
		return false, false, ReasonNoReferenceTips, "no reference L2 tips available"
	}

	// Reference and node both at height 0 means the chain hasn't produced blocks yet,
	// so treat the node as synced to avoid false negatives.
	if tips.Proposed.Number == 0 && status.LatestBlockNumber == 0 {
		return true, true, ReasonInWindow, "node is synced"
	}

	// Peer claims to be more than maxBlockDiff blocks ahead of our reference.
	// Our reference is likely stale (poller stuck or behind the network),
	// so we can't judge — mark as unknown rather than falsely reporting synced.
	if status.LatestBlockNumber > tips.Proposed.Number+maxBlockDiff {
		return false, false, ReasonAheadOfReference, fmt.Sprintf(
			"peer is ahead of reference: node=%d, reference=%d — reference may be stale",
			status.LatestBlockNumber, tips.Proposed.Number)
	}

	// Underflow-safe block-number check: if reference is still below
	// maxBlockDiff, any non-zero peer height is acceptable.
	var minAcceptable uint64
	if tips.Proposed.Number > maxBlockDiff {
		minAcceptable = tips.Proposed.Number - maxBlockDiff
	}

	if status.LatestBlockNumber < minAcceptable {
		return false, true, ReasonBehind, fmt.Sprintf(
			"latest block number too far behind: node=%d, reference=%d",
			status.LatestBlockNumber, tips.Proposed.Number)
	}

	historicalHash := r.findLatestBlockHashInHistory(status.LatestBlockNumber)

	hashOk := historicalHash != "" && historicalHash == status.LatestBlockHash

	if !hashOk {
		return false, true, ReasonHashNotInHistory, fmt.Sprintf(
			"latest block hash not found in history at height %d: node=%s, reference=%s",
			status.LatestBlockNumber, status.LatestBlockHash, tips.Proposed.Hash)
	}

	return true, true, ReasonInWindow, "node is synced"
}

// updateNodeInfo fetches node info including ENR from the Aztec RPC
func (r *AztecFeederClient) updateNodeInfo(ctx context.Context) error {
	requestBody := rpcRequestBody("node_getNodeInfo")

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.rpcURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var rpcResponse struct {
		JSONRPC string    `json:"jsonrpc"`
		ID      int       `json:"id"`
		Result  *NodeInfo `json:"result"`
		Error   *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &rpcResponse); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if rpcResponse.Error != nil {
		return fmt.Errorf("RPC error %d: %s", rpcResponse.Error.Code, rpcResponse.Error.Message)
	}

	if rpcResponse.Result == nil {
		return fmt.Errorf("empty result from RPC")
	}

	// Decode ENR to extract CCV
	if rpcResponse.Result.ENR != "" {
		enrInfo, err := utils.DecodeENR(rpcResponse.Result.ENR)
		if err != nil {
			r.logger.Error("Failed to decode reference node ENR", "err", err, "enr", rpcResponse.Result.ENR)
			return fmt.Errorf("failed to decode ENR: %w", err)
		}

		r.referenceENRMu.Lock()
		r.referenceENRInfo = enrInfo
		r.referenceENRMu.Unlock()
	}

	return nil
}

// GetReferenceCCV returns the Compressed Components Version (CCV) from the reference node's ENR
// Returns empty string if not available
func (r *AztecFeederClient) GetReferenceCCV() string {
	r.referenceENRMu.RLock()
	defer r.referenceENRMu.RUnlock()

	if r.referenceENRInfo == nil {
		return ""
	}

	return r.referenceENRInfo.CompressedComponentsVersion
}
