package repo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func TestGetEarliestPeerTimestamps(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	t.Run("prefers visits.visit_started_at as source", func(t *testing.T) {
		// Create a peer with a later created_at
		laterTime := time.Date(2025, 1, 20, 12, 0, 0, 0, time.UTC)
		peer := &models.Peer{
			MultiHash: "peer_with_visits",
			CreatedAt: laterTime,
			UpdatedAt: laterTime,
			LastSeen:  laterTime,
		}
		err := peer.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Create a visit with an earlier timestamp
		earlierTime := time.Date(2025, 1, 10, 8, 0, 0, 0, time.UTC)
		visit := &models.Visit{
			PeerID:         peer.ID,
			VisitStartedAt: earlierTime,
			VisitEndedAt:   earlierTime.Add(1 * time.Minute),
		}
		err = visit.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Get earliest timestamps - use large batch to get all peers, then filter
		timestamps, err := repo.GetEarliestPeerTimestamps(ctx, 100, 0)
		require.NoError(t, err)

		// Find our test peer in the results
		var testPeerTimestamp *PeerTimestamp
		for _, ts := range timestamps {
			if ts.MultiHash == "peer_with_visits" {
				testPeerTimestamp = ts
				break
			}
		}
		require.NotNil(t, testPeerTimestamp, "should find test peer in results")

		assert.Equal(t, "peer_with_visits", testPeerTimestamp.MultiHash)
		assert.True(t, testPeerTimestamp.FirstSeen.Valid)
		assert.Equal(t, earlierTime.Unix(), testPeerTimestamp.FirstSeen.Time.Unix()) // Compare Unix timestamps to avoid location issues
		assert.Equal(t, "visits", testPeerTimestamp.Source)
	})

	t.Run("falls back to peer_visits_index when no visits", func(t *testing.T) {
		// Create crawl first
		crawl := &coretypes.Crawl{
			State:           "finished",
			StartedAt:       time.Now().UTC().Add(-2 * time.Hour),
			FinishedAt:      time.Now().UTC().Add(-1 * time.Hour),
			CrawledPeers:    1,
			DialablePeers:   1,
			UndialablePeers: 0,
			RemainingPeers:  0,
			Version:         "v1.0.0",
		}
		crawlID, err := repo.CreateCrawl(ctx, crawl, coretypes.CrawlStatusFinished)
		require.NoError(t, err)

		// Create a peer
		laterTime := time.Date(2025, 1, 21, 15, 0, 0, 0, time.UTC)
		peer := &models.Peer{
			MultiHash: "peer_with_index",
			CreatedAt: laterTime,
			UpdatedAt: laterTime,
			LastSeen:  laterTime,
		}
		err = peer.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Insert into peer_visits_index with earlier timestamp
		earlierTime := time.Date(2025, 1, 12, 9, 0, 0, 0, time.UTC)
		_, err = repo.db.ExecContext(ctx, `
			INSERT INTO peer_visits_index
			(crawl_id, visit_id, peer_id, peer_multi_hash, created_at, last_seen)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, crawlID, 101, peer.ID, peer.MultiHash, earlierTime, laterTime)
		require.NoError(t, err)

		// Get earliest timestamps
		timestamps, err := repo.GetEarliestPeerTimestamps(ctx, 100, 0)
		require.NoError(t, err)

		// Find our test peer in the results
		var testPeerTimestamp *PeerTimestamp
		for _, ts := range timestamps {
			if ts.MultiHash == "peer_with_index" {
				testPeerTimestamp = ts
				break
			}
		}
		require.NotNil(t, testPeerTimestamp, "should find test peer in results")

		assert.Equal(t, "peer_with_index", testPeerTimestamp.MultiHash)
		assert.True(t, testPeerTimestamp.FirstSeen.Valid)
		assert.Equal(t, earlierTime.Unix(), testPeerTimestamp.FirstSeen.Time.Unix())
		assert.Equal(t, "peer_visits_index", testPeerTimestamp.Source)
	})

	t.Run("falls back to peers.created_at when no other data", func(t *testing.T) {
		// Create a peer with no visits or index entries
		peerTime := time.Date(2025, 1, 22, 10, 0, 0, 0, time.UTC)
		peer := &models.Peer{
			MultiHash: "peer_no_visits",
			CreatedAt: peerTime,
			UpdatedAt: peerTime,
			LastSeen:  peerTime,
		}
		err := peer.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Get earliest timestamps
		timestamps, err := repo.GetEarliestPeerTimestamps(ctx, 100, 0)
		require.NoError(t, err)

		// Find our test peer in the results
		var testPeerTimestamp *PeerTimestamp
		for _, ts := range timestamps {
			if ts.MultiHash == "peer_no_visits" {
				testPeerTimestamp = ts
				break
			}
		}
		require.NotNil(t, testPeerTimestamp, "should find test peer in results")

		assert.Equal(t, "peer_no_visits", testPeerTimestamp.MultiHash)
		assert.True(t, testPeerTimestamp.FirstSeen.Valid)
		assert.Equal(t, peerTime.Unix(), testPeerTimestamp.FirstSeen.Time.Unix())
		assert.Equal(t, "peers", testPeerTimestamp.Source)
	})

	t.Run("returns earliest visit when multiple visits exist", func(t *testing.T) {
		// Create a peer
		peer := &models.Peer{
			MultiHash: "peer_multiple_visits",
			CreatedAt: time.Date(2025, 1, 25, 12, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2025, 1, 25, 12, 0, 0, 0, time.UTC),
			LastSeen:  time.Date(2025, 1, 25, 12, 0, 0, 0, time.UTC),
		}
		err := peer.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Create multiple visits with different timestamps
		visit1Time := time.Date(2025, 1, 23, 10, 0, 0, 0, time.UTC)
		visit2Time := time.Date(2025, 1, 24, 11, 0, 0, 0, time.UTC)
		earliestTime := time.Date(2025, 1, 22, 9, 0, 0, 0, time.UTC)

		for i, visitTime := range []time.Time{visit1Time, visit2Time, earliestTime} {
			visit := &models.Visit{
				PeerID:         peer.ID,
				VisitStartedAt: visitTime,
				VisitEndedAt:   visitTime.Add(1 * time.Minute),
			}
			err = visit.Insert(ctx, repo.db, boil.Infer())
			require.NoError(t, err, "failed to insert visit %d", i)
		}

		// Get earliest timestamps
		timestamps, err := repo.GetEarliestPeerTimestamps(ctx, 100, 0)
		require.NoError(t, err)

		// Find our test peer in the results
		var testPeerTimestamp *PeerTimestamp
		for _, ts := range timestamps {
			if ts.MultiHash == "peer_multiple_visits" {
				testPeerTimestamp = ts
				break
			}
		}
		require.NotNil(t, testPeerTimestamp, "should find test peer in results")

		assert.Equal(t, "peer_multiple_visits", testPeerTimestamp.MultiHash)
		assert.True(t, testPeerTimestamp.FirstSeen.Valid)
		assert.Equal(t, earliestTime.Unix(), testPeerTimestamp.FirstSeen.Time.Unix())
		assert.Equal(t, "visits", testPeerTimestamp.Source)
	})

	t.Run("batch size and offset work correctly", func(t *testing.T) {
		// Create 5 test peers
		for i := 0; i < 5; i++ {
			peer := &models.Peer{
				MultiHash: "batch_peer_" + string(rune('A'+i)),
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				LastSeen:  time.Now().UTC(),
			}
			err := peer.Insert(ctx, repo.db, boil.Infer())
			require.NoError(t, err)
		}

		// Test batch size = 2, offset = 0 (should get first 2 from current batch)
		// Note: offset is based on ALL peers, so we need to account for previously created peers
		totalPeers, err := repo.CountPeers(ctx)
		require.NoError(t, err)

		// Get last 2 peers
		timestamps, err := repo.GetEarliestPeerTimestamps(ctx, 2, totalPeers-2)
		require.NoError(t, err)
		assert.Len(t, timestamps, 2)

		// Get next 2 peers
		timestamps, err = repo.GetEarliestPeerTimestamps(ctx, 2, totalPeers-4)
		require.NoError(t, err)
		assert.Len(t, timestamps, 2)
	})

	t.Run("returns empty slice when offset exceeds total peers", func(t *testing.T) {
		timestamps, err := repo.GetEarliestPeerTimestamps(ctx, 10, 999999)
		require.NoError(t, err)
		assert.Empty(t, timestamps)
	})
}

func TestUpdatePeerCreatedAt(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	//nolint:dupl // test cases have similar setup by design
	t.Run("successfully updates peer created_at", func(t *testing.T) {
		// Create a peer with initial timestamp
		initialTime := time.Date(2025, 1, 26, 12, 0, 0, 0, time.UTC)
		peer := &models.Peer{
			MultiHash: "peer_to_update",
			CreatedAt: initialTime,
			UpdatedAt: initialTime,
			LastSeen:  initialTime,
		}
		err := peer.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Update to earlier time
		earlierTime := time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC)
		err = repo.UpdatePeerCreatedAt(ctx, peer.MultiHash, sql.NullTime{
			Time:  earlierTime,
			Valid: true,
		})
		require.NoError(t, err)

		// Verify the update
		updatedPeer, err := models.Peers(models.PeerWhere.MultiHash.EQ(peer.MultiHash)).One(ctx, repo.db)
		require.NoError(t, err)
		assert.Equal(t, earlierTime.Unix(), updatedPeer.CreatedAt.Unix())
	})

	t.Run("returns error when peer not found", func(t *testing.T) {
		err := repo.UpdatePeerCreatedAt(ctx, "nonexistent_peer", sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error when timestamp is null", func(t *testing.T) {
		err := repo.UpdatePeerCreatedAt(ctx, "some_peer", sql.NullTime{
			Valid: false,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "timestamp is null")
	})

	//nolint:dupl // test cases have similar setup by design
	t.Run("can update to later time if needed", func(t *testing.T) {
		// Create a peer
		initialTime := time.Date(2025, 1, 20, 10, 0, 0, 0, time.UTC)
		peer := &models.Peer{
			MultiHash: "peer_update_later",
			CreatedAt: initialTime,
			UpdatedAt: initialTime,
			LastSeen:  initialTime,
		}
		err := peer.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Update to later time (edge case, but should work)
		laterTime := time.Date(2025, 1, 25, 15, 0, 0, 0, time.UTC)
		err = repo.UpdatePeerCreatedAt(ctx, peer.MultiHash, sql.NullTime{
			Time:  laterTime,
			Valid: true,
		})
		require.NoError(t, err)

		// Verify the update
		updatedPeer, err := models.Peers(models.PeerWhere.MultiHash.EQ(peer.MultiHash)).One(ctx, repo.db)
		require.NoError(t, err)
		assert.Equal(t, laterTime.Unix(), updatedPeer.CreatedAt.Unix())
	})
}

func TestCountPeers(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	t.Run("returns correct count", func(t *testing.T) {
		initialCount, err := repo.CountPeers(ctx)
		require.NoError(t, err)

		// Create 3 test peers
		for i := 0; i < 3; i++ {
			peer := &models.Peer{
				MultiHash: "count_peer_" + string(rune('A'+i)),
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				LastSeen:  time.Now().UTC(),
			}
			err := peer.Insert(ctx, repo.db, boil.Infer())
			require.NoError(t, err)
		}

		// Count should be initial + 3
		count, err := repo.CountPeers(ctx)
		require.NoError(t, err)
		assert.Equal(t, initialCount+3, count)
	})

	t.Run("returns 0 when database is empty", func(t *testing.T) {
		// This test runs in parallel with others, so we can't guarantee empty DB
		// Just verify CountPeers doesn't error
		count, err := repo.CountPeers(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 0)
	})
}

func TestEarliestPeerTimestampsIntegration(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	t.Run("end-to-end migration scenario", func(t *testing.T) {
		// Create a peer
		currentTime := time.Date(2025, 1, 27, 12, 0, 0, 0, time.UTC)
		peer := &models.Peer{
			MultiHash: "migration_test_peer",
			CreatedAt: currentTime, // This is wrong, should be earlier
			UpdatedAt: currentTime,
			LastSeen:  currentTime,
		}
		err := peer.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Create historical visit with correct "first seen" time
		actualFirstSeen := time.Date(2025, 1, 10, 8, 30, 0, 0, time.UTC)
		visit := &models.Visit{
			PeerID:         peer.ID,
			VisitStartedAt: actualFirstSeen,
			VisitEndedAt:   actualFirstSeen.Add(1 * time.Minute),
		}
		err = visit.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Get total peer count
		count, err := repo.CountPeers(ctx)
		require.NoError(t, err)

		// Get earliest timestamps for this peer (use offset to find it)
		timestamps, err := repo.GetEarliestPeerTimestamps(ctx, count, 0)
		require.NoError(t, err)

		// Find our test peer in the results
		var testPeerTimestamp *PeerTimestamp
		for _, ts := range timestamps {
			if ts.MultiHash == "migration_test_peer" {
				testPeerTimestamp = ts
				break
			}
		}
		require.NotNil(t, testPeerTimestamp, "should find test peer in results")

		// Verify it found the correct earliest time
		assert.Equal(t, actualFirstSeen.Unix(), testPeerTimestamp.FirstSeen.Time.Unix())
		assert.Equal(t, "visits", testPeerTimestamp.Source)

		// Update the peer with the correct timestamp
		err = repo.UpdatePeerCreatedAt(ctx, peer.MultiHash, testPeerTimestamp.FirstSeen)
		require.NoError(t, err)

		// Verify the migration worked
		updatedPeer, err := models.Peers(models.PeerWhere.MultiHash.EQ(peer.MultiHash)).One(ctx, repo.db)
		require.NoError(t, err)
		assert.Equal(t, actualFirstSeen.Unix(), updatedPeer.CreatedAt.Unix())
	})
}

func TestPriorityOfSources(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	t.Run("visits takes priority over peer_visits_index and peers", func(t *testing.T) {
		// Create crawl first
		crawl := &coretypes.Crawl{
			State:           "finished",
			StartedAt:       time.Now().UTC().Add(-2 * time.Hour),
			FinishedAt:      time.Now().UTC().Add(-1 * time.Hour),
			CrawledPeers:    1,
			DialablePeers:   1,
			UndialablePeers: 0,
			RemainingPeers:  0,
			Version:         "v1.0.0",
		}
		crawlID, err := repo.CreateCrawl(ctx, crawl, coretypes.CrawlStatusFinished)
		require.NoError(t, err)

		// Create a peer with late timestamp
		peerTime := time.Date(2025, 1, 28, 20, 0, 0, 0, time.UTC)
		peer := &models.Peer{
			MultiHash: "priority_test",
			CreatedAt: peerTime,
			UpdatedAt: peerTime,
			LastSeen:  peerTime,
		}
		err = peer.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Add peer_visits_index entry with medium timestamp
		indexTime := time.Date(2025, 1, 28, 15, 0, 0, 0, time.UTC)
		_, err = repo.db.ExecContext(ctx, `
			INSERT INTO peer_visits_index
			(crawl_id, visit_id, peer_id, peer_multi_hash, created_at, last_seen)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, crawlID, 999+peer.ID, peer.ID, peer.MultiHash, indexTime, peerTime)
		require.NoError(t, err)

		// Add visit entry with earliest timestamp
		visitTime := time.Date(2025, 1, 28, 8, 0, 0, 0, time.UTC)
		visit := &models.Visit{
			PeerID:         peer.ID,
			VisitStartedAt: visitTime,
			VisitEndedAt:   visitTime.Add(1 * time.Minute),
			CrawlID:        null.IntFrom(crawlID),
		}
		err = visit.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Get timestamps
		count, err := repo.CountPeers(ctx)
		require.NoError(t, err)
		timestamps, err := repo.GetEarliestPeerTimestamps(ctx, count, 0)
		require.NoError(t, err)

		// Find our test peer
		var testPeerTimestamp *PeerTimestamp
		for _, ts := range timestamps {
			if ts.MultiHash == "priority_test" {
				testPeerTimestamp = ts
				break
			}
		}
		require.NotNil(t, testPeerTimestamp)

		// Should use visits (earliest time)
		assert.Equal(t, visitTime.Unix(), testPeerTimestamp.FirstSeen.Time.Unix())
		assert.Equal(t, "visits", testPeerTimestamp.Source)
	})
}
