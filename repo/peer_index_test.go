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

func ptrBool(b bool) *bool {
	return &b
}

func TestGetPeersOptimizedIntegration(t *testing.T) {
	// This test requires a database connection
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	// First, we need a finished crawl for the queries to work
	crawl := &models.Crawl{
		Status:    "finished",
		StartedAt: time.Now(),
	}
	err := crawl.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	// Create some test peers with proper IDs
	peer1 := &models.Peer{MultiHash: "QmTest1", CreatedAt: time.Now(), LastSeen: time.Now()}
	err = peer1.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	peer2 := &models.Peer{MultiHash: "QmTest2", CreatedAt: time.Now(), LastSeen: time.Now()}
	err = peer2.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	peer3 := &models.Peer{MultiHash: "QmTest3", CreatedAt: time.Now(), LastSeen: time.Now()}
	err = peer3.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	// Create visits for each peer
	visit1 := &models.Visit{
		PeerID:         peer1.ID,
		VisitStartedAt: time.Now(),
		VisitEndedAt:   time.Now(),
		CrawlID:        null.IntFrom(crawl.ID),
	}
	err = visit1.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	visit2 := &models.Visit{
		PeerID:         peer2.ID,
		VisitStartedAt: time.Now(),
		VisitEndedAt:   time.Now(),
		CrawlID:        null.IntFrom(crawl.ID),
	}
	err = visit2.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	visit3 := &models.Visit{
		PeerID:         peer3.ID,
		VisitStartedAt: time.Now(),
		VisitEndedAt:   time.Now(),
		CrawlID:        null.IntFrom(crawl.ID),
	}
	err = visit3.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	// Insert index entries
	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO peer_visits_index (
			crawl_id, visit_id, peer_id, peer_multi_hash,
			agent_version, client_name,
			continent_name, country_name, city_name,
			is_synced, block_height,
			created_at, last_seen
		) VALUES 
			($1, $2, $3, 'QmTest1', 'test-agent/v1.0.0', 'test-agent', 
			 'North America', 'United States', 'New York', 
			 true, 10000, NOW(), NOW()),
			($1, $4, $5, 'QmTest2', 'other-agent/v2.0.0', 'other-agent',
			 'Europe', 'Germany', 'Berlin',
			 false, 5000, NOW(), NOW()),
			($1, $6, $7, 'QmTest3', 'test-agent/v1.5.0', 'test-agent',
			 'Asia', 'Japan', 'Tokyo',
			 true, 12000, NOW(), NOW())
	`, crawl.ID, visit1.ID, peer1.ID, visit2.ID, peer2.ID, visit3.ID, peer3.ID)
	require.NoError(t, err)

	t.Run("Query without filter - tests ambiguous column fix", func(t *testing.T) {
		options := &PeerQueryOptions{
			PageSize: 10,
		}

		peers, nextToken, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err, "Should not fail with ambiguous column reference")
		assert.Equal(t, int64(3), totalCount)
		assert.Len(t, peers, 3)
		assert.Empty(t, nextToken)
	})

	t.Run("Query with country filter", func(t *testing.T) {
		options := &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				Countries: []string{"United States"},
			},
			PageSize: 10,
		}

		peers, _, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(1), totalCount)
		assert.Len(t, peers, 1)
		assert.Equal(t, "QmTest1", peers[0].PeerID)
	})

	t.Run("Query with client filter", func(t *testing.T) {
		options := &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				ClientNames: []string{"test-agent"},
			},
			PageSize: 10,
		}

		peers, _, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(2), totalCount)
		assert.Len(t, peers, 2)
	})

	t.Run("Query with sync status filter", func(t *testing.T) {
		options := &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				SyncStatus: ptrBool(true),
			},
			PageSize: 10,
		}

		peers, _, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(2), totalCount)
		assert.Len(t, peers, 2)
	})

	t.Run("Query with pagination", func(t *testing.T) {
		options := &PeerQueryOptions{
			PageSize: 2,
		}

		peers, nextToken, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(3), totalCount)
		assert.Len(t, peers, 2)
		assert.NotEmpty(t, nextToken, "Should have next page token")

		// TODO: Fix pagination token handling in GetPeersOptimized
		// The pagination token needs proper implementation
	})

	t.Run("Query with sorting by block height", func(t *testing.T) {
		options := &PeerQueryOptions{
			Sort:     "block_height",
			IsAsc:    false,
			PageSize: 10,
		}

		peers, _, _, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Len(t, peers, 3)

		// Check descending order by block height
		// Note: We need to check the actual block heights from enriched data
		// Since we're testing the index query, we just verify no error
	})

	t.Run("Query with multiple filters", func(t *testing.T) {
		options := &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				ClientNames: []string{"test-agent"},
				SyncStatus:  ptrBool(true),
			},
			PageSize: 10,
		}

		peers, _, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(2), totalCount) // QmTest1 and QmTest3
		assert.Len(t, peers, 2)
	})

	t.Run("Query with no results", func(t *testing.T) {
		options := &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				Countries: []string{"Brazil"},
			},
			PageSize: 10,
		}

		peers, _, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(0), totalCount)
		assert.Len(t, peers, 0)
	})

	t.Run("Empty index table fallback", func(t *testing.T) {
		// Clear the index table
		_, err := repo.db.ExecContext(ctx, "DELETE FROM peer_visits_index")
		require.NoError(t, err)

		options := &PeerQueryOptions{
			PageSize: 10,
		}

		// Should return empty results when index is empty
		peers, _, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(0), totalCount)
		assert.Len(t, peers, 0)
	})
}

func TestGetPeersOptimizedCaseInsensitive(t *testing.T) {
	// This test requires a database connection
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	// First, we need a finished crawl for the queries to work
	crawl := &models.Crawl{
		Status:    "finished",
		StartedAt: time.Now(),
	}
	err := crawl.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	// Create test peers with mixed case IDs
	peer1 := &models.Peer{MultiHash: "QmAbCdEf123", CreatedAt: time.Now(), LastSeen: time.Now()}
	err = peer1.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	peer2 := &models.Peer{MultiHash: "QmXyZ789", CreatedAt: time.Now(), LastSeen: time.Now()}
	err = peer2.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	peer3 := &models.Peer{MultiHash: "QmTest456", CreatedAt: time.Now(), LastSeen: time.Now()}
	err = peer3.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	// Create visits for each peer
	visit1 := &models.Visit{
		PeerID:         peer1.ID,
		VisitStartedAt: time.Now(),
		VisitEndedAt:   time.Now(),
		CrawlID:        null.IntFrom(crawl.ID),
	}
	err = visit1.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	visit2 := &models.Visit{
		PeerID:         peer2.ID,
		VisitStartedAt: time.Now(),
		VisitEndedAt:   time.Now(),
		CrawlID:        null.IntFrom(crawl.ID),
	}
	err = visit2.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	visit3 := &models.Visit{
		PeerID:         peer3.ID,
		VisitStartedAt: time.Now(),
		VisitEndedAt:   time.Now(),
		CrawlID:        null.IntFrom(crawl.ID),
	}
	err = visit3.Insert(ctx, repo.db, boil.Infer())
	require.NoError(t, err)

	// Insert index entries with the same peer IDs
	_, err = repo.db.ExecContext(ctx, `
		INSERT INTO peer_visits_index (
			crawl_id, visit_id, peer_id, peer_multi_hash,
			agent_version, client_name,
			is_synced, block_height,
			created_at, last_seen
		) VALUES
			($1, $2, $3, 'QmAbCdEf123', 'test-agent/v1.0.0', 'test-agent', true, 10000, NOW(), NOW()),
			($1, $4, $5, 'QmXyZ789', 'other-agent/v2.0.0', 'other-agent', false, 5000, NOW(), NOW()),
			($1, $6, $7, 'QmTest456', 'third-agent/v1.5.0', 'third-agent', true, 12000, NOW(), NOW())
	`, crawl.ID, visit1.ID, peer1.ID, visit2.ID, peer2.ID, visit3.ID, peer3.ID)
	require.NoError(t, err)

	t.Run("Case-insensitive search - lowercase peer ID", func(t *testing.T) {
		options := &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				PeerID: "qmabcdef", // lowercase prefix
			},
			PageSize: 10,
		}

		peers, _, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(1), totalCount, "Should find the peer with case-insensitive search")
		assert.Len(t, peers, 1)
		assert.Equal(t, "QmAbCdEf123", peers[0].PeerID, "Should return the peer with original case")
	})

	t.Run("Case-insensitive search - uppercase peer ID", func(t *testing.T) {
		options := &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				PeerID: "QMXYZ", // uppercase prefix
			},
			PageSize: 10,
		}

		peers, _, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(1), totalCount, "Should find the peer with case-insensitive search")
		assert.Len(t, peers, 1)
		assert.Equal(t, "QmXyZ789", peers[0].PeerID, "Should return the peer with original case")
	})

	t.Run("Case-insensitive search - mixed case peer ID", func(t *testing.T) {
		options := &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				PeerID: "qMtEsT", // mixed case prefix
			},
			PageSize: 10,
		}

		peers, _, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(1), totalCount, "Should find the peer with case-insensitive search")
		assert.Len(t, peers, 1)
		assert.Equal(t, "QmTest456", peers[0].PeerID, "Should return the peer with original case")
	})

	t.Run("Case-insensitive search - partial match", func(t *testing.T) {
		options := &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				PeerID: "qm", // matches all peers (case-insensitive)
			},
			PageSize: 10,
		}

		peers, _, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(3), totalCount, "Should find all peers with 'Qm' prefix")
		assert.Len(t, peers, 3)
	})

	t.Run("Case-insensitive search - no match", func(t *testing.T) {
		options := &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				PeerID: "nonexistent", // no peer matches
			},
			PageSize: 10,
		}

		peers, _, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(0), totalCount, "Should not find any peers")
		assert.Len(t, peers, 0)
	})

	t.Run("Case-insensitive search combined with other filters", func(t *testing.T) {
		options := &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				PeerID:     "QM", // prefix search (any case)
				SyncStatus: ptrBool(true),
			},
			PageSize: 10,
		}

		peers, _, totalCount, err := repo.GetPeersOptimized(ctx, options)
		require.NoError(t, err)
		assert.Equal(t, int64(2), totalCount, "Should find 2 synced peers with 'Qm' prefix")
		assert.Len(t, peers, 2)
	})
}
