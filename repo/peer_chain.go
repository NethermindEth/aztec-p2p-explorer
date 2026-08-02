package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"

	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func (r *PeerRepository) CreatePeerChainInfo(ctx context.Context, peerID string, state *coretypes.PeerChainInfo) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		pID, err := r.findOrUpsertPeer(ctx, tx, peerID, null.Int{}, null.Int{}, nil)
		if err != nil {
			return err
		}

		ps := &models.PeersState{
			PeerID:    pID,
			CreatedAt: time.Now(),
		}

		if state.BlockHeightObtained {
			ps.BlockHeight = null.Int64From(int64(state.BlockHeight)) //nolint:gosec // ignore integer overflow
		}

		if state.SpecVersionObtained {
			ps.SpecVersion = null.StringFrom(state.SpecVersion)
		}

		if state.SyncStatusObtained {
			ps.IsSynced = null.BoolFrom(state.SyncStatus)
		}

		err = ps.Insert(ctx, tx, boil.Infer())
		if err != nil {
			return fmt.Errorf("insert peer state: %w", err)
		}

		return nil
	})
}
