package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NethermindEth/aztec-p2p-explorer/core/types"
)

func TestGetPeersForMap(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	t.Run("Get all peers with geographic data", func(t *testing.T) {
		t.Parallel()
		options := &PeerQueryOptions{
			PageSize: 100,
		}

		entries, err := repo.GetPeersForMap(ctx, options)
		require.NoError(t, err)
		require.NotEmpty(t, entries)

		// Verify all entries have geographic data
		for _, entry := range entries {
			assert.NotEmpty(t, entry.PeerMultiHash)
			assert.True(t, entry.CityName.Valid)
			assert.True(t, entry.CountryName.Valid)
			assert.True(t, entry.ContinentName.Valid)
			assert.True(t, entry.Latitude.Valid)
			assert.True(t, entry.Longitude.Valid)

			// Check for reasonable coordinate values
			assert.True(t, entry.Latitude.Float64 >= -90 && entry.Latitude.Float64 <= 90)
			assert.True(t, entry.Longitude.Float64 >= -180 && entry.Longitude.Float64 <= 180)
		}
	})

	t.Run("Filter by country", func(t *testing.T) {
		t.Parallel()
		filter := &types.PeerFilter{
			Countries: []string{"United States"},
		}
		options := &PeerQueryOptions{
			Filter:   filter,
			PageSize: 100,
		}

		entries, err := repo.GetPeersForMap(ctx, options)
		require.NoError(t, err)

		// Should have US peers (New York and San Francisco from fixtures)
		assert.Len(t, entries, 2)
		for _, entry := range entries {
			assert.Equal(t, "United States", entry.CountryName.String)
		}
	})

	t.Run("Filter by client", func(t *testing.T) {
		t.Parallel()
		filter := &types.PeerFilter{
			ClientNames: []string{"alpha-node"},
		}
		options := &PeerQueryOptions{
			Filter:   filter,
			PageSize: 100,
		}

		entries, err := repo.GetPeersForMap(ctx, options)
		require.NoError(t, err)

		// Should have alpha-node client peers (original QmPeer1 + new QmPeer5)
		assert.Len(t, entries, 2)
		// Check that we have both alpha-node peers
		peerHashes := make([]string, len(entries))
		for i, entry := range entries {
			peerHashes[i] = entry.PeerMultiHash
		}
		assert.Contains(t, peerHashes, "QmPeer1")
		assert.Contains(t, peerHashes, "QmPeer5")
	})

	t.Run("DISTINCT ON peer_id ensures unique peers", func(t *testing.T) {
		t.Parallel()
		options := &PeerQueryOptions{
			PageSize: 100,
		}

		entries, err := repo.GetPeersForMap(ctx, options)
		require.NoError(t, err)

		// Check that peer IDs are unique
		peerIDs := make(map[int64]bool)
		for _, entry := range entries {
			assert.False(t, peerIDs[entry.PeerID], "Duplicate peer ID found: %d", entry.PeerID)
			peerIDs[entry.PeerID] = true
		}
	})
}
