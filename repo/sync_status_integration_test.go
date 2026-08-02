package repo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

// TestSyncStatusCountIntegration tests the full flow from crawl to API
func TestSyncStatusCountIntegration(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	// Create a crawl
	crawl := &types.Crawl{
		CrawledPeers:    100,
		DialablePeers:   80,
		UndialablePeers: 20,
		StartedAt:       time.Now().Add(-30 * time.Minute),
		FinishedAt:      time.Now(),
		State:           "finished",
		Version:         "v1.0.0",
	}
	crawlID, err := repo.CreateCrawl(ctx, crawl, types.CrawlStatusFinished)
	require.NoError(t, err)

	// Create agent version
	agent := &models.AgentVersion{
		AgentVersion: "test/v1.0.0",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = agent.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	// Create peers with different sync states
	for i := 0; i < 100; i++ {
		// Create peer
		peer := &models.Peer{
			MultiHash:      fmt.Sprintf("peer_%d", i),
			AgentVersionID: null.IntFrom(agent.ID),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			LastSeen:       time.Now(),
		}
		err := peer.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Create visit to associate peer with crawl
		visit := &models.Visit{
			CrawlID:         null.IntFrom(crawlID),
			PeerID:          peer.ID,
			ConnectDuration: null.StringFrom("100ms"),
			CrawlDuration:   null.StringFrom("200ms"),
			VisitStartedAt:  time.Now().Add(-25 * time.Minute),
			VisitEndedAt:    time.Now().Add(-20 * time.Minute),
		}
		err = visit.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Create state with varying sync statuses
		var syncStatus null.Bool
		switch {
		case i < 40:
			// 40% synced
			syncStatus = null.BoolFrom(true)
		case i < 70:
			// 30% not synced
			syncStatus = null.BoolFrom(false)
		default:
			// 30% unknown (null)
			syncStatus = null.Bool{}
		}

		state := &models.PeersState{
			PeerID:      peer.ID,
			BlockHeight: null.Int64From(int64(100000 + i)),
			IsSynced:    syncStatus,
			CreatedAt:   time.Now(),
		}
		err = state.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)
	}

	// Calculate and store sync status counts
	err = repo.CalculateAndStoreSyncStatusCount(ctx, crawlID)
	require.NoError(t, err)

	// Get the counts using the new method
	counts, err := repo.GetLatestCrawlSyncStatusCount(ctx)
	require.NoError(t, err)
	require.NotNil(t, counts)

	// Verify the counts match our test data
	assert.Equal(t, 40, counts["synced"])
	assert.Equal(t, 30, counts["not_synced"])
	assert.Equal(t, 30, counts["unknown"])
	total := counts["synced"] + counts["not_synced"] + counts["unknown"]
	assert.Equal(t, 100, total)

	// Compare with real-time calculation
	realTimeCounts, err := repo.GetSyncStatusCount(ctx)
	require.NoError(t, err)

	// They should match
	assert.Equal(t, realTimeCounts["synced"], counts["synced"])
	assert.Equal(t, realTimeCounts["not_synced"], counts["not_synced"])
	assert.Equal(t, realTimeCounts["unknown"], counts["unknown"])
}
