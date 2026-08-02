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

	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func TestGetPeersOptimizedPagination(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	// Create a finished crawl
	crawl := &models.Crawl{
		Status:    "finished",
		StartedAt: time.Now(),
	}
	err := crawl.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	// Create 10 test peers with different characteristics for pagination testing
	for i := 1; i <= 10; i++ {
		peerHash := fmt.Sprintf("QmTest%02d", i)
		peer := &models.Peer{
			MultiHash: peerHash,
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute),
			LastSeen:  time.Now().Add(time.Duration(i) * time.Minute),
		}
		err = peer.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Create visit for each peer
		visit := &models.Visit{
			PeerID:         peer.ID,
			VisitStartedAt: time.Now(),
			VisitEndedAt:   time.Now(),
			CrawlID:        null.IntFrom(crawl.ID),
		}
		err = visit.Insert(ctx, repo.db, boil.Infer())
		require.NoError(t, err)

		// Insert into index with varying block heights for sorting tests
		blockHeight := int64(i * 1000)
		country := "United States"
		if i%2 == 0 {
			country = "Germany"
		}

		_, err = repo.db.ExecContext(ctx, `
			INSERT INTO peer_visits_index (
				crawl_id, visit_id, peer_id, peer_multi_hash,
				agent_version, client_name,
				country_name, block_height,
				created_at, last_seen
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, crawl.ID, visit.ID, peer.ID, peerHash,
			fmt.Sprintf("agent-v%d", i), fmt.Sprintf("agent%d", i%3),
			country, blockHeight,
			peer.CreatedAt, peer.LastSeen)
		require.NoError(t, err)
	}

	t.Run("Basic pagination - default sort by last_seen desc", func(t *testing.T) {
		// First page
		options := &PeerQueryOptions{
			PageSize: 3,
		}

		peers1, nextToken1, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(10), totalCount)
		assert.Len(t, peers1, 3)
		assert.NotEmpty(t, nextToken1)

		// Verify order - should be newest first (last_seen desc by default)
		// Since last_seen = created_at + i minutes, QmTest10 has the latest last_seen
		assert.Equal(t, "QmTest10", peers1[0].PeerID)
		assert.Equal(t, "QmTest09", peers1[1].PeerID)
		assert.Equal(t, "QmTest08", peers1[2].PeerID)

		// Second page
		options = &PeerQueryOptions{
			PageSize:        3,
			PaginationToken: nextToken1,
		}

		peers2, nextToken2, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(10), totalCount)
		assert.Len(t, peers2, 3)
		assert.NotEmpty(t, nextToken2)

		// Verify no overlap with first page
		peerIDsPage1 := getPeerIDs(peers1)
		peerIDsPage2 := getPeerIDs(peers2)
		for _, id := range peerIDsPage2 {
			assert.NotContains(t, peerIDsPage1, id, "Pages should not overlap")
		}

		// Verify order continues
		assert.Equal(t, "QmTest07", peers2[0].PeerID)
		assert.Equal(t, "QmTest06", peers2[1].PeerID)
		assert.Equal(t, "QmTest05", peers2[2].PeerID)

		// Third page
		options = &PeerQueryOptions{
			PageSize:        3,
			PaginationToken: nextToken2,
		}

		peers3, nextToken3, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(10), totalCount)
		assert.Len(t, peers3, 3)
		assert.NotEmpty(t, nextToken3)

		// Fourth (last) page
		options = &PeerQueryOptions{
			PageSize:        3,
			PaginationToken: nextToken3,
		}

		peers4, nextToken4, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(10), totalCount)
		assert.Len(t, peers4, 1) // Only 1 item left
		assert.Empty(t, nextToken4, "Last page should have no next token")

		assert.Equal(t, "QmTest01", peers4[0].PeerID)
	})

	t.Run("Pagination with block_height sorting", func(t *testing.T) {
		// First page - highest block heights first
		options := &PeerQueryOptions{
			PageSize: 4,
			Sort:     "block_height",
			IsAsc:    false,
		}

		peers1, nextToken1, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(10), totalCount)
		assert.Len(t, peers1, 4)
		assert.NotEmpty(t, nextToken1)

		// Verify descending block height order
		assert.Equal(t, "QmTest10", peers1[0].PeerID) // block_height: 10000
		assert.Equal(t, "QmTest09", peers1[1].PeerID) // block_height: 9000
		assert.Equal(t, "QmTest08", peers1[2].PeerID) // block_height: 8000
		assert.Equal(t, "QmTest07", peers1[3].PeerID) // block_height: 7000

		// Second page
		options = &PeerQueryOptions{
			PageSize:        4,
			Sort:            "block_height",
			IsAsc:           false,
			PaginationToken: nextToken1,
		}

		peers2, nextToken2, _, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Len(t, peers2, 4)
		assert.NotEmpty(t, nextToken2)

		// Verify continued ordering
		assert.Equal(t, "QmTest06", peers2[0].PeerID) // block_height: 6000
		assert.Equal(t, "QmTest05", peers2[1].PeerID) // block_height: 5000
		assert.Equal(t, "QmTest04", peers2[2].PeerID) // block_height: 4000
		assert.Equal(t, "QmTest03", peers2[3].PeerID) // block_height: 3000
	})

	t.Run("Pagination with filter", func(t *testing.T) {
		// Only get peers from Germany (even numbered ones)
		options := &PeerQueryOptions{
			PageSize: 2,
			Filter: &coretypes.PeerFilter{
				Countries: []string{"Germany"},
			},
		}

		peers1, nextToken1, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(5), totalCount) // 5 peers from Germany (2,4,6,8,10)
		assert.Len(t, peers1, 2)
		assert.NotEmpty(t, nextToken1)

		// Should get the newest German peers first
		assert.Equal(t, "QmTest10", peers1[0].PeerID)
		assert.Equal(t, "QmTest08", peers1[1].PeerID)

		// Second page with same filter
		options = &PeerQueryOptions{
			PageSize: 2,
			Filter: &coretypes.PeerFilter{
				Countries: []string{"Germany"},
			},
			PaginationToken: nextToken1,
		}

		peers2, nextToken2, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(5), totalCount)
		assert.Len(t, peers2, 2)
		assert.NotEmpty(t, nextToken2)

		assert.Equal(t, "QmTest06", peers2[0].PeerID)
		assert.Equal(t, "QmTest04", peers2[1].PeerID)

		// Third page
		options.PaginationToken = nextToken2
		peers3, nextToken3, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(5), totalCount)
		assert.Len(t, peers3, 1) // Only 1 left
		assert.Empty(t, nextToken3)

		assert.Equal(t, "QmTest02", peers3[0].PeerID)
	})

	t.Run("Pagination with ascending sort", func(t *testing.T) {
		options := &PeerQueryOptions{
			PageSize: 3,
			Sort:     "created_at",
			IsAsc:    true, // Ascending - oldest first
		}

		peers1, nextToken1, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(10), totalCount)
		assert.Len(t, peers1, 3)

		// Should get oldest peers first
		assert.Equal(t, "QmTest01", peers1[0].PeerID)
		assert.Equal(t, "QmTest02", peers1[1].PeerID)
		assert.Equal(t, "QmTest03", peers1[2].PeerID)

		// Continue pagination
		options.PaginationToken = nextToken1
		peers2, _, _, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Len(t, peers2, 3)

		assert.Equal(t, "QmTest04", peers2[0].PeerID)
		assert.Equal(t, "QmTest05", peers2[1].PeerID)
		assert.Equal(t, "QmTest06", peers2[2].PeerID)
	})

	t.Run("Edge case - page size larger than results", func(t *testing.T) {
		options := &PeerQueryOptions{
			PageSize: 20, // Larger than total count
		}

		peers, nextToken, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(10), totalCount)
		assert.Len(t, peers, 10)
		assert.Empty(t, nextToken, "Should have no next token when all results fit in one page")
	})

	t.Run("Edge case - invalid pagination token", func(t *testing.T) {
		options := &PeerQueryOptions{
			PageSize:        3,
			PaginationToken: "invalid-token",
		}

		_, _, _, err := repo.GetPeersOptimized(ctx, options)
		assert.Error(t, err, "Should error on invalid pagination token")
	})

	t.Run("Consistency - same token returns same results", func(t *testing.T) {
		options := &PeerQueryOptions{
			PageSize: 3,
		}

		_, nextToken, _, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		require.NotEmpty(t, nextToken)

		// Use same token twice
		options2 := &PeerQueryOptions{
			PageSize:        3,
			PaginationToken: nextToken,
		}

		peersA, _, _, err := repo.GetPeersOptimized(ctx, options2)
		require.NoError(t, err)

		peersB, _, _, err := repo.GetPeersOptimized(ctx, options2)
		require.NoError(t, err)

		// Should get same results
		assert.Equal(t, len(peersA), len(peersB))
		for i := range peersA {
			assert.Equal(t, peersA[i].PeerID, peersB[i].PeerID)
		}
	})

	t.Run("Pagination preserves filter and sort across pages", func(t *testing.T) {
		// Complex query with filter and custom sort
		options := &PeerQueryOptions{
			PageSize: 2,
			Filter: &coretypes.PeerFilter{
				ClientNames: []string{"agent0", "agent1"}, // Will match peers 1,2,3,4,6,7,9,10
			},
			Sort:  "block_height",
			IsAsc: true, // Ascending
		}

		allPeers := []string{}
		token := ""
		pageCount := 0

		// Collect all pages
		for pageCount < 10 { // Safety limit
			opts := &PeerQueryOptions{
				PageSize:        options.PageSize,
				Filter:          options.Filter,
				Sort:            options.Sort,
				IsAsc:           options.IsAsc,
				PaginationToken: token,
			}

			peers, nextToken, _, err := repo.GetPeersOptimized(ctx, opts)
			require.NoError(t, err)

			for _, p := range peers {
				allPeers = append(allPeers, p.PeerID)
			}

			if nextToken == "" {
				break
			}
			token = nextToken
			pageCount++
		}

		// Verify we got the right peers in the right order
		// agent0: 3,6,9  agent1: 1,4,7,10
		// Sorted by block_height ascending: 1(1000), 3(3000), 4(4000), 6(6000), 7(7000), 9(9000), 10(10000)
		expected := []string{"QmTest01", "QmTest03", "QmTest04", "QmTest06", "QmTest07", "QmTest09", "QmTest10"}
		assert.Equal(t, expected, allPeers)
	})
}

func getPeerIDs(peers []*PeerInfo) []string {
	ids := make([]string, len(peers))
	for i, p := range peers {
		ids[i] = p.PeerID
	}
	return ids
}
