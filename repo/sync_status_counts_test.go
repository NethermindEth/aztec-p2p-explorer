package repo

import (
	"context"
	"testing"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func TestCreateCrawlSyncStatusCount(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()
	db := repo.db

	// Create a crawl first
	crawl := &models.Crawl{
		StartedAt:       time.Now().Add(-1 * time.Hour),
		FinishedAt:      time.Now(),
		Status:          "finished",
		DialablePeers:   null.IntFrom(100),
		UndialablePeers: null.IntFrom(10),
	}
	err := crawl.Insert(ctx, db, boil.Infer())
	require.NoError(t, err)

	// Create sync status count
	err = repo.CreateCrawlSyncStatusCount(ctx, crawl.ID, 50, 30, 20, 100)
	require.NoError(t, err)

	// Verify it was created
	count, err := repo.GetCrawlSyncStatusCount(ctx, crawl.ID)
	require.NoError(t, err)
	require.NotNil(t, count)
	assert.Equal(t, crawl.ID, count.CrawlID)
	assert.Equal(t, 50, count.SyncedCount)
	assert.Equal(t, 30, count.NotSyncedCount)
	assert.Equal(t, 20, count.UnknownCount)
	assert.Equal(t, 100, count.TotalCount)
}

func TestGetLatestCrawlSyncStatusCount(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()
	db := repo.db

	// Create multiple crawls
	crawl1 := &models.Crawl{
		StartedAt:       time.Now().Add(-3 * time.Hour),
		FinishedAt:      time.Now().Add(-2 * time.Hour),
		Status:          "finished",
		DialablePeers:   null.IntFrom(100),
		UndialablePeers: null.IntFrom(10),
	}
	err := crawl1.Insert(ctx, db, boil.Infer())
	require.NoError(t, err)

	crawl2 := &models.Crawl{
		StartedAt:       time.Now().Add(-1 * time.Hour),
		FinishedAt:      time.Now(),
		Status:          "finished",
		DialablePeers:   null.IntFrom(110),
		UndialablePeers: null.IntFrom(5),
	}
	err = crawl2.Insert(ctx, db, boil.Infer())
	require.NoError(t, err)

	// Create sync status counts for both crawls
	err = repo.CreateCrawlSyncStatusCount(ctx, crawl1.ID, 40, 35, 25, 100)
	require.NoError(t, err)

	err = repo.CreateCrawlSyncStatusCount(ctx, crawl2.ID, 60, 30, 10, 100)
	require.NoError(t, err)

	// Get latest should return counts from crawl2
	counts, err := repo.GetLatestCrawlSyncStatusCount(ctx)
	require.NoError(t, err)
	require.NotNil(t, counts)
	assert.Equal(t, 60, counts["synced"])
	assert.Equal(t, 30, counts["not_synced"])
	assert.Equal(t, 10, counts["unknown"])
	assert.Equal(t, 100, counts["total"])
}

func TestCalculateAndStoreSyncStatusCount(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	// Use existing test data (setupTestRepo creates some peers and states)
	// Get the crawl that was created in setupTestRepo
	crawls, err := models.Crawls().All(ctx, repo.db)
	require.NoError(t, err)
	require.NotEmpty(t, crawls)

	crawlID := crawls[len(crawls)-1].ID // Get the latest crawl

	// Delete any existing sync status count for this crawl first
	existing, _ := models.CrawlSyncStatusCounts(models.CrawlSyncStatusCountWhere.CrawlID.EQ(crawlID)).One(ctx, repo.db)
	if existing != nil {
		_, err = existing.Delete(ctx, repo.db)
		require.NoError(t, err)
	}

	// Calculate and store sync status counts
	err = repo.CalculateAndStoreSyncStatusCount(ctx, crawlID)
	require.NoError(t, err)

	// Verify counts were stored
	count, err := repo.GetCrawlSyncStatusCount(ctx, crawlID)
	require.NoError(t, err)
	require.NotNil(t, count)

	// The test data should have some synced and unknown peers
	assert.Greater(t, count.SyncedCount, 0)
	assert.Greater(t, count.UnknownCount, 0)
	// Verify that total count matches the sum of individual counts
	expectedTotal := count.SyncedCount + count.NotSyncedCount + count.UnknownCount
	assert.Equal(t, expectedTotal, count.TotalCount)
	assert.Greater(t, count.TotalCount, 0)
}

func TestGetLatestCrawlSyncStatusCount_NoData(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()
	db := repo.db

	// Clear all crawls
	_, err := models.Crawls().DeleteAll(ctx, db)
	require.NoError(t, err)

	// Should return nil when no data exists
	counts, err := repo.GetLatestCrawlSyncStatusCount(ctx)
	require.NoError(t, err)
	assert.Nil(t, counts)
}

// TestGetSyncStatusCountByCrawlID_SmoothedScenarios locks in the seven
// scenarios from the smoothing spec.
//
// Each peer is given a deterministic peers_states history; the test creates a
// crawl that "visits" all of them, then calls GetSyncStatusCountByCrawlID and
// asserts the smoothed weighted contributions per bucket.
func TestGetSyncStatusCountByCrawlID_SmoothedScenarios(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()
	db := repo.db

	// Helper: create a peer with the given multi_hash string and return its DB id.
	createPeer := func(name string) int64 {
		t.Helper()
		now := time.Now()
		p := &models.Peer{
			MultiHash: name,
			CreatedAt: now,
			UpdatedAt: now,
			LastSeen:  now,
		}
		require.NoError(t, p.Insert(ctx, db, boil.Infer()))
		return p.ID
	}

	// Helper: append n peers_states rows for a peer.
	// `states` are ordered oldest -> newest. NULL is represented as nil.
	appendStates := func(peerID int64, states []*bool) {
		t.Helper()
		base := time.Now().Add(-time.Duration(len(states)) * time.Minute)
		for i, s := range states {
			row := &models.PeersState{
				PeerID:    peerID,
				CreatedAt: base.Add(time.Duration(i) * time.Minute),
			}
			if s != nil {
				row.IsSynced = null.BoolFrom(*s)
			}
			require.NoError(t, row.Insert(ctx, db, boil.Infer()))
		}
	}

	// Helper: ptr-to-bool for terse fixture construction.
	yes := func() *bool { v := true; return &v }
	no := func() *bool { v := false; return &v }
	// nil for unknown — used inline as `nil`.

	// Build per-scenario fixtures.
	type scenario struct {
		name          string
		states        []*bool // oldest -> newest
		wantSynced    float64
		wantNotSynced float64
		wantUnknown   float64
	}

	// Within tolerance for float comparisons.
	const eps = 1e-9

	scenarios := []scenario{
		{
			name:          "steady_state_10_synced",
			states:        []*bool{yes(), yes(), yes(), yes(), yes(), yes(), yes(), yes(), yes(), yes()},
			wantSynced:    1.0,
			wantNotSynced: 0.0,
			wantUnknown:   0.0,
		},
		{
			name:          "transient_unknown_glitch_9_synced_1_null",
			states:        []*bool{yes(), yes(), yes(), yes(), yes(), yes(), yes(), yes(), yes(), nil},
			wantSynced:    1.0, // 9/9 — NULL excluded from denominator
			wantNotSynced: 0.0,
			wantUnknown:   0.0,
		},
		{
			name:          "real_not_synced_flip_9_synced_1_not_synced",
			states:        []*bool{yes(), yes(), yes(), yes(), yes(), yes(), yes(), yes(), yes(), no()},
			wantSynced:    9.0 / 10.0,
			wantNotSynced: 1.0 / 10.0,
			wantUnknown:   0.0,
		},
		{
			name:          "all_unknown_data_deficit",
			states:        []*bool{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil},
			wantSynced:    0.0,
			wantNotSynced: 0.0,
			wantUnknown:   1.0,
		},
		{
			name:          "cold_start_3_known",
			states:        []*bool{yes(), yes(), no()},
			wantSynced:    2.0 / 3.0,
			wantNotSynced: 1.0 / 3.0,
			wantUnknown:   0.0,
		},
		{
			// 5 oldest are not_synced; 10 newest are synced. With window=10
			// only the synced ones count.
			name: "history_trim_15_rows_uses_only_newest_10",
			states: []*bool{
				no(), no(), no(), no(), no(),
				yes(), yes(), yes(), yes(), yes(), yes(), yes(), yes(), yes(), yes(),
			},
			wantSynced:    1.0,
			wantNotSynced: 0.0,
			wantUnknown:   0.0,
		},
	}

	// Create a crawl. Status `finished` matches what production writes.
	crawl := &models.Crawl{
		StartedAt:       time.Now().Add(-1 * time.Hour),
		FinishedAt:      time.Now(),
		Status:          "finished",
		DialablePeers:   null.IntFrom(len(scenarios)),
		UndialablePeers: null.IntFrom(0),
	}
	require.NoError(t, crawl.Insert(ctx, db, boil.Infer()))

	// Per-peer fixtures + a visit row connecting the peer to this crawl.
	for _, sc := range scenarios {
		peerID := createPeer(sc.name)
		appendStates(peerID, sc.states)
		v := &models.Visit{
			PeerID:         peerID,
			CrawlID:        null.IntFrom(crawl.ID),
			VisitStartedAt: time.Now().Add(-30 * time.Minute),
			VisitEndedAt:   time.Now().Add(-29 * time.Minute),
		}
		require.NoError(t, v.Insert(ctx, db, boil.Infer()))
	}

	// Act.
	got, err := repo.GetSyncStatusCountByCrawlID(ctx, crawl.ID)
	require.NoError(t, err)
	require.NotNil(t, got)

	// Assert the aggregate sum equals the in-crawl peer count
	// (invariant #7 from the spec).
	wantPeerCount := float64(len(scenarios))
	assert.InDelta(t, wantPeerCount, got["total"], eps,
		"sum across all buckets must equal in-crawl peer count")
	assert.InDelta(t, wantPeerCount, got["synced"]+got["not_synced"]+got["unknown"], eps,
		"sum across synced/not_synced/unknown must equal total")

	// Sanity sums: per-scenario expected contributions roll up to the bucket totals.
	var expSynced, expNotSynced, expUnknown float64
	for _, sc := range scenarios {
		expSynced += sc.wantSynced
		expNotSynced += sc.wantNotSynced
		expUnknown += sc.wantUnknown
	}
	assert.InDelta(t, expSynced, got["synced"], eps, "synced bucket")
	assert.InDelta(t, expNotSynced, got["not_synced"], eps, "not_synced bucket")
	assert.InDelta(t, expUnknown, got["unknown"], eps, "unknown bucket")
}
