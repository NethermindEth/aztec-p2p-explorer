package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/NethermindEth/aztec-p2p-explorer/repo/queries"
)

type PeerTimestamp struct {
	MultiHash string
	FirstSeen sql.NullTime
	Source    string
}

func (r *PeerRepository) GetEarliestPeerTimestamps(ctx context.Context, batchSize, offset int) ([]*PeerTimestamp, error) {
	rows, err := r.db.QueryContext(ctx, queries.GetEarliestPeerTimestamps, batchSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query earliest peer timestamps: %w", err)
	}
	defer rows.Close()

	var results []*PeerTimestamp
	for rows.Next() {
		var pt PeerTimestamp
		err := rows.Scan(&pt.MultiHash, &pt.FirstSeen, &pt.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to scan peer timestamp: %w", err)
		}
		results = append(results, &pt)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating peer timestamps: %w", err)
	}

	return results, nil
}

func (r *PeerRepository) UpdatePeerCreatedAt(ctx context.Context, multiHash string, createdAt sql.NullTime) error {
	if !createdAt.Valid {
		return fmt.Errorf("created_at timestamp is null for peer %s", multiHash)
	}

	query := `
		UPDATE peers
		SET created_at = $1
		WHERE multi_hash = $2
	`

	result, err := r.db.ExecContext(ctx, query, createdAt.Time, multiHash)
	if err != nil {
		return fmt.Errorf("failed to update peer created_at: %w", err)
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAff == 0 {
		return fmt.Errorf("peer %s not found", multiHash)
	}

	return nil
}

func (r *PeerRepository) CountPeers(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM peers").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count peers: %w", err)
	}
	return count, nil
}
