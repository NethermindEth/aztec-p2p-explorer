package core

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	//nolint:lll
	testValidENR    = "enr:-M64QNyulHqCeOHgo-lIqP_ukjtnbAM1DSMunMRlu8uWYlvuELTnEyQNovl3373XzuqPlxi02zTiSpCID8pWceg1Qq8HhWF6dGVjsTAwLTExMTU1MTExLWVlNmQ0ZTkzLTQxODkzMzcyMDctMmVmZDNmZDYtMjMzOWRhNDWCaWSCdjSCaXCEK4Mr4IlzZWNwMjU2azGhA0A6Sf7KJ1htR_CcPlWalV1WkwtzPipwBkI-jBzY5QWFg3RjcIKd0IN1ZHCCndCDdmVyhjAuODcuMg"
	testExpectedCCV = "00-11155111-ee6d4e93-4189337207-2efd3fd6-2339da45"
)

// createMockAztecServer creates a test HTTP server that simulates Aztec RPC responses
func createMockAztecServer(t *testing.T, responses map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
			ID     int    `json:"id"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
		}

		if result, exists := responses[req.Method]; exists {
			// Check if this is an error response
			if errorMap, isErrorMap := result.(map[string]interface{}); isErrorMap {
				if errorResponse, hasError := errorMap["error"]; hasError {
					response["error"] = errorResponse
				} else {
					response["result"] = result
				}
			} else {
				response["result"] = result
			}
		} else {
			response["error"] = map[string]interface{}{
				"code":    -32601,
				"message": "Method not found",
			}
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
}

func setupAztecFeederClientWithTips(t *testing.T) *AztecFeederClient {
	tips := &L2Tips{
		Proposed:  BlockInfo{Number: 100, Hash: "0xlatest100"},
		Finalised: NestedTip{Block: BlockInfo{Number: 95, Hash: "0xfinalised95"}},
	}

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": tips,
	})
	t.Cleanup(server.Close)

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx := context.Background()
	err := client.updateL2Tips(ctx)
	require.NoError(t, err)

	return client
}

func TestNewAztecFeederClient(t *testing.T) {
	logger := slog.Default()
	rpcURL := "http://test.example.com"

	client := NewAztecFeederClient(logger, rpcURL)

	assert.Equal(t, rpcURL, client.rpcURL)
	assert.Equal(t, 5*time.Second, client.pollInterval)
	assert.Equal(t, 20, client.maxHistorySize)
	assert.NotNil(t, client.latestHistory)
	assert.NotNil(t, client.finalizedHistory)
	assert.NotNil(t, client.provenHistory)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.stopChan)
	assert.NotNil(t, client.logger)
}

func TestNewAztecFeederClient_WithOptions(t *testing.T) {
	logger := slog.Default()
	rpcURL := "http://test.example.com"

	client := NewAztecFeederClient(logger, rpcURL, WithHistorySize(10))

	assert.Equal(t, 10, client.maxHistorySize)
}

// TestAztecFeederClient_L2TipsSchemaPin guards against silent drift in the
// node_getL2Tips RPC response schema. The fixture under
// testdata/aztec_rpc/node_getL2Tips.json is a real response captured from the
// mainnet endpoint. If Aztec renames a field again (as happened with
// `latest` → `proposed` in the 4.x release), this test fails loudly — which
// is what we want, because the old behaviour silently marked every peer as
// "syncing". See #203.
func TestAztecFeederClient_L2TipsSchemaPin(t *testing.T) {
	testTipsSchemaPin(t, "../testdata/aztec_rpc/node_getL2Tips.json")
}

// Same schema pin for the v5 method, against a real mainnet response.
func TestAztecFeederClient_ChainTipsSchemaPin(t *testing.T) {
	testTipsSchemaPin(t, "../testdata/aztec_rpc/node_getChainTips.json")
}

func testTipsSchemaPin(t *testing.T, fixture string) {
	t.Helper()

	body, err := os.ReadFile(fixture)
	require.NoError(t, err, "fixture must exist")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)

	client := NewAztecFeederClient(slog.Default(), server.URL)
	require.NoError(t, client.updateL2Tips(context.Background()))

	tips := client.GetL2Tips()
	require.NotNil(t, tips)

	// Zero Proposed means the response parsed but nothing landed in the
	// struct — the failure mode behind #203. Proven/finalised assert the
	// nested {block, checkpoint} shape.
	assert.Greater(t, tips.Proposed.Number, uint64(0), "Proposed.Number must parse from the fixture")
	assert.NotEmpty(t, tips.Proposed.Hash, "Proposed.Hash must parse from the fixture")
	assert.Greater(t, tips.Finalised.Block.Number, uint64(0), "Finalised.Block.Number must parse")
	assert.NotEmpty(t, tips.Finalised.Block.Hash, "Finalised.Block.Hash must parse")
	assert.Greater(t, tips.Proven.Block.Number, uint64(0), "Proven.Block.Number must parse")
	assert.NotEmpty(t, tips.Proven.Block.Hash, "Proven.Block.Hash must parse")
}

func TestAztecFeederClient_UpdateL2Tips_Success(t *testing.T) {
	expectedTips := &L2Tips{
		Proposed: BlockInfo{Number: 100, Hash: "0xlatest"},
		Finalised: NestedTip{
			Block: BlockInfo{Number: 95, Hash: "0xfinalised"},
		},
		Proven: NestedTip{
			Block: BlockInfo{Number: 90, Hash: "0xproven"},
		},
	}

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": expectedTips,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx := context.Background()
	err := client.updateL2Tips(ctx)

	require.NoError(t, err)

	tips := client.GetL2Tips()
	require.NotNil(t, tips)
	assert.Equal(t, expectedTips.Proposed.Number, tips.Proposed.Number)
	assert.Equal(t, expectedTips.Proposed.Hash, tips.Proposed.Hash)
	assert.Equal(t, expectedTips.Finalised.Block.Number, tips.Finalised.Block.Number)
	assert.Equal(t, expectedTips.Finalised.Block.Hash, tips.Finalised.Block.Hash)
	assert.Equal(t, expectedTips.Proven.Block.Number, tips.Proven.Block.Number)
	assert.Equal(t, expectedTips.Proven.Block.Hash, tips.Proven.Block.Hash)
}

// Serves exactly one tips method (-32601 otherwise); counts per-method requests.
type versionedMockServer struct {
	*httptest.Server
	mu       sync.Mutex
	method   string
	tips     *L2Tips
	requests map[string]int
}

func newVersionedMockServer(t *testing.T, method string, tips *L2Tips) *versionedMockServer {
	m := &versionedMockServer{method: method, tips: tips, requests: map[string]int{}}
	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
			ID     int    `json:"id"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

		m.mu.Lock()
		m.requests[req.Method]++
		serving, servingTips := m.method, m.tips
		m.mu.Unlock()

		response := map[string]interface{}{"jsonrpc": "2.0", "id": req.ID}
		if req.Method == serving {
			response["result"] = servingTips
		} else {
			response["error"] = map[string]interface{}{
				"code":    -32601,
				"message": "Method not found: " + req.Method,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(response))
	}))
	t.Cleanup(m.Close)
	return m
}

func (m *versionedMockServer) serve(method string, tips *L2Tips) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.method = method
	m.tips = tips
}

func (m *versionedMockServer) requestCount(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.requests[method]
}

func TestAztecFeederClient_UpdateL2Tips_ChainTipsMethod(t *testing.T) {
	tips := &L2Tips{Proposed: BlockInfo{Number: 100, Hash: "0xlatest"}}
	server := newVersionedMockServer(t, methodGetChainTips, tips)

	client := NewAztecFeederClient(slog.Default(), server.URL)
	require.NoError(t, client.updateL2Tips(context.Background()))

	got := client.GetL2Tips()
	require.NotNil(t, got)
	assert.Equal(t, uint64(100), got.Proposed.Number)
	assert.Equal(t, methodGetChainTips, client.tipsMethod)
	assert.Equal(t, 0, server.requestCount(methodGetL2Tips), "no fallback needed on a v5 node")
}

// Pre-v5 node: the first update falls back to node_getL2Tips and pins it.
func TestAztecFeederClient_UpdateL2Tips_FallbackPinsL2Tips(t *testing.T) {
	tips := &L2Tips{Proposed: BlockInfo{Number: 100, Hash: "0xlatest"}}
	server := newVersionedMockServer(t, methodGetL2Tips, tips)

	client := NewAztecFeederClient(slog.Default(), server.URL)

	require.NoError(t, client.updateL2Tips(context.Background()))
	assert.Equal(t, methodGetL2Tips, client.tipsMethod)

	require.NoError(t, client.updateL2Tips(context.Background()))

	assert.Equal(t, 1, server.requestCount(methodGetChainTips), "new method probed only once")
	assert.Equal(t, 2, server.requestCount(methodGetL2Tips), "pinned method used for both updates")
}

// A node upgrade mid-flight clears the pin and re-probes on the next poll.
func TestAztecFeederClient_UpdateL2Tips_RepinsAfterNodeUpgrade(t *testing.T) {
	server := newVersionedMockServer(t, methodGetL2Tips,
		&L2Tips{Proposed: BlockInfo{Number: 100, Hash: "0xold"}})

	client := NewAztecFeederClient(slog.Default(), server.URL)
	require.NoError(t, client.updateL2Tips(context.Background()))
	assert.Equal(t, methodGetL2Tips, client.tipsMethod)
	assert.Equal(t, uint64(100), client.GetLatestHeight())

	server.serve(methodGetChainTips,
		&L2Tips{Proposed: BlockInfo{Number: 200, Hash: "0xnew"}})

	require.NoError(t, client.updateL2Tips(context.Background()))
	assert.Equal(t, methodGetChainTips, client.tipsMethod)
	assert.Equal(t, uint64(200), client.GetLatestHeight())
}

func TestAztecFeederClient_UpdateL2Tips_RPCError(t *testing.T) {
	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": map[string]interface{}{
			"error": map[string]interface{}{
				"code":    -32000,
				"message": "Internal error",
			},
		},
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx := context.Background()
	err := client.updateL2Tips(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "RPC error -32000: Internal error")
}

func TestAztecFeederClient_UpdateL2Tips_NetworkError(t *testing.T) {
	logger := slog.Default()
	client := NewAztecFeederClient(logger, "http://nonexistent.example.com")

	ctx := context.Background()
	err := client.updateL2Tips(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestAztecFeederClient_GetLatestHeight(t *testing.T) {
	expectedTips := &L2Tips{
		Proposed: BlockInfo{Number: 150, Hash: "0xlatest"},
	}

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": expectedTips,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	// Initially should return 0
	height := client.GetLatestHeight()
	assert.Equal(t, uint64(0), height)

	// After updating tips
	ctx := context.Background()
	err := client.updateL2Tips(ctx)
	require.NoError(t, err)

	height = client.GetLatestHeight()
	assert.Equal(t, uint64(150), height)
}

func TestAztecFeederClient_TipsHistory(t *testing.T) {
	logger := slog.Default()
	client := NewAztecFeederClient(logger, "http://test.example.com", WithHistorySize(3))

	// Test storing tips in history
	tips1 := L2Tips{Proposed: BlockInfo{Number: 100, Hash: "0x100"}}
	tips2 := L2Tips{Proposed: BlockInfo{Number: 101, Hash: "0x101"}}
	tips3 := L2Tips{Proposed: BlockInfo{Number: 102, Hash: "0x102"}}
	tips4 := L2Tips{Proposed: BlockInfo{Number: 103, Hash: "0x103"}}

	client.storeTipsInHistory(tips1)
	client.storeTipsInHistory(tips2)
	client.storeTipsInHistory(tips3)

	// Should find all stored tips
	assert.Equal(t, "0x100", client.findLatestBlockHashInHistory(100))
	assert.Equal(t, "0x101", client.findLatestBlockHashInHistory(101))
	assert.Equal(t, "0x102", client.findLatestBlockHashInHistory(102))

	// Add one more to overflow the buffer (size 3)
	client.storeTipsInHistory(tips4)

	// Should find newer tips but not the oldest
	assert.Equal(t, "", client.findLatestBlockHashInHistory(100)) // Overwritten
	assert.Equal(t, "0x101", client.findLatestBlockHashInHistory(101))
	assert.Equal(t, "0x102", client.findLatestBlockHashInHistory(102))
	assert.Equal(t, "0x103", client.findLatestBlockHashInHistory(103))
}

func TestAztecFeederClient_FindLatestBlockHashInHistory(t *testing.T) {
	logger := slog.Default()
	client := NewAztecFeederClient(logger, "http://test.example.com")

	// Store some historical tips
	tips := L2Tips{
		Proposed:  BlockInfo{Number: 100, Hash: "0x100"},
		Finalised: NestedTip{Block: BlockInfo{Number: 95, Hash: "0x95"}},
		Proven:    NestedTip{Block: BlockInfo{Number: 90, Hash: "0x90"}},
	}
	client.storeTipsInHistory(tips)

	// Test finding hashes - only latest blocks are stored in latestHistory
	assert.Equal(t, "0x100", client.findLatestBlockHashInHistory(100))

	// Finalised and proven blocks won't be found in latestHistory
	assert.Equal(t, "", client.findLatestBlockHashInHistory(95))
	assert.Equal(t, "", client.findLatestBlockHashInHistory(90))

	// Test non-existent block
	assert.Equal(t, "", client.findLatestBlockHashInHistory(999))
}

func TestAztecFeederClient_ValidateNodeSync_Success(t *testing.T) {
	tips := &L2Tips{
		Proposed:  BlockInfo{Number: 100, Hash: "0xlatest100"},
		Finalised: NestedTip{Block: BlockInfo{Number: 95, Hash: "0xfinalised95"}},
		Proven:    NestedTip{Block: BlockInfo{Number: 90, Hash: "0xproven90"}},
	}

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": tips,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx := context.Background()
	err := client.updateL2Tips(ctx)
	require.NoError(t, err)

	// Test perfectly synced node
	nodeStatus := &AztecNodeStatus{
		CompressedComponentsVersion: "1.0.0",
		LatestBlockNumber:           100,
		LatestBlockHash:             "0xlatest100",
	}

	// Get the current tips to use as reference
	referenceTips := client.GetL2Tips()
	require.NotNil(t, referenceTips)

	synced, obtained, reason, details := client.ValidateNodeSync(referenceTips, nodeStatus, 5)
	assert.True(t, synced)
	assert.True(t, obtained)
	assert.Equal(t, ReasonInWindow, reason)
	assert.Equal(t, "node is synced", details)
}

func TestAztecFeederClient_ValidateNodeSync_LatestBlockBehind(t *testing.T) {
	tips := &L2Tips{
		Proposed:  BlockInfo{Number: 100, Hash: "0xlatest100"},
		Finalised: NestedTip{Block: BlockInfo{Number: 95, Hash: "0xfinalised95"}},
	}

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": tips,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx := context.Background()
	err := client.updateL2Tips(ctx)
	require.NoError(t, err)

	// Test node too far behind (maxBlockDiff = 5, node at 90, reference at 100)
	nodeStatus := &AztecNodeStatus{
		LatestBlockNumber: 90, // 10 blocks behind, exceeds maxBlockDiff of 5
		LatestBlockHash:   "0xhash90",
	}

	// Get the current tips to use as reference
	referenceTips := client.GetL2Tips()
	require.NotNil(t, referenceTips)

	synced, obtained, reason, details := client.ValidateNodeSync(referenceTips, nodeStatus, 5)
	assert.False(t, synced)
	assert.True(t, obtained, "reference tip is populated, so we can determine sync state")
	assert.Equal(t, ReasonBehind, reason)
	assert.Contains(t, details, "latest block number too far behind")
	assert.Contains(t, details, "node=90, reference=100")
}

func TestAztecFeederClient_ValidateNodeSync_HashMismatch(t *testing.T) {
	client := setupAztecFeederClientWithTips(t)

	// Test hash mismatch for latest block
	nodeStatus := &AztecNodeStatus{
		LatestBlockNumber: 100,
		LatestBlockHash:   "0xwronghash", // Wrong hash
	}

	// Get the current tips to use as reference
	referenceTips := client.GetL2Tips()
	require.NotNil(t, referenceTips)

	synced, obtained, reason, details := client.ValidateNodeSync(referenceTips, nodeStatus, 5)
	assert.False(t, synced)
	assert.True(t, obtained)
	assert.Equal(t, ReasonHashNotInHistory, reason)
	assert.Contains(t, details, "latest block hash not found in history")
	assert.Contains(t, details, "node=0xwronghash")
}

func TestAztecFeederClient_ValidateNodeSync_NoTips(t *testing.T) {
	logger := slog.Default()
	client := NewAztecFeederClient(logger, "http://test.example.com")

	// No tips loaded yet
	nodeStatus := &AztecNodeStatus{
		LatestBlockNumber: 100,
		LatestBlockHash:   "0xlatest100",
	}

	// Try to validate with nil reference tips
	synced, obtained, reason, details := client.ValidateNodeSync(nil, nodeStatus, 5)
	assert.False(t, synced)
	assert.False(t, obtained, "no reference tips → unknown, not false")
	assert.Equal(t, ReasonNoReferenceTips, reason)
	assert.Equal(t, "no reference L2 tips available", details)
}

func TestAztecFeederClient_ValidateNodeSync_PeerAheadOfReference(t *testing.T) {
	// Reference is stuck at 100; a well-behaved peer reports 200. We can't
	// judge whether the peer is really synced — the unknown branch should fire.
	client := setupAztecFeederClientWithTips(t)

	nodeStatus := &AztecNodeStatus{
		LatestBlockNumber: 200, // 100 ahead of reference (maxBlockDiff=5)
		LatestBlockHash:   "0xahead",
	}

	referenceTips := client.GetL2Tips()
	require.NotNil(t, referenceTips)

	synced, obtained, reason, details := client.ValidateNodeSync(referenceTips, nodeStatus, 5)
	assert.False(t, synced)
	assert.False(t, obtained, "peer ahead of reference → unknown")
	assert.Equal(t, ReasonAheadOfReference, reason)
	assert.Contains(t, details, "peer is ahead of reference")
	assert.Contains(t, details, "reference may be stale")
}

func TestAztecFeederClient_ValidateNodeSync_WithHistory(t *testing.T) {
	tips := &L2Tips{
		Proposed:  BlockInfo{Number: 100, Hash: "0xlatest100"},
		Finalised: NestedTip{Block: BlockInfo{Number: 95, Hash: "0xfinalised95"}},
	}

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": tips,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx := context.Background()
	err := client.updateL2Tips(ctx)
	require.NoError(t, err)

	// Store some historical tips that the node might be at
	historicalTips := L2Tips{
		Proposed:  BlockInfo{Number: 98, Hash: "0xhistorical98"},
		Finalised: NestedTip{Block: BlockInfo{Number: 93, Hash: "0xhistorical93"}},
	}
	client.storeTipsInHistory(historicalTips)

	// Test node at historical block (within tolerance)
	nodeStatus := &AztecNodeStatus{
		LatestBlockNumber: 98, // 2 blocks behind, within maxBlockDiff of 5
		LatestBlockHash:   "0xhistorical98",
	}

	// Get the current tips to use as reference
	referenceTips := client.GetL2Tips()
	require.NotNil(t, referenceTips)

	synced, obtained, reason, details := client.ValidateNodeSync(referenceTips, nodeStatus, 5)
	assert.True(t, synced)
	assert.True(t, obtained)
	assert.Equal(t, ReasonInWindow, reason)
	assert.Equal(t, "node is synced", details)
}

func TestAztecFeederClient_ValidateNodeSync_HistoryHashNotFound(t *testing.T) {
	client := setupAztecFeederClientWithTips(t)

	// Test node at block within tolerance but with wrong hash
	nodeStatus := &AztecNodeStatus{
		LatestBlockNumber: 98,               // Within tolerance
		LatestBlockHash:   "0xnotinhistory", // Hash not in history
	}

	// Get the current tips to use as reference
	referenceTips := client.GetL2Tips()
	require.NotNil(t, referenceTips)

	synced, obtained, reason, details := client.ValidateNodeSync(referenceTips, nodeStatus, 5)
	assert.False(t, synced)
	assert.True(t, obtained)
	assert.Equal(t, ReasonHashNotInHistory, reason)
	assert.Contains(t, details, "latest block hash not found in history")
	assert.Contains(t, details, "height 98")
}

func TestAztecFeederClient_Start(t *testing.T) {
	tips := &L2Tips{
		Proposed: BlockInfo{Number: 100, Hash: "0xlatest100"},
	}

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": tips,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := client.Start(ctx)
	require.NoError(t, err)

	// Should have populated initial tips
	height := client.GetLatestHeight()
	assert.Equal(t, uint64(100), height)

	// Cancel context to stop polling
	cancel()
}

func TestAztecFeederClient_Start_InitialError(t *testing.T) {
	// Server that returns error
	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": map[string]interface{}{
			"error": map[string]interface{}{
				"code":    -32000,
				"message": "Server error",
			},
		},
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx := context.Background()
	err := client.Start(ctx)

	// A failed first poll starts the client degraded instead of failing.
	require.NoError(t, err)
	assert.Nil(t, client.GetL2Tips())
}

func TestAztecFeederClient_ConcurrentAccess(t *testing.T) {
	tips := &L2Tips{
		Proposed: BlockInfo{Number: 100, Hash: "0xlatest100"},
	}

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": tips,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx := context.Background()
	err := client.updateL2Tips(ctx)
	require.NoError(t, err)

	// Test concurrent access to GetLatestHeight and GetL2Tips
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			height := client.GetLatestHeight()
			assert.Equal(t, uint64(100), height)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			tips := client.GetL2Tips()
			require.NotNil(t, tips)
			assert.Equal(t, uint64(100), tips.Proposed.Number)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}

func TestAztecFeederClient_EdgeCases(t *testing.T) {
	t.Run("Empty response result", func(t *testing.T) {
		server := createMockAztecServer(t, map[string]interface{}{
			"node_getL2Tips": nil,
		})
		defer server.Close()

		logger := slog.Default()
		client := NewAztecFeederClient(logger, server.URL)

		ctx := context.Background()
		err := client.updateL2Tips(ctx)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty result from RPC")
	})

	t.Run("Malformed JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		logger := slog.Default()
		client := NewAztecFeederClient(logger, server.URL)

		ctx := context.Background()
		err := client.updateL2Tips(ctx)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal response")
	})

	t.Run("HTTP error status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		logger := slog.Default()
		client := NewAztecFeederClient(logger, server.URL)

		ctx := context.Background()
		err := client.updateL2Tips(ctx)

		require.Error(t, err)
		// Should succeed in reading the response body but fail to unmarshal
		assert.Contains(t, strings.ToLower(err.Error()), "unmarshal")
	})
}

func TestAztecFeederClient_UpdateNodeInfo_Success(t *testing.T) {
	// Valid ENR with aztec field containing CCV (from enr_test.go)
	nodeInfo := &NodeInfo{
		NodeVersion:     "1.0.0",
		L1ChainID:       1,
		ProtocolVersion: 1,
		ENR:             testValidENR,
	}

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getNodeInfo": nodeInfo,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx := context.Background()
	err := client.updateNodeInfo(ctx)

	require.NoError(t, err)

	// Verify ENR info was decoded and stored
	ccv := client.GetReferenceCCV()
	assert.Equal(t, testExpectedCCV, ccv, "Should have extracted correct CCV from the ENR")
}

func TestAztecFeederClient_UpdateNodeInfo_RPCError(t *testing.T) {
	server := createMockAztecServer(t, map[string]interface{}{
		"node_getNodeInfo": map[string]interface{}{
			"error": map[string]interface{}{
				"code":    -32000,
				"message": "Internal error",
			},
		},
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx := context.Background()
	err := client.updateNodeInfo(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "RPC error -32000: Internal error")
}

func TestAztecFeederClient_UpdateNodeInfo_EmptyENR(t *testing.T) {
	nodeInfo := &NodeInfo{
		NodeVersion:     "1.0.0",
		L1ChainID:       1,
		ProtocolVersion: 1,
		ENR:             "", // Empty ENR
	}

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getNodeInfo": nodeInfo,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx := context.Background()
	err := client.updateNodeInfo(ctx)

	// Should succeed but not set ENR info
	require.NoError(t, err)

	// CCV should be empty since ENR was empty
	ccv := client.GetReferenceCCV()
	assert.Empty(t, ccv)
}

func TestAztecFeederClient_UpdateNodeInfo_InvalidENR(t *testing.T) {
	nodeInfo := &NodeInfo{
		NodeVersion:     "1.0.0",
		L1ChainID:       1,
		ProtocolVersion: 1,
		ENR:             "invalid-enr-string",
	}

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getNodeInfo": nodeInfo,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx := context.Background()
	err := client.updateNodeInfo(ctx)

	// Should fail because ENR decoding fails
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode ENR")
}

func TestAztecFeederClient_GetReferenceCCV_NotInitialized(t *testing.T) {
	logger := slog.Default()
	client := NewAztecFeederClient(logger, "http://test.example.com")

	// Before any node info is loaded
	ccv := client.GetReferenceCCV()
	assert.Empty(t, ccv)
}

func TestAztecFeederClient_GetReferenceCCV_WithValidENR(t *testing.T) {
	nodeInfo := &NodeInfo{
		NodeVersion:     "1.0.0",
		L1ChainID:       1,
		ProtocolVersion: 1,
		ENR:             testValidENR,
	}

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getNodeInfo": nodeInfo,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx := context.Background()
	err := client.updateNodeInfo(ctx)
	require.NoError(t, err)

	// Should have decoded and stored CCV
	ccv := client.GetReferenceCCV()
	assert.Equal(t, testExpectedCCV, ccv, "Should have extracted correct CCV from the ENR")
}

func TestAztecFeederClient_Start_WithNodeInfo(t *testing.T) {
	tips := &L2Tips{
		Proposed: BlockInfo{Number: 100, Hash: "0xlatest100"},
	}

	nodeInfo := &NodeInfo{
		NodeVersion:     "1.0.0",
		L1ChainID:       1,
		ProtocolVersion: 1,
		ENR:             testValidENR,
	}

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips":   tips,
		"node_getNodeInfo": nodeInfo,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := client.Start(ctx)
	require.NoError(t, err)

	// Should have populated initial tips
	height := client.GetLatestHeight()
	assert.Equal(t, uint64(100), height)

	// Should have populated ENR info (may warn but shouldn't fail startup)
	ccv := client.GetReferenceCCV()
	assert.Equal(t, testExpectedCCV, ccv, "Should have extracted correct CCV from the ENR")

	// Cancel context to stop polling
	cancel()
}

func TestAztecFeederClient_Start_NodeInfoFailure(t *testing.T) {
	tips := &L2Tips{
		Proposed: BlockInfo{Number: 100, Hash: "0xlatest100"},
	}

	// NodeInfo returns error but tips work
	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": tips,
		"node_getNodeInfo": map[string]interface{}{
			"error": map[string]interface{}{
				"code":    -32000,
				"message": "Node info not available",
			},
		},
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Should still start successfully even if NodeInfo fails (just logs warning)
	err := client.Start(ctx)
	require.NoError(t, err)

	// Should have populated initial tips
	height := client.GetLatestHeight()
	assert.Equal(t, uint64(100), height)

	// CCV will be empty since NodeInfo failed
	ccv := client.GetReferenceCCV()
	assert.Empty(t, ccv)

	cancel()
}

func TestAztecFeederClient_ConcurrentAccess_WithCCV(t *testing.T) {
	tips := &L2Tips{
		Proposed: BlockInfo{Number: 100, Hash: "0xlatest100"},
	}

	nodeInfo := &NodeInfo{
		NodeVersion:     "1.0.0",
		L1ChainID:       1,
		ProtocolVersion: 1,
		ENR:             testValidENR,
	}

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips":   tips,
		"node_getNodeInfo": nodeInfo,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	ctx := context.Background()
	err := client.updateL2Tips(ctx)
	require.NoError(t, err)
	err = client.updateNodeInfo(ctx)
	require.NoError(t, err)

	// Test concurrent access to GetLatestHeight, GetL2Tips, and GetReferenceCCV
	done := make(chan bool, 3)

	go func() {
		for i := 0; i < 100; i++ {
			height := client.GetLatestHeight()
			assert.Equal(t, uint64(100), height)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			tips := client.GetL2Tips()
			require.NotNil(t, tips)
			assert.Equal(t, uint64(100), tips.Proposed.Number)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			ccv := client.GetReferenceCCV()
			assert.Equal(t, testExpectedCCV, ccv)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	<-done
	<-done
	<-done
}

func TestAztecFeederClient_GetTipsAge_BeforeFirstSuccess(t *testing.T) {
	logger := slog.Default()
	client := NewAztecFeederClient(logger, "http://test.example.com")

	// Before any successful update, age is the negative sentinel.
	age := client.GetTipsAge()
	assert.Equal(t, time.Duration(-1), age, "sentinel before first success")
}

func TestAztecFeederClient_GetTipsAge_AfterSuccess(t *testing.T) {
	tips := &L2Tips{
		Proposed: BlockInfo{Number: 100, Hash: "0xlatest100"},
	}
	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": tips,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	require.NoError(t, client.updateL2Tips(context.Background()))

	age := client.GetTipsAge()
	assert.Greater(t, age, time.Duration(0), "age must be positive after success")
	assert.Less(t, age, time.Second, "age should be sub-second immediately after update")
}

func TestAztecFeederClient_GetTipsAge_NotResetOnFailure(t *testing.T) {
	tips := &L2Tips{
		Proposed: BlockInfo{Number: 100, Hash: "0xlatest100"},
	}
	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": tips,
	})
	defer server.Close()

	logger := slog.Default()
	client := NewAztecFeederClient(logger, server.URL)

	// First success captures lastTipsSuccessAt.
	require.NoError(t, client.updateL2Tips(context.Background()))
	firstAge := client.GetTipsAge()

	// Switch to a failing endpoint and call again.
	client.rpcURL = "http://nonexistent.example.com"
	require.Error(t, client.updateL2Tips(context.Background()))

	// Age must keep growing — failure does NOT reset it.
	secondAge := client.GetTipsAge()
	assert.GreaterOrEqual(t, secondAge, firstAge,
		"failure must not reset lastTipsSuccessAt")
}

func pollErrorsCount(t *testing.T, reason string) float64 {
	t.Helper()
	return testutil.ToFloat64(pollErrorsTotal.WithLabelValues(reason))
}

func TestUpdateL2Tips_PollErrorsCounter_HTTP(t *testing.T) {
	before := pollErrorsCount(t, "http")

	logger := slog.Default()
	client := NewAztecFeederClient(logger, "http://nonexistent.example.com")
	require.Error(t, client.updateL2Tips(context.Background()))

	after := pollErrorsCount(t, "http")
	assert.InDelta(t, 1.0, after-before, 0.0001, "http counter incremented by 1")
}

func TestUpdateL2Tips_PollErrorsCounter_RPCError(t *testing.T) {
	before := pollErrorsCount(t, "rpc_error")

	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": map[string]interface{}{
			"error": map[string]interface{}{
				"code":    -32000,
				"message": "Internal error",
			},
		},
	})
	defer server.Close()

	client := NewAztecFeederClient(slog.Default(), server.URL)
	require.Error(t, client.updateL2Tips(context.Background()))

	after := pollErrorsCount(t, "rpc_error")
	assert.InDelta(t, 1.0, after-before, 0.0001, "rpc_error counter incremented by 1")
}

//nolint:dupl // structurally similar to EmptyResult by design: each branch needs its own test.
func TestUpdateL2Tips_PollErrorsCounter_Decode(t *testing.T) {
	before := pollErrorsCount(t, "decode")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not valid json"))
	}))
	t.Cleanup(server.Close)

	client := NewAztecFeederClient(slog.Default(), server.URL)
	require.Error(t, client.updateL2Tips(context.Background()))

	after := pollErrorsCount(t, "decode")
	assert.InDelta(t, 1.0, after-before, 0.0001, "decode counter incremented by 1")
}

//nolint:dupl // structurally similar to Decode by design: each branch needs its own test.
func TestUpdateL2Tips_PollErrorsCounter_EmptyResult(t *testing.T) {
	before := pollErrorsCount(t, "empty_result")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":null}`))
	}))
	t.Cleanup(server.Close)

	client := NewAztecFeederClient(slog.Default(), server.URL)
	require.Error(t, client.updateL2Tips(context.Background()))

	after := pollErrorsCount(t, "empty_result")
	assert.InDelta(t, 1.0, after-before, 0.0001, "empty_result counter incremented by 1")
}

func TestUpdateL2Tips_PollErrorsCounter_SchemaMismatch(t *testing.T) {
	before := pollErrorsCount(t, "schema_mismatch")

	// Result shape doesn't match `proposed` — Number=0, Hash="" trips the schema guard.
	emptyTips := &L2Tips{}
	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": emptyTips,
	})
	defer server.Close()

	client := NewAztecFeederClient(slog.Default(), server.URL)
	require.Error(t, client.updateL2Tips(context.Background()))

	after := pollErrorsCount(t, "schema_mismatch")
	assert.InDelta(t, 1.0, after-before, 0.0001, "schema_mismatch counter incremented by 1")
}

func TestUpdateL2Tips_PollErrorsCounter_NotIncrementedOnSuccess(t *testing.T) {
	tips := &L2Tips{Proposed: BlockInfo{Number: 100, Hash: "0xlatest"}}
	server := createMockAztecServer(t, map[string]interface{}{
		"node_getL2Tips": tips,
	})
	defer server.Close()

	// Sum across every reason label so we don't have to enumerate them.
	totalBefore := pollErrorsCount(t, "http") +
		pollErrorsCount(t, "rpc_error") +
		pollErrorsCount(t, "decode") +
		pollErrorsCount(t, "empty_result") +
		pollErrorsCount(t, "schema_mismatch")

	client := NewAztecFeederClient(slog.Default(), server.URL)
	require.NoError(t, client.updateL2Tips(context.Background()))

	totalAfter := pollErrorsCount(t, "http") +
		pollErrorsCount(t, "rpc_error") +
		pollErrorsCount(t, "decode") +
		pollErrorsCount(t, "empty_result") +
		pollErrorsCount(t, "schema_mismatch")

	assert.InDelta(t, 0.0, totalAfter-totalBefore, 0.0001,
		"no counter advances on a successful poll")
}
