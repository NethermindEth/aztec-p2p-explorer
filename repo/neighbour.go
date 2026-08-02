//nolint:misspell
package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/lib/pq"

	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func (r *PeerRepository) CreateNeighbors(ctx context.Context, neighbors *coretypes.Neighbors) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		return r.upsertNeighbor(ctx, tx, neighbors.PeerID, neighbors.Neighbors)
	})
}

// upsertNeighbor upserts a neighbor record
func (r *PeerRepository) upsertNeighbor(ctx context.Context, exec boil.ContextExecutor, peerID string, neighbors []string) error {
	pID, err := r.findOrUpsertPeer(ctx, exec, peerID, null.Int{}, null.Int{}, nil)
	if err != nil {
		return err
	}

	nIDs := types.Int64Array{}
	for _, n := range neighbors {
		nID, err := r.findOrUpsertPeer(ctx, exec, n, null.Int{}, null.Int{}, nil)
		if err != nil {
			return err
		}

		nIDs = append(nIDs, nID)
	}

	neighbor := &models.Neighbor{
		PeerID:      pID,
		NeighborIds: nIDs,
		UpdatedAt:   time.Now(),
	}

	err = neighbor.Upsert(ctx, exec, true,
		[]string{models.NeighborColumns.PeerID},
		boil.Whitelist(models.NeighborColumns.NeighborIds, models.NeighborColumns.UpdatedAt), boil.Infer())
	if err != nil {
		return fmt.Errorf("insert neighbor: %w", err)
	}

	return nil
}

type PeerNeighbors struct {
	PeerID    string         `json:"id" boil:"peer_id"`
	Neighbors pq.StringArray `json:"neighbors" boil:"neighbor_ids" swaggertype:"array,string"`
}

// GetPeersNeighbors retrieves all peers and their neighbors
func (r *PeerRepository) GetPeersNeighbors(ctx context.Context) ([]*PeerNeighbors, error) {
	var results []*PeerNeighbors
	err := models.NewQuery(
		qm.Select(`
			peers.multi_hash as peer_id,
			COALESCE(
				array_agg(neighbor_peers.multi_hash) 
				FILTER (WHERE neighbor_peers.multi_hash IS NOT NULL), 
				ARRAY[]::TEXT[]
			) AS neighbor_ids
		`),
		qm.From("peers"),
		qm.LeftOuterJoin("neighbors ON neighbors.peer_id = peers.id"),
		qm.LeftOuterJoin("unnest(neighbors.neighbor_ids) WITH ORDINALITY AS u(neighbor_id, ord) ON true"),
		qm.LeftOuterJoin("peers AS neighbor_peers ON neighbor_peers.id = neighbor_id::integer"),
		qm.GroupBy("peers.multi_hash"),
		qm.OrderBy("peers.multi_hash"),
	).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch neighbor peers: %w", err)
	}

	return results, nil
}

// GetPeerNeighbor retrieves the neighbors of a peer with the given peer ID
func (r *PeerRepository) GetPeerNeighbors(ctx context.Context, peerID string) (*PeerNeighbors, error) {
	peers, err := models.Peers(
		qm.Select("neighbor_peers.multi_hash as multi_hash"),
		qm.InnerJoin("neighbors ON neighbors.peer_id = peers.id"),
		qm.InnerJoin("peers AS neighbor_peers ON neighbor_peers.id = ANY(neighbors.neighbor_ids)"),
		qm.Where("peers.multi_hash = ?", peerID),
	).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch neighbor peers: %w", err)
	}

	neighborPeerIDs := make([]string, len(peers))
	for i, r := range peers {
		neighborPeerIDs[i] = r.MultiHash
	}

	return &PeerNeighbors{
		PeerID:    peerID,
		Neighbors: neighborPeerIDs,
	}, nil
}
