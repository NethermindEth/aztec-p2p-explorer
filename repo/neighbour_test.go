//nolint:misspell
package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func TestCreateNeighbors(t *testing.T) {
	t.Parallel()

	r, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	neighbor := &coretypes.Neighbors{
		PeerID:    "TestPeer123",
		Neighbors: []string{"TestNeighbor1", "TestNeighbor2"},
	}

	err := r.CreateNeighbors(context.Background(), neighbor)
	require.NoError(t, err)

	// Verify the neighbor was inserted correctly
	dbPeer, err := models.Peers(models.PeerWhere.MultiHash.EQ("TestPeer123")).One(context.Background(), r.db)
	require.NoError(t, err)
	dbNeighbor, err := models.Neighbors(models.NeighborWhere.PeerID.EQ(dbPeer.ID)).One(context.Background(), r.db)
	require.NoError(t, err)
	assert.Len(t, dbNeighbor.NeighborIds, 2)
}

func TestGetPeersNeighbors(t *testing.T) {
	t.Parallel()

	r, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	peers, err := r.GetPeersNeighbors(context.Background())
	require.NoError(t, err)
	assert.Len(t, peers, 6)

	expectedPeersNeighbors := map[string][]string{
		"QmPeer1": {"QmPeer2", "QmPeer3"},
		"QmPeer2": {"QmPeer1", "QmPeer3"},
		"QmPeer3": {"QmPeer1", "QmPeer2"},
		"QmPeer4": {},
	}

	for _, peer := range peers {
		assert.ElementsMatch(t, expectedPeersNeighbors[peer.PeerID], peer.Neighbors)
	}
}

func TestGetPeerNeighbors(t *testing.T) {
	t.Parallel()

	r, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	testCases := []struct {
		peerID    string
		neighbors []string
	}{
		{
			peerID:    "QmPeer1",
			neighbors: []string{"QmPeer2", "QmPeer3"},
		},
		{
			peerID:    "QmPeer2",
			neighbors: []string{"QmPeer1", "QmPeer3"},
		},
		{
			peerID:    "QmPeer3",
			neighbors: []string{"QmPeer1", "QmPeer2"},
		},
	}

	for _, tc := range testCases {
		neighbors, err := r.GetPeerNeighbors(context.Background(), tc.peerID)
		require.NoError(t, err)
		assert.ElementsMatch(t, tc.neighbors, neighbors.Neighbors)
	}
}
