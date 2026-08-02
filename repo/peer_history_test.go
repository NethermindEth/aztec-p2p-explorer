package repo

import (
	"context"
	"testing"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func TestUpsertDailyPeerHistory(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()
	testDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("insert new record", func(t *testing.T) {
		err := repo.UpsertDailyPeerHistory(ctx, testDate, 100)
		require.NoError(t, err)

		// Verify the record was created using GetDailyPeerHistory
		start := testDate
		end := testDate.AddDate(0, 0, 1)
		history, err := repo.GetDailyPeerHistory(ctx, start, end)
		require.NoError(t, err)
		require.Len(t, history, 1)
		assert.Equal(t, testDate.Format("2006-01-02"), history[0].Date.Format("2006-01-02"))
		assert.Equal(t, 100, history[0].PeerCount)
	})

	t.Run("update with higher count", func(t *testing.T) {
		// Upsert with higher count
		err := repo.UpsertDailyPeerHistory(ctx, testDate, 150)
		require.NoError(t, err)

		// Verify the count was updated
		start := testDate
		end := testDate.AddDate(0, 0, 1)
		history, err := repo.GetDailyPeerHistory(ctx, start, end)
		require.NoError(t, err)
		require.Len(t, history, 1)
		assert.Equal(t, 150, history[0].PeerCount)
	})

	t.Run("update with lower count keeps higher", func(t *testing.T) {
		// Try to upsert with lower count
		err := repo.UpsertDailyPeerHistory(ctx, testDate, 120)
		require.NoError(t, err)

		// Verify the count remains at the higher value
		start := testDate
		end := testDate.AddDate(0, 0, 1)
		history, err := repo.GetDailyPeerHistory(ctx, start, end)
		require.NoError(t, err)
		require.Len(t, history, 1)
		assert.Equal(t, 150, history[0].PeerCount, "should keep the maximum value")
	})

	t.Run("multiple dates", func(t *testing.T) {
		date2 := time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC)
		date3 := time.Date(2025, 1, 17, 0, 0, 0, 0, time.UTC)

		err := repo.UpsertDailyPeerHistory(ctx, date2, 200)
		require.NoError(t, err)

		err = repo.UpsertDailyPeerHistory(ctx, date3, 250)
		require.NoError(t, err)

		// Verify all three records exist (testDate, date2, date3)
		start := testDate
		end := date3.AddDate(0, 0, 1)
		history, err := repo.GetDailyPeerHistory(ctx, start, end)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(history), 3)

		// Verify the values
		assert.Equal(t, 150, history[0].PeerCount) // testDate
		assert.Equal(t, 200, history[1].PeerCount) // date2
		assert.Equal(t, 250, history[2].PeerCount) // date3
	})
}

func TestGetDailyPeerHistory(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	// Insert test data
	baseDate := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 10; i++ {
		date := baseDate.AddDate(0, 0, i)
		peerCount := 100 + i*10
		err := repo.UpsertDailyPeerHistory(ctx, date, peerCount)
		require.NoError(t, err)
	}

	t.Run("get all records in range", func(t *testing.T) {
		start := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
		end := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)

		history, err := repo.GetDailyPeerHistory(ctx, start, end)
		require.NoError(t, err)
		assert.Len(t, history, 10)

		// Verify ordering (ascending by date)
		for i := 0; i < len(history)-1; i++ {
			assert.True(t, history[i].Date.Before(history[i+1].Date))
		}

		// Verify first and last entries
		assert.Equal(t, start.Format("2006-01-02"), history[0].Date.Format("2006-01-02"))
		assert.Equal(t, 100, history[0].PeerCount)
		assert.Equal(t, 190, history[9].PeerCount)
	})

	t.Run("get partial range", func(t *testing.T) {
		start := time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC)
		end := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

		history, err := repo.GetDailyPeerHistory(ctx, start, end)
		require.NoError(t, err)
		assert.Len(t, history, 3) // 12, 13, 14 (end is exclusive)
	})

	t.Run("zero end date uses current time", func(t *testing.T) {
		start := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
		end := time.Time{} // zero value

		history, err := repo.GetDailyPeerHistory(ctx, start, end)
		require.NoError(t, err)
		assert.NotEmpty(t, history)
	})

	t.Run("start after end returns error", func(t *testing.T) {
		start := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
		end := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)

		history, err := repo.GetDailyPeerHistory(ctx, start, end)
		require.Error(t, err)
		assert.Nil(t, history)
		assert.Contains(t, err.Error(), "start date must be before end date")
	})

	t.Run("empty result for future range", func(t *testing.T) {
		start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)

		history, err := repo.GetDailyPeerHistory(ctx, start, end)
		require.NoError(t, err)
		assert.Empty(t, history)
	})
}

func TestCalculateAndStorePeerHistory(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	t.Run("successful calculation and storage", func(t *testing.T) {
		// Create a crawl
		crawl := &coretypes.Crawl{
			State:           "completed",
			StartedAt:       time.Now().UTC().Add(-1 * time.Hour),
			FinishedAt:      time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
			CrawledPeers:    100,
			DialablePeers:   80,
			UndialablePeers: 20,
			RemainingPeers:  0,
			Version:         "v1.0.0",
		}
		crawlID, err := repo.CreateCrawl(ctx, crawl, coretypes.CrawlStatusFinished)
		require.NoError(t, err)

		// Create some test peers and visits in peer_visits_index
		// Note: In a real scenario, this would be populated by the processor
		// For testing, we'll insert test data directly
		for i := 0; i < 5; i++ {
			peerHash := "peer_" + string(rune('A'+i))

			// Create peer first
			peer := &models.Peer{
				MultiHash: peerHash,
				CreatedAt: time.Now(),
				LastSeen:  time.Now(),
			}
			err := peer.Insert(ctx, repo.db, boil.Infer())
			require.NoError(t, err)

			// Now insert into peer_visits_index
			_, err = repo.db.ExecContext(ctx, `
				INSERT INTO peer_visits_index
				(crawl_id, visit_id, peer_id, peer_multi_hash, created_at, last_seen)
				VALUES ($1, $2, $3, $4, NOW(), NOW())
			`, crawlID, i+1, peer.ID, peerHash)
			require.NoError(t, err)
		}

		// Call the function under test
		err = repo.CalculateAndStorePeerHistory(ctx, crawlID)
		require.NoError(t, err)

		// Verify crawls.ccv_filtered_peer_count was updated
		dbCrawl, err := models.Crawls(models.CrawlWhere.ID.EQ(crawlID)).One(ctx, repo.db)
		require.NoError(t, err)
		assert.True(t, dbCrawl.CCVFilteredPeerCount.Valid)
		assert.Equal(t, 5, dbCrawl.CCVFilteredPeerCount.Int)

		// Verify daily_peer_history was updated
		expectedDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		start := expectedDate
		end := expectedDate.AddDate(0, 0, 1)
		history, err := repo.GetDailyPeerHistory(ctx, start, end)
		require.NoError(t, err)
		require.Len(t, history, 1)
		assert.Equal(t, 5, history[0].PeerCount)
	})

	t.Run("multiple crawls same day takes max", func(t *testing.T) {
		testDate := time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC)

		// Create first crawl with 3 peers
		crawl1 := &coretypes.Crawl{
			State:           "completed",
			StartedAt:       testDate.Add(-1 * time.Hour),
			FinishedAt:      testDate.Add(1 * time.Hour),
			CrawledPeers:    3,
			DialablePeers:   3,
			UndialablePeers: 0,
			RemainingPeers:  0,
			Version:         "v1.0.0",
		}
		crawlID1, err := repo.CreateCrawl(ctx, crawl1, coretypes.CrawlStatusFinished)
		require.NoError(t, err)

		// Create peers and visits for crawl 1
		for i := 0; i < 3; i++ {
			peerHash := "crawl1_peer_" + string(rune('A'+i))

			// Create peer first
			peer := &models.Peer{
				MultiHash: peerHash,
				CreatedAt: time.Now(),
				LastSeen:  time.Now(),
			}
			err := peer.Insert(ctx, repo.db, boil.Infer())
			require.NoError(t, err)

			// Now insert into peer_visits_index
			_, err = repo.db.ExecContext(ctx, `
				INSERT INTO peer_visits_index
				(crawl_id, visit_id, peer_id, peer_multi_hash, created_at, last_seen)
				VALUES ($1, $2, $3, $4, NOW(), NOW())
			`, crawlID1, 100+i, peer.ID, peerHash)
			require.NoError(t, err)
		}

		err = repo.CalculateAndStorePeerHistory(ctx, crawlID1)
		require.NoError(t, err)

		// Create second crawl with 7 peers on the same day
		crawl2 := &coretypes.Crawl{
			State:           "completed",
			StartedAt:       testDate.Add(2 * time.Hour),
			FinishedAt:      testDate.Add(3 * time.Hour),
			CrawledPeers:    7,
			DialablePeers:   7,
			UndialablePeers: 0,
			RemainingPeers:  0,
			Version:         "v1.0.0",
		}
		crawlID2, err := repo.CreateCrawl(ctx, crawl2, coretypes.CrawlStatusFinished)
		require.NoError(t, err)

		// Create peers and visits for crawl 2
		for i := 0; i < 7; i++ {
			peerHash := "crawl2_peer_" + string(rune('A'+i))

			// Create peer first
			peer := &models.Peer{
				MultiHash: peerHash,
				CreatedAt: time.Now(),
				LastSeen:  time.Now(),
			}
			err := peer.Insert(ctx, repo.db, boil.Infer())
			require.NoError(t, err)

			// Now insert into peer_visits_index
			_, err = repo.db.ExecContext(ctx, `
				INSERT INTO peer_visits_index
				(crawl_id, visit_id, peer_id, peer_multi_hash, created_at, last_seen)
				VALUES ($1, $2, $3, $4, NOW(), NOW())
			`, crawlID2, 200+i, peer.ID, peerHash)
			require.NoError(t, err)
		}

		err = repo.CalculateAndStorePeerHistory(ctx, crawlID2)
		require.NoError(t, err)

		// Verify daily_peer_history contains the maximum (7)
		start := testDate
		end := testDate.AddDate(0, 0, 1)
		history, err := repo.GetDailyPeerHistory(ctx, start, end)
		require.NoError(t, err)
		require.Len(t, history, 1)
		assert.Equal(t, 7, history[0].PeerCount, "should store the maximum peer count for the day")
	})

	t.Run("crawl not found returns error", func(t *testing.T) {
		err := repo.CalculateAndStorePeerHistory(ctx, 99999)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "crawl 99999 not found")
	})

	t.Run("zero peers", func(t *testing.T) {
		crawl := &coretypes.Crawl{
			State:           "completed",
			StartedAt:       time.Now().UTC().Add(-1 * time.Hour),
			FinishedAt:      time.Date(2025, 1, 17, 12, 0, 0, 0, time.UTC),
			CrawledPeers:    0,
			DialablePeers:   0,
			UndialablePeers: 0,
			RemainingPeers:  0,
			Version:         "v1.0.0",
		}
		crawlID, err := repo.CreateCrawl(ctx, crawl, coretypes.CrawlStatusFinished)
		require.NoError(t, err)

		// No peers added to peer_visits_index

		err = repo.CalculateAndStorePeerHistory(ctx, crawlID)
		require.NoError(t, err)

		// Verify crawls.ccv_filtered_peer_count is 0
		dbCrawl, err := models.Crawls(models.CrawlWhere.ID.EQ(crawlID)).One(ctx, repo.db)
		require.NoError(t, err)
		assert.True(t, dbCrawl.CCVFilteredPeerCount.Valid)
		assert.Equal(t, 0, dbCrawl.CCVFilteredPeerCount.Int)

		// Verify daily_peer_history has 0
		expectedDate := time.Date(2025, 1, 17, 0, 0, 0, 0, time.UTC)
		start := expectedDate
		end := expectedDate.AddDate(0, 0, 1)
		history, err := repo.GetDailyPeerHistory(ctx, start, end)
		require.NoError(t, err)
		require.Len(t, history, 1)
		assert.Equal(t, 0, history[0].PeerCount)
	})
}

func TestUpdateCrawlPeerCount(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	//nolint:dupl // test cases have similar setup by design
	t.Run("update existing crawl", func(t *testing.T) {
		// Create a crawl
		crawl := &models.Crawl{
			StartedAt:       time.Now().UTC().Add(-1 * time.Hour),
			FinishedAt:      time.Now().UTC(),
			Status:          "finished",
			DialablePeers:   null.IntFrom(100),
			UndialablePeers: null.IntFrom(10),
		}
		err := crawl.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Update the peer count
		err = repo.UpdateCrawlPeerCount(ctx, crawl.ID, 150)
		require.NoError(t, err)

		// Verify it was updated
		dbCrawl, err := models.Crawls(models.CrawlWhere.ID.EQ(crawl.ID)).One(ctx, repo.db)
		require.NoError(t, err)
		assert.True(t, dbCrawl.CCVFilteredPeerCount.Valid)
		assert.Equal(t, 150, dbCrawl.CCVFilteredPeerCount.Int)
	})

	t.Run("update non-existent crawl fails", func(t *testing.T) {
		err := repo.UpdateCrawlPeerCount(ctx, 99999, 100)
		require.Error(t, err)
	})

	//nolint:dupl // test cases have similar setup by design
	t.Run("update with zero", func(t *testing.T) {
		crawl := &models.Crawl{
			StartedAt:       time.Now().UTC().Add(-1 * time.Hour),
			FinishedAt:      time.Now().UTC(),
			Status:          "finished",
			DialablePeers:   null.IntFrom(50),
			UndialablePeers: null.IntFrom(5),
		}
		err := crawl.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		err = repo.UpdateCrawlPeerCount(ctx, crawl.ID, 0)
		require.NoError(t, err)

		dbCrawl, err := models.Crawls(models.CrawlWhere.ID.EQ(crawl.ID)).One(ctx, repo.db)
		require.NoError(t, err)
		assert.True(t, dbCrawl.CCVFilteredPeerCount.Valid)
		assert.Equal(t, 0, dbCrawl.CCVFilteredPeerCount.Int)
	})
}

func TestGetDailyPeerCountForDate(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	t.Run("get count from peer_visits_index", func(t *testing.T) {
		testDate := time.Date(2025, 1, 20, 10, 0, 0, 0, time.UTC)

		// Create a crawl for this date
		crawl := &coretypes.Crawl{
			State:           "completed",
			StartedAt:       testDate.Add(-1 * time.Hour),
			FinishedAt:      testDate,
			CrawledPeers:    5,
			DialablePeers:   5,
			UndialablePeers: 0,
			RemainingPeers:  0,
			Version:         "v1.0.0",
		}
		crawlID, err := repo.CreateCrawl(ctx, crawl, coretypes.CrawlStatusFinished)
		require.NoError(t, err)

		// Create peers and visits
		for i := 0; i < 5; i++ {
			peerHash := "pvi_peer_" + string(rune('A'+i))

			// Create peer first
			peer := &models.Peer{
				MultiHash: peerHash,
				CreatedAt: time.Now(),
				LastSeen:  time.Now(),
			}
			err := peer.Insert(ctx, repo.db, boil.Infer())
			require.NoError(t, err)

			// Insert into peer_visits_index
			_, err = repo.db.ExecContext(ctx, `
				INSERT INTO peer_visits_index
				(crawl_id, visit_id, peer_id, peer_multi_hash, created_at, last_seen)
				VALUES ($1, $2, $3, $4, NOW(), NOW())
			`, crawlID, 300+i, peer.ID, peerHash)
			require.NoError(t, err)
		}

		// Get count for this date
		count, source, err := repo.GetDailyPeerCountForDate(ctx, testDate)
		require.NoError(t, err)
		assert.Equal(t, 5, count)
		assert.Equal(t, "peer_visits_index", source)
	})

	t.Run("fallback to crawls.dialable_peers when no peer_visits_index data", func(t *testing.T) {
		testDate := time.Date(2025, 1, 21, 14, 0, 0, 0, time.UTC)

		// Create a crawl with dialable_peers but no peer_visits_index entries
		crawl := &models.Crawl{
			StartedAt:       testDate.Add(-1 * time.Hour),
			FinishedAt:      testDate,
			Status:          "finished",
			DialablePeers:   null.IntFrom(42),
			UndialablePeers: null.IntFrom(8),
		}
		err := crawl.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Don't insert any peer_visits_index entries

		// Get count for this date - should fall back to dialable_peers
		count, source, err := repo.GetDailyPeerCountForDate(ctx, testDate)
		require.NoError(t, err)
		assert.Equal(t, 42, count)
		assert.Equal(t, "crawls.dialable_peers", source)
	})

	t.Run("return zero when no data available", func(t *testing.T) {
		testDate := time.Date(2025, 1, 22, 0, 0, 0, 0, time.UTC)

		// Don't create any crawls for this date

		count, source, err := repo.GetDailyPeerCountForDate(ctx, testDate)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Equal(t, "crawls.dialable_peers", source)
	})

	t.Run("max across multiple crawls same day from peer_visits_index", func(t *testing.T) {
		testDate := time.Date(2025, 1, 23, 0, 0, 0, 0, time.UTC)

		// Create first crawl with 3 peers
		crawl1 := &coretypes.Crawl{
			State:           "completed",
			StartedAt:       testDate.Add(1 * time.Hour),
			FinishedAt:      testDate.Add(2 * time.Hour),
			CrawledPeers:    3,
			DialablePeers:   3,
			UndialablePeers: 0,
			RemainingPeers:  0,
			Version:         "v1.0.0",
		}
		crawlID1, err := repo.CreateCrawl(ctx, crawl1, coretypes.CrawlStatusFinished)
		require.NoError(t, err)

		// Create peers and visits for crawl 1
		for i := 0; i < 3; i++ {
			peerHash := "max_crawl1_peer_" + string(rune('A'+i))
			peer := &models.Peer{
				MultiHash: peerHash,
				CreatedAt: time.Now(),
				LastSeen:  time.Now(),
			}
			err := peer.Insert(ctx, repo.db, boil.Infer())
			require.NoError(t, err)

			_, err = repo.db.ExecContext(ctx, `
				INSERT INTO peer_visits_index
				(crawl_id, visit_id, peer_id, peer_multi_hash, created_at, last_seen)
				VALUES ($1, $2, $3, $4, NOW(), NOW())
			`, crawlID1, 400+i, peer.ID, peerHash)
			require.NoError(t, err)
		}

		// Create second crawl with 8 peers on the same day
		crawl2 := &coretypes.Crawl{
			State:           "completed",
			StartedAt:       testDate.Add(5 * time.Hour),
			FinishedAt:      testDate.Add(6 * time.Hour),
			CrawledPeers:    8,
			DialablePeers:   8,
			UndialablePeers: 0,
			RemainingPeers:  0,
			Version:         "v1.0.0",
		}
		crawlID2, err := repo.CreateCrawl(ctx, crawl2, coretypes.CrawlStatusFinished)
		require.NoError(t, err)

		// Create peers and visits for crawl 2
		for i := 0; i < 8; i++ {
			peerHash := "max_crawl2_peer_" + string(rune('A'+i))
			peer := &models.Peer{
				MultiHash: peerHash,
				CreatedAt: time.Now(),
				LastSeen:  time.Now(),
			}
			err := peer.Insert(ctx, repo.db, boil.Infer())
			require.NoError(t, err)

			_, err = repo.db.ExecContext(ctx, `
				INSERT INTO peer_visits_index
				(crawl_id, visit_id, peer_id, peer_multi_hash, created_at, last_seen)
				VALUES ($1, $2, $3, $4, NOW(), NOW())
			`, crawlID2, 500+i, peer.ID, peerHash)
			require.NoError(t, err)
		}

		// Get count for this date - should return max (8)
		count, source, err := repo.GetDailyPeerCountForDate(ctx, testDate)
		require.NoError(t, err)
		assert.Equal(t, 8, count)
		assert.Equal(t, "peer_visits_index", source)
	})

	t.Run("max across multiple crawls same day from dialable_peers", func(t *testing.T) {
		testDate := time.Date(2025, 1, 24, 0, 0, 0, 0, time.UTC)

		// Create first crawl with dialable_peers = 20
		crawl1 := &models.Crawl{
			StartedAt:       testDate.Add(1 * time.Hour),
			FinishedAt:      testDate.Add(2 * time.Hour),
			Status:          "finished",
			DialablePeers:   null.IntFrom(20),
			UndialablePeers: null.IntFrom(5),
		}
		err := crawl1.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Create second crawl with dialable_peers = 35
		crawl2 := &models.Crawl{
			StartedAt:       testDate.Add(5 * time.Hour),
			FinishedAt:      testDate.Add(6 * time.Hour),
			Status:          "finished",
			DialablePeers:   null.IntFrom(35),
			UndialablePeers: null.IntFrom(10),
		}
		err = crawl2.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Get count for this date - should return max (35)
		count, source, err := repo.GetDailyPeerCountForDate(ctx, testDate)
		require.NoError(t, err)
		assert.Equal(t, 35, count)
		assert.Equal(t, "crawls.dialable_peers", source)
	})

	t.Run("date normalisation", func(t *testing.T) {
		// Create a crawl at a specific time on Jan 25
		testDate := time.Date(2025, 1, 25, 16, 30, 45, 123, time.UTC)

		crawl := &coretypes.Crawl{
			State:           "completed",
			StartedAt:       testDate.Add(-1 * time.Hour),
			FinishedAt:      testDate,
			CrawledPeers:    10,
			DialablePeers:   10,
			UndialablePeers: 0,
			RemainingPeers:  0,
			Version:         "v1.0.0",
		}
		crawlID, err := repo.CreateCrawl(ctx, crawl, coretypes.CrawlStatusFinished)
		require.NoError(t, err)

		// Create peers and visits
		for i := 0; i < 10; i++ {
			peerHash := "norm_peer_" + string(rune('A'+i))
			peer := &models.Peer{
				MultiHash: peerHash,
				CreatedAt: time.Now(),
				LastSeen:  time.Now(),
			}
			err := peer.Insert(ctx, repo.db, boil.Infer())
			require.NoError(t, err)

			_, err = repo.db.ExecContext(ctx, `
				INSERT INTO peer_visits_index
				(crawl_id, visit_id, peer_id, peer_multi_hash, created_at, last_seen)
				VALUES ($1, $2, $3, $4, NOW(), NOW())
			`, crawlID, 600+i, peer.ID, peerHash)
			require.NoError(t, err)
		}

		// Query with different times on same date
		testCases := []time.Time{
			time.Date(2025, 1, 25, 0, 0, 0, 0, time.UTC),      // Midnight
			time.Date(2025, 1, 25, 23, 59, 59, 0, time.UTC),   // End of day
			time.Date(2025, 1, 25, 12, 30, 0, 0, time.UTC),    // Noon
			time.Date(2025, 1, 25, 16, 30, 45, 123, time.UTC), // Exact match
		}

		for _, tc := range testCases {
			count, source, err := repo.GetDailyPeerCountForDate(ctx, tc)
			require.NoError(t, err, "failed for time %v", tc)
			assert.Equal(t, 10, count, "count mismatch for time %v", tc)
			assert.Equal(t, "peer_visits_index", source, "source mismatch for time %v", tc)
		}
	})
}
