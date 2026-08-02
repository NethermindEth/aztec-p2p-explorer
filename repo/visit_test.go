package repo

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
	"github.com/NethermindEth/aztec-p2p-explorer/testutil"
)

func setupTestRepo(t *testing.T) (*PeerRepository, func()) {
	db, dbTearDown := testutil.PrepareTestDatabase(t)
	logger := slog.Default()
	repo := NewPeerRepository(db, logger)
	return repo, dbTearDown
}

func TestCreatePeerVisit(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	crawl := &coretypes.Crawl{
		State:     "completed",
		StartedAt: time.Now().UTC(),
		Version:   "v1.0.0",
	}
	crawlID, err := repo.CreateCrawl(context.Background(), crawl, coretypes.CrawlStatusFinished)
	require.NoError(t, err)

	visit := &coretypes.Visit{
		VisitStartedAt:  time.Now().UTC(),
		VisitEndedAt:    time.Now().UTC().Add(time.Minute),
		ConnectErrorStr: "no error",
		CrawlErrorStr:   "no error",
		ConnectDuration: "4ms",
		CrawlDuration:   "59s",
	}

	peer := &coretypes.Peer{
		PeerID:       "TestPeer123",
		AgentVersion: "test-agent/1.0.0",
		Protocols:    []string{"testProto1"},
		MultiAddrs: []*coretypes.MultiAddr{
			{Address: "/ip4/127.0.0.1/tcp/1234"},
		},
	}

	err = repo.CreatePeerVisit(context.Background(), crawlID, visit, peer)
	require.NoError(t, err)

	// Verify the agent version was inserted correctly
	dbAgentVersion, err := models.AgentVersions(models.AgentVersionWhere.AgentVersion.EQ(peer.AgentVersion)).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, peer.AgentVersion, dbAgentVersion.AgentVersion)

	// Verify the protocols were inserted correctly
	dbProtocol, err := models.Protocols(models.ProtocolWhere.Protocol.EQ(peer.Protocols[0])).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, peer.Protocols[0], dbProtocol.Protocol)

	dbProtocolSet, err := models.ProtocolsSets(
		models.ProtocolsSetWhere.ProtocolIds.EQ(types.Int64Array{int64(dbProtocol.ID)})).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, int64(dbProtocol.ID), dbProtocolSet.ProtocolIds[0])

	// Verify the multi-address was inserted correctly
	dbMultiAddr, err := models.MultiAddresses(models.MultiAddressWhere.Maddr.EQ(peer.MultiAddrs[0].Address)).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, peer.MultiAddrs[0].Address, dbMultiAddr.Maddr)

	// Verify the peer was inserted correctly
	dbPeer, err := models.Peers(
		models.PeerWhere.MultiHash.EQ(peer.PeerID),
		qm.Load(models.PeerRels.MultiAddresses),
	).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, peer.PeerID, dbPeer.MultiHash)
	assert.Equal(t, dbAgentVersion.ID, dbPeer.AgentVersionID.Int)
	assert.Equal(t, dbProtocolSet.ID, dbPeer.ProtocolsSetID.Int)
	assert.Equal(t, dbMultiAddr.ID, dbPeer.R.MultiAddresses[0].ID)

	// Verify the peer visit was inserted correctly
	dbVisit, err := models.Visits(models.VisitWhere.PeerID.EQ(dbPeer.ID)).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, visit.VisitStartedAt.Unix(), dbVisit.VisitStartedAt.Unix())
	assert.Equal(t, visit.VisitEndedAt.Unix(), dbVisit.VisitEndedAt.Unix())
	assert.Equal(t, visit.ConnectErrorStr, dbVisit.ConnectError.String)
	assert.Equal(t, visit.CrawlErrorStr, dbVisit.CrawlError.String)
	assert.Equal(t, "00:00:00.004", dbVisit.ConnectDuration.String) // TODO: need a tool to parse duration strings in interval format
	assert.Equal(t, "00:00:59", dbVisit.CrawlDuration.String)
	assert.Equal(t, crawlID, dbVisit.CrawlID.Int)
	assert.Equal(t, int64(dbMultiAddr.ID), dbVisit.MultiAddressIds[0])
	assert.Equal(t, dbAgentVersion.ID, dbVisit.AgentVersionID.Int)
	assert.Equal(t, dbProtocolSet.ID, dbVisit.ProtocolsSetID.Int)
}

func TestCreateVisitWithDuplicateMultiAddrs(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	crawl := &coretypes.Crawl{
		Version: "v1.0.0",
	}
	crawlID, err := repo.CreateCrawl(context.Background(), crawl, coretypes.CrawlStatusFinished)
	require.NoError(t, err)

	visitStartedAt, err := time.Parse(time.RFC3339, "2024-08-09T17:51:07.802642+08:00")
	require.NoError(t, err)
	visitEndedAt, err := time.Parse(time.RFC3339, "2024-08-09T17:51:22.804044+08:00")
	require.NoError(t, err)

	visit := &coretypes.Visit{
		VisitStartedAt:  visitStartedAt,
		VisitEndedAt:    visitEndedAt,
		ConnectErrorStr: "io_timeout",
		CrawlErrorStr:   "",
		ConnectDuration: "15.001197958s",
		CrawlDuration:   "15.001416792s",
	}

	peer := &coretypes.Peer{
		PeerID:       "12D3KooWLvos6kiUH2kgxtyRqFsSKhg2yvfGwAbLpe48qit38u3N",
		AgentVersion: "",
		Protocols:    nil,
		MultiAddrs: []*coretypes.MultiAddr{
			{Address: "/ip4/134.255.180.42/udp/2121/quic-v1/webtransport/certhash/a"},
			{Address: "/ip4/134.255.180.42/udp/2121/quic-v1/webtransport/certhash/a"},
		},
	}

	err = repo.CreatePeerVisit(context.Background(), crawlID, visit, peer)
	require.Error(t, err)
}

func TestGetPeers(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	t.Run("Basic retrieval", func(t *testing.T) {
		t.Parallel()

		peers, nextToken, _, err := repo.GetPeers(ctx, &PeerQueryOptions{})
		require.NoError(t, err)
		assert.Len(t, peers, 4)
		assert.Empty(t, nextToken)
	})

	t.Run("Filtering by client name", func(t *testing.T) {
		t.Parallel()

		peers, _, _, err := repo.GetPeers(ctx, &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				ClientNames: []string{"alpha-node"},
			},
		})
		require.NoError(t, err)
		assert.Len(t, peers, 1)
		assert.Equal(t, "QmPeer1", peers[0].PeerID)
		assert.Equal(t, "alpha-node/v1.0.0", peers[0].AgentVersion.String)
		assert.Equal(t, int64(2000), peers[0].BlockHeight.Int64)
		assert.Equal(t, "v1.0.0", peers[0].SpecVersion.String)
		assert.Equal(t, true, peers[0].IsSynced.Bool)
	})

	t.Run("Filtering by sync status", func(t *testing.T) {
		t.Parallel()

		syncStatus := true
		peers, _, _, err := repo.GetPeers(ctx, &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				SyncStatus: &syncStatus,
			},
		})
		require.NoError(t, err)
		assert.Len(t, peers, 2)
	})

	//nolint:dupl
	t.Run("Sorting by created_at ascending", func(t *testing.T) {
		t.Parallel()

		peers, _, _, err := repo.GetPeers(ctx, &PeerQueryOptions{
			Sort:  "created_at",
			IsAsc: true,
		})
		require.NoError(t, err)
		assert.Len(t, peers, 4)
		assert.Equal(t, "QmPeer1", peers[0].PeerID)
		assert.Equal(t, "QmPeer2", peers[1].PeerID)
		assert.Equal(t, "QmPeer3", peers[2].PeerID)
		assert.Equal(t, "QmPeer4", peers[3].PeerID)
	})

	//nolint:dupl
	t.Run("Sorting by created_at descending", func(t *testing.T) {
		t.Parallel()

		peers, _, _, err := repo.GetPeers(ctx, &PeerQueryOptions{
			Sort:  "created_at",
			IsAsc: false,
		})
		require.NoError(t, err)
		assert.Len(t, peers, 4)
		assert.Equal(t, "QmPeer4", peers[0].PeerID)
		assert.Equal(t, "QmPeer3", peers[1].PeerID)
		assert.Equal(t, "QmPeer2", peers[2].PeerID)
		assert.Equal(t, "QmPeer1", peers[3].PeerID)
	})

	t.Run("Pagination, no filter, sorting by created_at ascending", func(t *testing.T) {
		t.Parallel()

		// First page
		peers, nextToken, _, err := repo.GetPeers(ctx, &PeerQueryOptions{
			Sort:     "created_at",
			PageSize: 2,
		})
		require.NoError(t, err)
		assert.Len(t, peers, 2)
		assert.Equal(t, "QmPeer4", peers[0].PeerID)
		assert.Equal(t, "QmPeer3", peers[1].PeerID)
		assert.NotEmpty(t, nextToken)

		// Second page
		peers, nextToken, _, err = repo.GetPeers(ctx, &PeerQueryOptions{
			Sort:            "created_at",
			PageSize:        2,
			PaginationToken: nextToken,
		})
		require.NoError(t, err)
		assert.Len(t, peers, 2)
		assert.Equal(t, "QmPeer2", peers[0].PeerID)
		assert.Equal(t, "QmPeer1", peers[1].PeerID)
		assert.Empty(t, nextToken)
	})

	t.Run("Filtering by continent names", func(t *testing.T) {
		t.Parallel()

		peers, _, _, err := repo.GetPeers(ctx, &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				Continents: []string{"North America"},
			},
		})
		require.NoError(t, err)
		assert.Len(t, peers, 2)
		assert.Equal(t, "QmPeer4", peers[0].PeerID)
		assert.Equal(t, "QmPeer1", peers[1].PeerID)
	})

	t.Run("Filtering by continent codes", func(t *testing.T) {
		t.Parallel()

		peers, _, _, err := repo.GetPeers(ctx, &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				ContinentsCodes: []string{"NA"},
			},
		})
		require.NoError(t, err)
		assert.Len(t, peers, 2)
		assert.Equal(t, "QmPeer4", peers[0].PeerID)
		assert.Equal(t, "QmPeer1", peers[1].PeerID)
	})

	t.Run("Filtering by country names", func(t *testing.T) {
		t.Parallel()

		peers, _, _, err := repo.GetPeers(ctx, &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				Countries: []string{"United Kingdom"},
			},
		})
		require.NoError(t, err)
		assert.Len(t, peers, 1)
		assert.Equal(t, "QmPeer2", peers[0].PeerID)
	})

	t.Run("Filtering by country codes", func(t *testing.T) {
		t.Parallel()

		peers, _, _, err := repo.GetPeers(ctx, &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				CountriesISOCodes: []string{"GB"},
			},
		})
		require.NoError(t, err)
		assert.Len(t, peers, 1)
		assert.Equal(t, "QmPeer2", peers[0].PeerID)
	})

	t.Run("Filtering by AS names", func(t *testing.T) {
		t.Parallel()

		peers, _, _, err := repo.GetPeers(ctx, &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				ASOrganizations: []string{"Google LLC"},
			},
		})
		require.NoError(t, err)
		assert.Len(t, peers, 1)
		assert.Equal(t, "QmPeer3", peers[0].PeerID)
	})

	t.Run("Filtering by AS numbers", func(t *testing.T) {
		t.Parallel()

		peers, _, _, err := repo.GetPeers(ctx, &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				ASNumbers: []string{"15169"},
			},
		})
		require.NoError(t, err)
		assert.Len(t, peers, 1)
		assert.Equal(t, "QmPeer3", peers[0].PeerID)
	})

	t.Run("Combined filters", func(t *testing.T) {
		t.Parallel()

		syncStatus := true

		peers, _, _, err := repo.GetPeers(ctx, &PeerQueryOptions{
			Filter: &coretypes.PeerFilter{
				ClientNames: []string{"alpha-node", "beta-node"},
				SyncStatus:  &syncStatus,
				Continents:  []string{"North America"},
			},
		})
		require.NoError(t, err)
		assert.Len(t, peers, 1)
		assert.Equal(t, "QmPeer1", peers[0].PeerID)
	})

	t.Run("Invalid pagination token", func(t *testing.T) {
		t.Parallel()

		_, _, _, err := repo.GetPeers(ctx, &PeerQueryOptions{
			PaginationToken: "invalid_token",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid pagination token")
	})

	t.Run("Pagination token mismatch", func(t *testing.T) {
		t.Parallel()

		// Create a valid token but with mismatched parameters
		token := PaginationToken{
			LastPeerID:    "QmPeer2",
			LastSortValue: time.Now(),
			SortColumn:    "last_seen",
			IsAsc:         true,
			Filter:        &coretypes.PeerFilter{},
		}

		encodedToken, err := token.EncodeToString()
		require.NoError(t, err)
		_, _, _, err = repo.GetPeers(ctx, &PeerQueryOptions{
			PaginationToken: encodedToken,
		})
		assert.Error(t, err)
	})

	t.Run("Get latest peers only", func(t *testing.T) {
		t.Parallel()

		peers, _, _, err := repo.GetPeers(ctx, &PeerQueryOptions{
			Latest: true,
		})
		require.NoError(t, err)
		assert.Len(t, peers, 4)
	})
}

func TestGetPeerByID(t *testing.T) {
	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	ctx := context.Background()

	t.Run("Successful retrieval", func(t *testing.T) {
		peerID := "QmPeer1"
		peer, err := repo.GetPeerByID(ctx, peerID)
		require.NoError(t, err)
		require.NotNil(t, peer)

		assert.Equal(t, peerID, peer.PeerID)
		assert.Equal(t, "alpha-node/v1.0.0", peer.AgentVersion.String)
		assert.Equal(t, int64(2000), peer.BlockHeight.Int64)
		assert.Equal(t, "v1.0.0", peer.SpecVersion.String)
		assert.True(t, peer.IsSynced.Bool)

		// Check MultiAddresses
		require.Len(t, peer.MultiAddresses, 1)
		maddr := peer.MultiAddresses[0]
		assert.Equal(t, "/ip4/192.0.2.1/tcp/8080", maddr.Address)
		require.Len(t, maddr.IPList, 1)
		ipInfo := maddr.IPList[0]
		assert.Equal(t, "192.0.2.1", ipInfo.IPAddress)
		assert.Equal(t, 8080, ipInfo.Port)
		assert.Equal(t, "Cloudflare, Inc.", ipInfo.ASOrganization)
		assert.Equal(t, uint(13335), ipInfo.ASNumber)
		assert.Equal(t, "North America", ipInfo.Continent)
		assert.Equal(t, "NA", ipInfo.ContinentCode)
		assert.Equal(t, "United States", ipInfo.Country)
		assert.Equal(t, "US", ipInfo.CountryISO)
		assert.InDelta(t, 40.7128, ipInfo.Latitude, 0.0001)
		assert.InDelta(t, -74.0060, ipInfo.Longitude, 0.0001)

		// Check Protocols
		assert.Equal(t, "/kad/1.0.0", peer.Protocols[0])
	})

	t.Run("Non-existent peer", func(t *testing.T) {
		peerID := "QmNonExistentPeer"
		peer, err := repo.GetPeerByID(ctx, peerID)
		assert.NoError(t, err)
		assert.Nil(t, peer)
	})

	t.Run("Peer with multiple MultiAddresses", func(t *testing.T) {
		peerID := "QmPeer3"
		peer, err := repo.GetPeerByID(ctx, peerID)
		require.NoError(t, err)
		require.NotNil(t, peer)

		assert.Equal(t, peerID, peer.PeerID)
		assert.Equal(t, "delta-node/v2.1.0", peer.AgentVersion.String)

		// Check MultiAddresses
		require.Len(t, peer.MultiAddresses, 1)
		maddr := peer.MultiAddresses[0]
		assert.Equal(t, "/ip4/203.0.113.2/tcp/7070/ip6/2001:db8::1/udp/1234", maddr.Address)
		require.Len(t, maddr.IPList, 2)

		// Check first IP
		ipInfo1 := maddr.IPList[0]
		assert.Equal(t, "203.0.113.1", ipInfo1.IPAddress)
		assert.Equal(t, 7070, ipInfo1.Port)

		// Check second IP
		ipInfo2 := maddr.IPList[1]
		assert.Equal(t, "2001:db8::1", ipInfo2.IPAddress)
		assert.Equal(t, 1234, ipInfo2.Port)
	})

	t.Run("Peer with null values", func(t *testing.T) {
		peerID := "QmPeer2"
		peer, err := repo.GetPeerByID(ctx, peerID)
		require.NoError(t, err)
		require.NotNil(t, peer)

		assert.Equal(t, peerID, peer.PeerID)
		assert.Equal(t, "beta-node/v0.9.0", peer.AgentVersion.String)
		assert.Equal(t, int64(2000), peer.BlockHeight.Int64)
		assert.Equal(t, "v1.1.0", peer.SpecVersion.String)
		assert.False(t, peer.IsSynced.Bool)
	})

	t.Run("Case-insensitive search - lowercase", func(t *testing.T) {
		peerID := "qmpeer1" // lowercase version
		peer, err := repo.GetPeerByID(ctx, peerID)
		require.NoError(t, err)
		require.NotNil(t, peer)

		assert.Equal(t, "QmPeer1", peer.PeerID) // Should match the original case
		assert.Equal(t, "alpha-node/v1.0.0", peer.AgentVersion.String)
	})

	t.Run("Case-insensitive search - uppercase", func(t *testing.T) {
		peerID := "QMPEER2" // uppercase version
		peer, err := repo.GetPeerByID(ctx, peerID)
		require.NoError(t, err)
		require.NotNil(t, peer)

		assert.Equal(t, "QmPeer2", peer.PeerID) // Should match the original case
		assert.Equal(t, "beta-node/v0.9.0", peer.AgentVersion.String)
	})

	t.Run("Case-insensitive search - mixed case", func(t *testing.T) {
		peerID := "qMpEeR3" // mixed case version
		peer, err := repo.GetPeerByID(ctx, peerID)
		require.NoError(t, err)
		require.NotNil(t, peer)

		assert.Equal(t, "QmPeer3", peer.PeerID) // Should match the original case
		assert.Equal(t, "delta-node/v2.1.0", peer.AgentVersion.String)
	})

	t.Run("Case-insensitive search - non-existent peer", func(t *testing.T) {
		peerID := "qMnOnExIsTeNtPeEr" // non-existent in any case
		peer, err := repo.GetPeerByID(ctx, peerID)
		assert.NoError(t, err)
		assert.Nil(t, peer)
	})
}
