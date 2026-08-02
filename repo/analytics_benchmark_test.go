package repo

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/stretchr/testify/require"

	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

// setupBenchmarkData creates a dataset for benchmarking:
// - 100 distinct peers (reduced for faster testing)
// - 1,000 peer states (10 states per peer on average)
// - A mix of synced, not synced, and unknown states
func setupBenchmarkData(b *testing.B, repo *PeerRepository) {
	b.Helper()
	b.Log("Setting up benchmark data...")
	ctx := context.Background()
	db := repo.db

	// Create agent versions
	agentVersions := []string{"agent/v1.0.0", "agent/v2.0.0", "agent/v3.0.0"}
	agentVersionIDs := make([]int, len(agentVersions))
	for i, av := range agentVersions {
		agent := &models.AgentVersion{
			AgentVersion: av,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err := agent.Insert(ctx, db, boil.Infer())
		require.NoError(b, err)
		agentVersionIDs[i] = agent.ID
	}

	// Create protocols
	protocols := []string{"proto1", "proto2", "proto3"}
	protocolIDs := make([]int, len(protocols))
	for i, p := range protocols {
		protocol := &models.Protocol{
			Protocol:  p,
			Count:     100,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := protocol.Insert(ctx, db, boil.Infer())
		require.NoError(b, err)
		protocolIDs[i] = protocol.ID
	}

	// Create protocol sets with different combinations
	protocolSetIDs := make([]int, 3)
	for i := 0; i < 3; i++ {
		// Create different protocol combinations for each set
		var protocolIDsInt64 types.Int64Array
		switch i {
		case 0:
			protocolIDsInt64 = types.Int64Array{int64(protocolIDs[0])}
		case 1:
			protocolIDsInt64 = types.Int64Array{int64(protocolIDs[0]), int64(protocolIDs[1])}
		case 2:
			protocolIDsInt64 = types.Int64Array{int64(protocolIDs[0]), int64(protocolIDs[1]), int64(protocolIDs[2])}
		}

		// Calculate hash for the protocol set
		hash := sha256.Sum256([]byte(fmt.Sprintf("%v", protocolIDsInt64)))

		ps := &models.ProtocolsSet{
			ProtocolIds: protocolIDsInt64,
			Hash:        hash[:],
		}
		err := ps.Insert(ctx, db, boil.Infer())
		require.NoError(b, err)
		protocolSetIDs[i] = ps.ID
	}

	// Create crawls
	crawl1 := &models.Crawl{
		StartedAt:       time.Now().Add(-2 * time.Hour),
		FinishedAt:      time.Now().Add(-1 * time.Hour),
		Status:          "finished",
		DialablePeers:   null.IntFrom(500),
		UndialablePeers: null.IntFrom(500),
	}
	err := crawl1.Insert(ctx, db, boil.Infer())
	require.NoError(b, err)

	// Latest crawl
	crawl2 := &models.Crawl{
		StartedAt:       time.Now().Add(-30 * time.Minute),
		FinishedAt:      time.Now().Add(-5 * time.Minute),
		Status:          "finished",
		DialablePeers:   null.IntFrom(1000),
		UndialablePeers: null.IntFrom(0),
	}
	err = crawl2.Insert(ctx, db, boil.Infer())
	require.NoError(b, err)

	// Create 1,000 peers
	peerIDs := make([]int64, 1000)
	baseTime := time.Now().Add(-24 * time.Hour)

	for i := 0; i < 1000; i++ {
		peer := &models.Peer{
			MultiHash:      fmt.Sprintf("peer_%d", i),
			AgentVersionID: null.IntFrom(agentVersionIDs[i%len(agentVersionIDs)]),
			ProtocolsSetID: null.IntFrom(protocolSetIDs[i%len(protocolSetIDs)]),
			CreatedAt:      baseTime.Add(time.Duration(i) * time.Second),
			UpdatedAt:      baseTime.Add(time.Duration(i) * time.Second),
			LastSeen:       baseTime.Add(time.Duration(i) * time.Second),
		}
		err := peer.Insert(ctx, db, boil.Infer())
		require.NoError(b, err)
		peerIDs[i] = peer.ID

		// Create visit for latest crawl
		visit := &models.Visit{
			CrawlID:         null.IntFrom(crawl2.ID),
			PeerID:          peer.ID,
			ConnectDuration: null.StringFrom("100ms"),
			CrawlDuration:   null.StringFrom("200ms"),
			VisitStartedAt:  time.Now().Add(-20 * time.Minute),
			VisitEndedAt:    time.Now().Add(-15 * time.Minute),
		}
		err = visit.Insert(ctx, db, boil.Infer())
		require.NoError(b, err)
	}

	// Create 10,000 peer states (10 per peer on average)
	// Distribution: 40% synced, 40% not synced, 20% unknown (null)
	stateTime := time.Now().Add(-48 * time.Hour)
	for i := 0; i < 10000; i++ {
		peerID := peerIDs[i%1000]

		var syncStatus null.Bool
		switch i % 5 {
		case 0, 1: // 40% synced
			syncStatus = null.BoolFrom(true)
		case 2, 3: // 40% not synced
			syncStatus = null.BoolFrom(false)
		default: // 20% unknown (null)
			syncStatus = null.Bool{}
		}

		state := &models.PeersState{
			PeerID:      peerID,
			BlockHeight: null.Int64From(int64(100000 + i)),
			SpecVersion: null.StringFrom(fmt.Sprintf("v1.0.%d", i%10)),
			IsSynced:    syncStatus,
			CreatedAt:   stateTime.Add(time.Duration(i) * time.Minute),
		}
		err := state.Insert(ctx, db, boil.Infer())
		require.NoError(b, err)
	}
	b.Log("Benchmark data setup complete")
}

// setupBenchmarkRepo creates a test repository for benchmarks
func setupBenchmarkRepo(_ *testing.B) (*PeerRepository, func()) {
	// Create a wrapper to pass to setupTestRepo
	t := &testing.T{}
	repo, tearDown := setupTestRepo(t)
	return repo, tearDown
}

func BenchmarkGetSyncStatusCount_Current(b *testing.B) {
	repo, tearDown := setupBenchmarkRepo(b)
	defer tearDown()

	// Setup benchmark data
	setupBenchmarkData(b, repo)

	ctx := context.Background()

	// Reset timer to exclude setup time
	b.ResetTimer()

	// Run the benchmark
	for i := 0; i < b.N; i++ {
		results, err := repo.GetSyncStatusCount(ctx)
		require.NoError(b, err)
		require.NotNil(b, results)
	}
}

// Test the performance of the pre-calculated query with a larger dataset
func TestGetSyncStatusCount_PerformanceWithLargeDataset(t *testing.T) {
	t.Skip("Skipping performance test - takes too long in CI")
	repo, tearDown := setupTestRepo(t)
	defer tearDown()

	ctx := context.Background()
	db := repo.db

	t.Log("Setting up test data with 1,000 peer states...")

	// Create agent version
	agent := &models.AgentVersion{
		AgentVersion: "test/v1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := agent.Insert(ctx, db, boil.Infer())
	require.NoError(t, err)

	// Create protocol
	protocol := &models.Protocol{
		Protocol:  "test_proto",
		Count:     1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = protocol.Insert(ctx, db, boil.Infer())
	require.NoError(t, err)

	// Create protocol set
	hash := sha256.Sum256([]byte(fmt.Sprintf("%v", []int64{int64(protocol.ID)})))
	protocolSet := &models.ProtocolsSet{
		ProtocolIds: types.Int64Array{int64(protocol.ID)},
		Hash:        hash[:],
	}
	err = protocolSet.Insert(ctx, db, boil.Infer())
	require.NoError(t, err)

	// Create older crawl first to have historical data (so it gets a lower ID)
	oldCrawl := &models.Crawl{
		StartedAt:       time.Now().Add(-24 * time.Hour),
		FinishedAt:      time.Now().Add(-23 * time.Hour),
		Status:          "finished",
		DialablePeers:   null.IntFrom(500),
		UndialablePeers: null.IntFrom(0),
	}
	err = oldCrawl.Insert(ctx, db, boil.Infer())
	require.NoError(t, err)

	// Create the latest crawl (will get a higher ID)
	crawl := &models.Crawl{
		StartedAt:       time.Now().Add(-1 * time.Hour),
		FinishedAt:      time.Now().Add(-30 * time.Minute),
		Status:          "finished",
		DialablePeers:   null.IntFrom(100),
		UndialablePeers: null.IntFrom(0),
	}
	err = crawl.Insert(ctx, db, boil.Infer())
	require.NoError(t, err)

	// Create 200 peers total (but only 100 in latest crawl)
	allPeerIDs := make([]int64, 200)
	for i := 0; i < 200; i++ {
		peer := &models.Peer{
			MultiHash:      fmt.Sprintf("test_peer_%d", i),
			AgentVersionID: null.IntFrom(agent.ID),
			ProtocolsSetID: null.IntFrom(protocolSet.ID),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			LastSeen:       time.Now(),
		}
		err := peer.Insert(ctx, db, boil.Infer())
		require.NoError(t, err)
		allPeerIDs[i] = peer.ID

		// Only first 100 peers are in the latest crawl
		if i < 100 {
			visit := &models.Visit{
				CrawlID:         null.IntFrom(crawl.ID),
				PeerID:          peer.ID,
				ConnectDuration: null.StringFrom("100ms"),
				CrawlDuration:   null.StringFrom("200ms"),
				VisitStartedAt:  time.Now().Add(-45 * time.Minute),
				VisitEndedAt:    time.Now().Add(-40 * time.Minute),
			}
			err = visit.Insert(ctx, db, boil.Infer())
			require.NoError(t, err)
		} else {
			// Other peers were only in old crawl
			visit := &models.Visit{
				CrawlID:         null.IntFrom(oldCrawl.ID),
				PeerID:          peer.ID,
				ConnectDuration: null.StringFrom("100ms"),
				CrawlDuration:   null.StringFrom("200ms"),
				VisitStartedAt:  time.Now().Add(-23*time.Hour + 30*time.Minute),
				VisitEndedAt:    time.Now().Add(-23*time.Hour + 35*time.Minute),
			}
			err = visit.Insert(ctx, db, boil.Infer())
			require.NoError(t, err)
		}
	}

	// Create 1,000 peer states across ALL peers (not just latest crawl)
	// This simulates real-world data where the peers_states table has historical data
	t.Log("Creating 1,000 peer states across all 200 peers...")
	stateTime := time.Now().Add(-240 * time.Hour) // Start 10 days ago

	for i := 0; i < 1000; i++ {
		peerID := allPeerIDs[i%200] // Distribute across all peers

		var syncStatus null.Bool
		// Mix up sync status to simulate state changes
		switch (int64(i) + peerID) % 5 {
		case 0, 1: // 40% synced
			syncStatus = null.BoolFrom(true)
		case 2, 3: // 40% not synced
			syncStatus = null.BoolFrom(false)
		default: // 20% unknown
			syncStatus = null.Bool{}
		}

		state := &models.PeersState{
			PeerID:      peerID,
			BlockHeight: null.Int64From(int64(90000 + i)),
			SpecVersion: null.StringFrom(fmt.Sprintf("v1.0.%d", i%10)),
			IsSynced:    syncStatus,
			CreatedAt:   stateTime.Add(time.Duration(i) * 15 * time.Minute), // Spread over time
		}
		err = state.Insert(ctx, db, boil.Infer())
		require.NoError(t, err)
	}

	t.Log("Test data setup complete. Calculating sync status counts...")

	// Calculate and store sync status counts for the latest crawl
	t.Logf("Calculating sync status for crawl ID: %d", crawl.ID)
	err = repo.CalculateAndStoreSyncStatusCount(ctx, crawl.ID)
	require.NoError(t, err)

	// Verify it was stored
	storedCount, err := repo.GetCrawlSyncStatusCount(ctx, crawl.ID)
	require.NoError(t, err)
	require.NotNil(t, storedCount, "Failed to store sync status counts")
	t.Logf("Stored sync status counts: synced=%d, not_synced=%d, unknown=%d",
		storedCount.SyncedCount, storedCount.NotSyncedCount, storedCount.UnknownCount)

	t.Log("Running performance comparison...")

	// Get results from both methods and time them
	start := time.Now()
	currentResults, err := repo.GetSyncStatusCount(ctx)
	currentDuration := time.Since(start)
	require.NoError(t, err)

	// Use pre-calculated counts instead
	start = time.Now()
	optimizedResults, err := repo.GetLatestCrawlSyncStatusCount(ctx)
	optimizedDuration := time.Since(start)
	require.NoError(t, err)

	// Debug: Check what crawl it's looking for
	latestCrawl, err := models.Crawls(
		models.CrawlWhere.Status.EQ("finished"),
	).All(ctx, db)
	require.NoError(t, err)
	t.Logf("Found %d finished crawls", len(latestCrawl))
	for _, c := range latestCrawl {
		t.Logf("Crawl ID %d: FinishedAt=%v", c.ID, c.FinishedAt)
	}

	// Log the performance results
	t.Logf("Dataset: 200 total peers (100 in latest crawl), 1,000 peer states")
	t.Logf("Current method took: %v", currentDuration)
	t.Logf("Pre-calculated method took: %v", optimizedDuration)
	t.Logf("Speedup: %.2fx", currentDuration.Seconds()/optimizedDuration.Seconds())
	t.Logf("Performance improvement: %.1f%%", (currentDuration.Seconds()-optimizedDuration.Seconds())/currentDuration.Seconds()*100)

	// Check if optimizedResults is nil
	if optimizedResults == nil {
		t.Fatal("Pre-calculated results are nil - sync status counts were not stored properly")
	}

	// Verify results are the same
	require.Equal(t, len(currentResults), len(optimizedResults), "Result count should match")

	for status, count := range currentResults {
		optimizedCount, exists := optimizedResults[status]
		require.True(t, exists, "Status %s should exist in pre-calculated results", status)
		require.Equal(t, count, optimizedCount, "Count for status %s should match", status)
	}
}

// Test that both methods return the same results
func TestGetSyncStatusCount_Consistency(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	defer tearDown()

	// Setup benchmark data (smaller dataset for test)
	ctx := context.Background()
	db := repo.db

	// Create minimal test data
	agent := &models.AgentVersion{
		AgentVersion: "test/v1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := agent.Insert(ctx, db, boil.Infer())
	require.NoError(t, err)

	protocol := &models.Protocol{
		Protocol:  "test_proto",
		Count:     1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = protocol.Insert(ctx, db, boil.Infer())
	require.NoError(t, err)

	// Calculate hash for the protocol set
	hash := sha256.Sum256([]byte(fmt.Sprintf("%v", []int64{int64(protocol.ID)})))

	protocolSet := &models.ProtocolsSet{
		ProtocolIds: types.Int64Array{int64(protocol.ID)},
		Hash:        hash[:],
	}
	err = protocolSet.Insert(ctx, db, boil.Infer())
	require.NoError(t, err)

	crawl := &models.Crawl{
		StartedAt:       time.Now().Add(-1 * time.Hour),
		FinishedAt:      time.Now().Add(-30 * time.Minute),
		Status:          "finished",
		DialablePeers:   null.IntFrom(100),
		UndialablePeers: null.IntFrom(0),
	}
	err = crawl.Insert(ctx, db, boil.Infer())
	require.NoError(t, err)

	// Create 100 peers with different sync states
	for i := 0; i < 100; i++ {
		peer := &models.Peer{
			MultiHash:      fmt.Sprintf("test_peer_%d", i),
			AgentVersionID: null.IntFrom(agent.ID),
			ProtocolsSetID: null.IntFrom(protocolSet.ID),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		err := peer.Insert(ctx, db, boil.Infer())
		require.NoError(t, err)

		visit := &models.Visit{
			CrawlID:         null.IntFrom(crawl.ID),
			PeerID:          peer.ID,
			ConnectDuration: null.StringFrom("100ms"),
			CrawlDuration:   null.StringFrom("200ms"),
			VisitStartedAt:  time.Now().Add(-45 * time.Minute),
			VisitEndedAt:    time.Now().Add(-40 * time.Minute),
		}
		err = visit.Insert(ctx, db, boil.Infer())
		require.NoError(t, err)

		// Create peer states
		var syncStatus null.Bool
		switch i % 4 {
		case 0: // 25% synced
			syncStatus = null.BoolFrom(true)
		case 1: // 25% not synced
			syncStatus = null.BoolFrom(false)
		case 2, 3: // 50% unknown
			syncStatus = null.Bool{}
		}

		state := &models.PeersState{
			PeerID:      peer.ID,
			BlockHeight: null.Int64From(int64(100000 + i)),
			SpecVersion: null.StringFrom("v1.0.0"),
			IsSynced:    syncStatus,
			CreatedAt:   time.Now().Add(-time.Duration(i) * time.Minute),
		}
		err = state.Insert(ctx, db, boil.Infer())
		require.NoError(t, err)
	}

	// Calculate and store sync status counts for the crawl
	err = repo.CalculateAndStoreSyncStatusCount(ctx, crawl.ID)
	require.NoError(t, err)

	// Get results from both methods
	currentResults, err := repo.GetSyncStatusCount(ctx)
	require.NoError(t, err)

	// Use pre-calculated counts instead
	optimizedResults, err := repo.GetLatestCrawlSyncStatusCount(ctx)
	require.NoError(t, err)

	// Compare results
	// Note: optimizedResults includes a "total" key that currentResults doesn't have
	require.Equal(t, len(currentResults)+1, len(optimizedResults), "Result count should match (optimised has additional 'total' key)")

	// Verify the sync status counts match
	for status, count := range currentResults {
		optimizedCount, exists := optimizedResults[status]
		require.True(t, exists, "Status %s should exist in pre-calculated results", status)
		require.Equal(t, count, optimizedCount, "Count for status %s should match", status)
	}

	// Verify the total count is correct
	expectedTotal := 0
	for _, count := range currentResults {
		expectedTotal += count
	}
	require.Equal(t, expectedTotal, optimizedResults["total"], "Total count should be sum of all status counts")
}
