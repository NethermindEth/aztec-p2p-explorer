package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"

	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func (r *PeerRepository) CreateAgentVersions(ctx context.Context, agentMap map[string]int) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		return r.upsertAgentVersions(ctx, tx, agentMap)
	})
}

// upsertAgentVersions upserts agent version records
func (r *PeerRepository) upsertAgentVersions(ctx context.Context, exec boil.ContextExecutor, agentMap map[string]int) error {
	for agentVersion, count := range agentMap {
		_, err := r.upsertAgentVersion(ctx, exec, agentVersion, count)
		if err != nil {
			return err
		}
	}
	return nil
}

// upsertAgentVersion upserts an agent version record and returns the agent version ID
func (r *PeerRepository) upsertAgentVersion(ctx context.Context, exec boil.ContextExecutor, agentVersion string, count int) (int, error) {
	if agentVersion == "" {
		return 0, nil
	}

	av := &models.AgentVersion{
		AgentVersion: agentVersion,
		Count:        count,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	columns := []string{models.AgentVersionColumns.AgentVersion}

	if count > 0 {
		columns = append(columns, models.AgentVersionColumns.Count, models.AgentVersionColumns.UpdatedAt)
	}

	err := av.Upsert(ctx, exec, true,
		[]string{models.AgentVersionColumns.AgentVersion},
		boil.Whitelist(columns...), boil.Infer())
	if err != nil {
		return 0, fmt.Errorf("upsert agent version: %w", err)
	}

	return av.ID, nil
}
