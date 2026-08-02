package core

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for AztecFeederClient against real Aztec RPC endpoints
// These tests require a running Aztec node or access to a live RPC endpoint
// Set AZTEC_RPC_URL environment variable to run these tests

func getTestRPCURL() string {
	return os.Getenv("AZTEC_RPC_URL")
}

func TestAztecFeederClient_Integration_RealRPC(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	rpcURL := getTestRPCURL()
	if rpcURL == "" {
		t.Skip("Skipping integration test: AZTEC_RPC_URL not set")
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	client := NewAztecFeederClient(logger, rpcURL, WithHistorySize(5))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("UpdateL2Tips", func(t *testing.T) {
		err := client.updateL2Tips(ctx)
		if err != nil {
			t.Logf("Failed to update L2 tips: %v", err)
			t.Logf("This might be expected if the RPC endpoint is not available or doesn't support node_getChainTips/node_getL2Tips")
			t.Skip("RPC endpoint not available or doesn't support required method")
		}

		tips := client.GetL2Tips()
		require.NotNil(t, tips, "L2 tips should not be nil after successful update")

		t.Logf("Retrieved L2 Tips:")
		t.Logf("  Proposed: Block %d, Hash: %s", tips.Proposed.Number, tips.Proposed.Hash)
		t.Logf("  Finalised: Block %d, Hash: %s", tips.Finalised.Block.Number, tips.Finalised.Block.Hash)
		t.Logf("  Proven:    Block %d, Hash: %s", tips.Proven.Block.Number, tips.Proven.Block.Hash)

		// Basic sanity checks
		assert.Greater(t, tips.Proposed.Number, uint64(0), "Proposed block number should be greater than 0")
		assert.NotEmpty(t, tips.Proposed.Hash, "Proposed block hash should not be empty")

		// Finalised block should be less than or equal to proposed
		assert.LessOrEqual(t, tips.Finalised.Block.Number, tips.Proposed.Number,
			"Finalised block should be <= proposed block")

		// Proven blocks can be ahead of finalised blocks on Aztec.
		t.Logf("Block relationship: Proposed: %d, Finalised: %d, Proven: %d",
			tips.Proposed.Number, tips.Finalised.Block.Number, tips.Proven.Block.Number)

		// Hash format validation (should start with 0x and be hex)
		assert.Regexp(t, `^0x[0-9a-fA-F]+$`, tips.Proposed.Hash, "Proposed hash should be valid hex")
		assert.Regexp(t, `^0x[0-9a-fA-F]+$`, tips.Finalised.Block.Hash, "Finalised hash should be valid hex")
		assert.Regexp(t, `^0x[0-9a-fA-F]+$`, tips.Proven.Block.Hash, "Proven hash should be valid hex")
	})

	t.Run("GetLatestHeight", func(t *testing.T) {
		// First update tips
		err := client.updateL2Tips(ctx)
		if err != nil {
			t.Skip("Cannot test GetLatestHeight without successful RPC connection")
		}

		height := client.GetLatestHeight()
		assert.Greater(t, height, uint64(0), "Latest height should be greater than 0")
		t.Logf("Latest height: %d", height)
	})

	t.Run("StartAndPoll", func(t *testing.T) {
		// Create a new client for this test to avoid interference
		pollClient := NewAztecFeederClient(logger, rpcURL, WithHistorySize(3))

		// Start the client
		err := pollClient.Start(ctx)
		if err != nil {
			t.Logf("Failed to start client: %v", err)
			t.Skip("Cannot start client - RPC endpoint might not be available")
		}

		// Wait a moment to ensure initial tips are loaded
		time.Sleep(1 * time.Second)

		// Verify that tips were loaded
		tips := pollClient.GetL2Tips()
		require.NotNil(t, tips, "Tips should be loaded after Start()")

		initialHeight := tips.Proposed.Number
		t.Logf("Initial height after start: %d", initialHeight)

		// The polling should continue in the background
		// We can't easily test that polling is working without waiting for new blocks,
		// but we can verify that the client is functional
		height := pollClient.GetLatestHeight()
		assert.Equal(t, initialHeight, height, "Height should match the tips data")
	})
}

func TestAztecFeederClient_Integration_NodeSyncValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	rpcURL := getTestRPCURL()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	client := NewAztecFeederClient(logger, rpcURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Update tips first
	err := client.updateL2Tips(ctx)
	if err != nil {
		t.Skip("Cannot perform validation test without RPC connection")
	}

	tips := client.GetL2Tips()
	require.NotNil(t, tips)

	t.Run("ValidatePerfectlySync", func(t *testing.T) {
		// Create a node status that matches the current tips exactly
		nodeStatus := &AztecNodeStatus{
			CompressedComponentsVersion: "test-integration-1.0.0",
			LatestBlockNumber:           tips.Proposed.Number,
			LatestBlockHash:             tips.Proposed.Hash,
		}

		// Get reference tips for validation
		referenceTips := client.GetL2Tips()
		require.NotNil(t, referenceTips)

		synced, obtained, _, details := client.ValidateNodeSync(referenceTips, nodeStatus, 10)
		assert.True(t, synced, "Node with matching tips should be considered synced")
		assert.True(t, obtained)
		assert.Equal(t, "node is synced", details)
		t.Logf("Validation result for perfect sync: %s", details)
	})

	t.Run("ValidateSlightlyBehind", func(t *testing.T) {
		// Create a node status that's slightly behind but within tolerance
		nodeStatus := &AztecNodeStatus{
			CompressedComponentsVersion: "test-integration-1.0.0",
			LatestBlockNumber:           tips.Proposed.Number - 2,     // 2 blocks behind
			LatestBlockHash:             "0x" + generateRandomHex(64), // Random hash
		}

		// Get reference tips for validation
		referenceTips := client.GetL2Tips()
		require.NotNil(t, referenceTips)

		// This should fail because the hashes won't be in history for a random hash
		synced, _, _, details := client.ValidateNodeSync(referenceTips, nodeStatus, 10)
		assert.False(t, synced, "Node with random hashes should not be considered synced")
		assert.Contains(t, details, "not found in history")
		t.Logf("Validation result for slightly behind with random hash: %s", details)
	})

	t.Run("ValidateTooFarBehind", func(t *testing.T) {
		// Create a node status that's too far behind
		maxBlockDiff := uint64(5)
		nodeStatus := &AztecNodeStatus{
			CompressedComponentsVersion: "test-integration-1.0.0",
			LatestBlockNumber:           tips.Proposed.Number - maxBlockDiff - 1, // Beyond tolerance
			LatestBlockHash:             "0x" + generateRandomHex(64),
		}

		// Get reference tips for validation
		referenceTips := client.GetL2Tips()
		require.NotNil(t, referenceTips)

		synced, _, _, details := client.ValidateNodeSync(referenceTips, nodeStatus, maxBlockDiff)
		assert.False(t, synced, "Node too far behind should not be considered synced")
		assert.Contains(t, details, "too far behind")
		t.Logf("Validation result for too far behind: %s", details)
	})

	t.Run("ValidateHashMismatch", func(t *testing.T) {
		// Create a node status with current block height but wrong hash
		nodeStatus := &AztecNodeStatus{
			CompressedComponentsVersion: "test-integration-1.0.0",
			LatestBlockNumber:           tips.Proposed.Number,
			LatestBlockHash:             "0x" + generateRandomHex(64), // Wrong hash
		}

		// Get reference tips for validation
		referenceTips := client.GetL2Tips()
		require.NotNil(t, referenceTips)

		synced, _, _, details := client.ValidateNodeSync(referenceTips, nodeStatus, 10)
		assert.False(t, synced, "Node with hash mismatch should not be considered synced")
		assert.Contains(t, details, "latest block hash not found in history")
		t.Logf("Validation result for hash mismatch: %s", details)
	})
}

func TestAztecFeederClient_Integration_HistoryTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	rpcURL := getTestRPCURL()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Use a small history size for easier testing
	client := NewAztecFeederClient(logger, rpcURL, WithHistorySize(3))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Update tips multiple times to populate history
	err := client.updateL2Tips(ctx)
	if err != nil {
		t.Skip("Cannot test history tracking without RPC connection")
	}

	tips1 := client.GetL2Tips()
	require.NotNil(t, tips1)

	t.Logf("First tips update - Proposed: %d, Finalised: %d, Proven: %d",
		tips1.Proposed.Number, tips1.Finalised.Block.Number, tips1.Proven.Block.Number)

	// Wait a bit and update again (in case blocks advance)
	time.Sleep(2 * time.Second)

	err = client.updateL2Tips(ctx)
	require.NoError(t, err)

	tips2 := client.GetL2Tips()
	require.NotNil(t, tips2)

	t.Logf("Second tips update - Proposed: %d, Finalised: %d, Proven: %d",
		tips2.Proposed.Number, tips2.Finalised.Block.Number, tips2.Proven.Block.Number)

	// Test hash lookup in history
	t.Run("HistoryLookup", func(t *testing.T) {
		// Try to find the latest block hash from first update
		foundHash := client.findLatestBlockHashInHistory(tips1.Proposed.Number)

		if tips1.Proposed.Number == tips2.Proposed.Number {
			// If blocks haven't advanced, we should find the same hash
			assert.Equal(t, tips1.Proposed.Hash, foundHash,
				"Should find the same hash for the same block number")
		} else {
			// If blocks have advanced, we might still find the old hash in history
			// depending on timing
			t.Logf("Block advanced from %d to %d", tips1.Proposed.Number, tips2.Proposed.Number)
			if foundHash != "" {
				assert.Equal(t, tips1.Proposed.Hash, foundHash,
					"Found hash should match the original hash")
			}
		}

		// Try to find finalised block hash
		foundFinalisedHash := client.findLatestBlockHashInHistory(tips1.Finalised.Block.Number)
		if foundFinalisedHash != "" {
			t.Logf("Found finalised hash in history: %s", foundFinalisedHash)
		}

		// Try to find proven block hash
		foundProvenHash := client.findLatestBlockHashInHistory(tips1.Proven.Block.Number)
		if foundProvenHash != "" {
			t.Logf("Found proven hash in history: %s", foundProvenHash)
		}
	})

	t.Run("HistoryValidation", func(t *testing.T) {
		// Test validation using ValidateNodeSync method
		nodeStatus := &AztecNodeStatus{
			CompressedComponentsVersion: "test-integration-1.0.0",
			LatestBlockNumber:           tips1.Proposed.Number,
			LatestBlockHash:             tips1.Proposed.Hash,
		}

		// Get reference tips for validation
		referenceTips := client.GetL2Tips()
		require.NotNil(t, referenceTips)

		synced, _, _, details := client.ValidateNodeSync(referenceTips, nodeStatus, 10)
		t.Logf("Validation of historical latest block (%d, %s): synced=%v, details=%s",
			tips1.Proposed.Number, tips1.Proposed.Hash, synced, details)

		// Test with invalid hash
		nodeStatus.LatestBlockHash = "0x" + generateRandomHex(64)
		synced, _, _, details = client.ValidateNodeSync(referenceTips, nodeStatus, 10)
		assert.False(t, synced, "Random hash should not validate")
		assert.Contains(t, details, "not found in history")
	})
}

func TestAztecFeederClient_Integration_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	t.Run("InvalidURL", func(t *testing.T) {
		client := NewAztecFeederClient(logger, "http://invalid-url-that-does-not-exist.com")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := client.updateL2Tips(ctx)
		assert.Error(t, err, "Should fail with invalid URL")
		// The error could be network-related, JSON parsing related, or empty result
		errorStr := err.Error()
		assert.True(t,
			errorStr != "" && (strings.Contains(errorStr, "failed to send request") ||
				strings.Contains(errorStr, "failed to unmarshal response") ||
				strings.Contains(errorStr, "failed to read response body") ||
				strings.Contains(errorStr, "invalid character") ||
				strings.Contains(errorStr, "empty result from RPC")),
			"Should fail with network or parsing error, got: %s", errorStr)
		t.Logf("Expected error with invalid URL: %v", err)
	})

	t.Run("InvalidMethod", func(t *testing.T) {
		// Use a valid RPC URL but we'll test what happens with method not found
		// Most RPC endpoints will return a method not found error
		rpcURL := getTestRPCURL()
		if rpcURL == "" {
			t.Skip("Need valid RPC URL for this test")
		}

		client := NewAztecFeederClient(logger, rpcURL)

		// We can't easily test invalid methods without modifying the client,
		// but we can test timeout scenarios
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		err := client.updateL2Tips(ctx)
		if err != nil {
			t.Logf("Error with very short timeout (expected): %v", err)
			assert.Contains(t, err.Error(), "context deadline exceeded")
		}
	})
}

// Helper function to generate random hex string for testing
func generateRandomHex(_ int) string {
	const hexChars = "0123456789abcdef"
	result := make([]byte, 64)
	for i := range result {
		result[i] = hexChars[i%len(hexChars)]
	}
	return string(result)
}

// Benchmark test to measure performance of real RPC calls
func BenchmarkAztecFeederClient_Integration_UpdateL2Tips(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	rpcURL := getTestRPCURL()
	if rpcURL == "" {
		b.Skip("Skipping benchmark: no RPC URL configured")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Reduce logging noise during benchmark
	}))

	client := NewAztecFeederClient(logger, rpcURL)
	ctx := context.Background()

	// Warmup
	err := client.updateL2Tips(ctx)
	if err != nil {
		b.Skipf("Cannot run benchmark: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := client.updateL2Tips(ctx)
		if err != nil {
			b.Fatalf("Benchmark failed on iteration %d: %v", i, err)
		}
	}
}

// Example test that demonstrates how to use the integration tests
func ExampleAztecFeederClient_integration() {
	// Set environment variable: export AZTEC_RPC_URL="https://your-aztec-rpc-endpoint"
	// Run with: go test -v ./core -run TestAztecFeederClient_Integration

	logger := slog.Default()
	client := NewAztecFeederClient(logger, "https://api.aztec.network/aztec-pxe/1")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Update L2 tips from the real network
	err := client.updateL2Tips(ctx)
	if err != nil {
		panic(err)
	}

	// Get the latest height
	height := client.GetLatestHeight()
	println("Latest Aztec block height:", height)

	// Validate a node's sync status
	nodeStatus := &AztecNodeStatus{
		LatestBlockNumber: height,
		LatestBlockHash:   "", // Empty hash for example
	}

	// Get reference tips for validation
	referenceTips := client.GetL2Tips()
	if referenceTips == nil {
		panic("No reference tips available")
	}

	synced, _, _, details := client.ValidateNodeSync(referenceTips, nodeStatus, 10)
	println("Node synced:", synced, "Details:", details)
}
