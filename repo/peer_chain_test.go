package repo

import (
	"context"
	"testing"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func TestCreatePeerChainInfos(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	// First, create a peer
	peer := &models.Peer{
		MultiHash: "TestPeer123",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err := peer.Insert(context.Background(), repo.db, boil.Infer())
	require.NoError(t, err)

	PeerChainInfo := &coretypes.PeerChainInfo{
		BlockHeight:         12345,
		BlockHeightObtained: true,
		SpecVersion:         "v1.2.3",
		SpecVersionObtained: true,
		SyncStatus:          true,
		SyncStatusObtained:  true,
	}

	err = repo.CreatePeerChainInfo(context.Background(), peer.MultiHash, PeerChainInfo)
	require.NoError(t, err)

	// Verify the peer state was inserted correctly
	dbPeer, err := models.Peers(models.PeerWhere.MultiHash.EQ("TestPeer123")).One(context.Background(), repo.db)
	require.NoError(t, err)
	dbPeerChainInfo, err := models.PeersStates(models.PeersStateWhere.PeerID.EQ(dbPeer.ID)).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, int64(PeerChainInfo.BlockHeight), dbPeerChainInfo.BlockHeight.Int64)
	assert.Equal(t, PeerChainInfo.SpecVersion, dbPeerChainInfo.SpecVersion.String)
	assert.True(t, dbPeerChainInfo.IsSynced.Bool)
}
