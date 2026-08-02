//nolint:dupl
package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPeerCountByCountryOptimized(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	t.Run("Without CCV filter", func(t *testing.T) {
		results, err := repo.GetPeerCountByCountryOptimized(ctx, "")
		require.NoError(t, err)
		assert.Len(t, results, 5) // US, Germany, Japan, France, Singapore (based on peer_visits_index fixture)

		// Check expected countries from fixture data
		found := make(map[string]bool)
		for _, result := range results {
			found[result.CountryName] = true
			assert.Greater(t, result.Count, 0)
		}
		assert.True(t, found["United States"])
		assert.True(t, found["Germany"])
		assert.True(t, found["Japan"])
		assert.True(t, found["France"])
		assert.True(t, found["Singapore"])
	})

	t.Run("With CCV filter for 0.13.0", func(t *testing.T) {
		results, err := repo.GetPeerCountByCountryOptimized(ctx, "0.13.0")
		require.NoError(t, err)
		assert.Len(t, results, 3) // Only 4 peers have spec_version 0.13.0: US(2), Germany(1), Japan(1)

		// Check country names and counts for 0.13.0 peers
		found := make(map[string]bool)
		for _, result := range results {
			found[result.CountryName] = true
			assert.Greater(t, result.Count, 0)
		}
		assert.True(t, found["United States"])
		assert.True(t, found["Germany"])
		assert.True(t, found["Japan"])
		// Should NOT have France or Singapore (they have 0.12.0)
		assert.False(t, found["France"])
		assert.False(t, found["Singapore"])
	})

	t.Run("With CCV filter for 0.12.0", func(t *testing.T) {
		results, err := repo.GetPeerCountByCountryOptimized(ctx, "0.12.0")
		require.NoError(t, err)
		assert.Len(t, results, 2) // Only 2 peers have spec_version 0.12.0: France(1), Singapore(1)

		// Check country names and counts for 0.12.0 peers
		found := make(map[string]bool)
		for _, result := range results {
			found[result.CountryName] = true
			assert.Greater(t, result.Count, 0)
		}
		assert.True(t, found["France"])
		assert.True(t, found["Singapore"])
		// Should NOT have US, Germany, Japan (they have 0.13.0)
		assert.False(t, found["United States"])
		assert.False(t, found["Germany"])
		assert.False(t, found["Japan"])
	})

	t.Run("With CCV filter matching no peers", func(t *testing.T) {
		results, err := repo.GetPeerCountByCountryOptimized(ctx, "0.99.0")
		require.NoError(t, err)
		assert.Len(t, results, 0) // No peers with this spec_version
	})
}

func TestGetPeerCountByAgentOptimized(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	t.Run("Without CCV filter", func(t *testing.T) {
		results, err := repo.GetPeerCountByAgentOptimized(ctx, "")
		require.NoError(t, err)
		assert.Len(t, results, 6) // All 6 agent versions from fixture data

		// Check expected agents from fixture data
		found := make(map[string]bool)
		for _, result := range results {
			found[result.Client] = true
			assert.Greater(t, result.Count, 0)
		}
		assert.True(t, found["alpha-node/v1.0.0"])
		assert.True(t, found["beta-node/v0.10.0"])
		assert.True(t, found["gamma-node/v0.5.0"])
		assert.True(t, found["delta-node/v0.3.0"])
		assert.True(t, found["alpha-node/v1.1.0"])
		assert.True(t, found["beta-node/v0.11.0"])
	})

	t.Run("With CCV filter for 0.13.0", func(t *testing.T) {
		results, err := repo.GetPeerCountByAgentOptimized(ctx, "0.13.0")
		require.NoError(t, err)
		assert.Len(t, results, 4) // Only 4 agents have spec_version 0.13.0

		// Check client names for 0.13.0 peers
		found := make(map[string]bool)
		for _, result := range results {
			found[result.Client] = true
			assert.Greater(t, result.Count, 0)
		}
		assert.True(t, found["alpha-node/v1.0.0"])
		assert.True(t, found["beta-node/v0.10.0"])
		assert.True(t, found["gamma-node/v0.5.0"])
		assert.True(t, found["delta-node/v0.3.0"])
		// Should NOT have the 0.12.0 versions
		assert.False(t, found["alpha-node/v1.1.0"])
		assert.False(t, found["beta-node/v0.11.0"])
	})

	t.Run("With CCV filter for 0.12.0", func(t *testing.T) {
		results, err := repo.GetPeerCountByAgentOptimized(ctx, "0.12.0")
		require.NoError(t, err)
		assert.Len(t, results, 2) // Only 2 agents have spec_version 0.12.0

		// Check client names for 0.12.0 peers
		found := make(map[string]bool)
		for _, result := range results {
			found[result.Client] = true
			assert.Greater(t, result.Count, 0)
		}
		assert.True(t, found["alpha-node/v1.1.0"])
		assert.True(t, found["beta-node/v0.11.0"])
		// Should NOT have the 0.13.0 versions
		assert.False(t, found["alpha-node/v1.0.0"])
		assert.False(t, found["beta-node/v0.10.0"])
		assert.False(t, found["gamma-node/v0.5.0"])
		assert.False(t, found["delta-node/v0.3.0"])
	})

	t.Run("With CCV filter matching no peers", func(t *testing.T) {
		results, err := repo.GetPeerCountByAgentOptimized(ctx, "0.99.0")
		require.NoError(t, err)
		assert.Len(t, results, 0) // No peers with this spec_version
	})
}

func TestGetSyncStatusCountOptimized(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	t.Run("Without CCV filter", func(t *testing.T) {
		results, err := repo.GetSyncStatusCountOptimized(ctx, "")
		require.NoError(t, err)
		assert.Len(t, results, 2) // synced and not_synced

		// Check sync status counts - based on fixture data:
		// 0.13.0 peers: Peer1(synced), Peer2(not_synced), Peer3(synced), Peer4(not_synced) = 2 synced, 2 not_synced
		// 0.12.0 peers: Peer5(synced), Peer6(not_synced) = 1 synced, 1 not_synced
		// Total: 3 synced, 3 not_synced
		assert.Equal(t, 3, results["synced"])
		assert.Equal(t, 3, results["not_synced"])
	})

	t.Run("With CCV filter for 0.13.0", func(t *testing.T) {
		results, err := repo.GetSyncStatusCountOptimized(ctx, "0.13.0")
		require.NoError(t, err)
		assert.Len(t, results, 2) // synced and not_synced

		// Check sync status counts for 0.13.0 peers only:
		// Peer1: synced=true, Peer2: synced=false, Peer3: synced=true, Peer4: synced=false
		assert.Equal(t, 2, results["synced"])
		assert.Equal(t, 2, results["not_synced"])
	})

	t.Run("With CCV filter for 0.12.0", func(t *testing.T) {
		results, err := repo.GetSyncStatusCountOptimized(ctx, "0.12.0")
		require.NoError(t, err)
		assert.Len(t, results, 2) // synced and not_synced

		// Check sync status counts for 0.12.0 peers only:
		// Peer5: synced=true, Peer6: synced=false
		assert.Equal(t, 1, results["synced"])
		assert.Equal(t, 1, results["not_synced"])
	})

	t.Run("With CCV filter matching no peers", func(t *testing.T) {
		results, err := repo.GetSyncStatusCountOptimized(ctx, "0.99.0")
		require.NoError(t, err)
		assert.Len(t, results, 0) // No peers with this spec_version
	})
}

func TestGetLatestCrawlTotalCountOptimized(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	t.Run("Without CCV filter", func(t *testing.T) {
		count, err := repo.GetLatestCrawlTotalCountOptimized(ctx, "")
		require.NoError(t, err)
		assert.Equal(t, int64(6), count) // 6 peers in peer_visits_index fixture data
	})

	t.Run("With CCV filter for 0.13.0", func(t *testing.T) {
		count, err := repo.GetLatestCrawlTotalCountOptimized(ctx, "0.13.0")
		require.NoError(t, err)
		assert.Equal(t, int64(4), count) // 4 peers have spec_version 0.13.0
	})

	t.Run("With CCV filter for 0.12.0", func(t *testing.T) {
		count, err := repo.GetLatestCrawlTotalCountOptimized(ctx, "0.12.0")
		require.NoError(t, err)
		assert.Equal(t, int64(2), count) // 2 peers have spec_version 0.12.0
	})

	t.Run("With CCV filter matching no peers", func(t *testing.T) {
		count, err := repo.GetLatestCrawlTotalCountOptimized(ctx, "0.99.0")
		require.NoError(t, err)
		assert.Equal(t, int64(0), count) // No peers with this spec_version
	})
}

// TestOptimizedMethodsConsistency tests that optimised methods work correctly
// These tests focus on the CCV filtering functionality rather than consistency
// with regular methods since they use different data sources (peer_visits_index vs joins)
func TestOptimizedMethodsConsistency(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	t.Run("CCV filtering produces subset of results", func(t *testing.T) {
		// Without filter should return all results
		allResults, err := repo.GetPeerCountByCountryOptimized(ctx, "")
		require.NoError(t, err)
		assert.Len(t, allResults, 5) // 5 countries total

		// With filter for 0.13.0 should return subset
		filteredResults, err := repo.GetPeerCountByCountryOptimized(ctx, "0.13.0")
		require.NoError(t, err)
		assert.Len(t, filteredResults, 3) // Only 3 countries with 0.13.0 peers
		assert.Less(t, len(filteredResults), len(allResults))

		// With filter for non-existing spec_version should return empty
		emptyResults, err := repo.GetPeerCountByCountryOptimized(ctx, "0.99.0")
		require.NoError(t, err)
		assert.Len(t, emptyResults, 0)
	})

	t.Run("All optimised methods handle empty CCV filter correctly", func(t *testing.T) {
		// All methods should work with empty CCV filter
		countries, err := repo.GetPeerCountByCountryOptimized(ctx, "")
		require.NoError(t, err)
		assert.Greater(t, len(countries), 0)

		agents, err := repo.GetPeerCountByAgentOptimized(ctx, "")
		require.NoError(t, err)
		assert.Greater(t, len(agents), 0)

		syncStatus, err := repo.GetSyncStatusCountOptimized(ctx, "")
		require.NoError(t, err)
		assert.Greater(t, len(syncStatus), 0)

		totalCount, err := repo.GetLatestCrawlTotalCountOptimized(ctx, "")
		require.NoError(t, err)
		assert.Greater(t, totalCount, int64(0))
	})
}
