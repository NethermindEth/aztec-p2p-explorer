package repo

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"sort"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/types"

	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func (r *PeerRepository) CreateProtocols(ctx context.Context, protocolMap map[string]int) error {
	return r.withTx(ctx, func(tx *sql.Tx) error {
		return r.upsertProtocols(ctx, tx, protocolMap)
	})
}

// upsertProtocols upserts protocol records
func (r *PeerRepository) upsertProtocols(
	ctx context.Context,
	exec boil.ContextExecutor,
	protocolMap map[string]int,
) error {
	for pr, count := range protocolMap {
		_, err := r.upsertProtocol(ctx, exec, pr, count)
		if err != nil {
			return err
		}
	}

	return nil
}

// upsertProtocolsWithSet upserts protocol records and create a protocol set record
func (r *PeerRepository) upsertProtocolsWithSet(
	ctx context.Context,
	exec boil.ContextExecutor,
	protocolMap map[string]int,
) (int, error) {
	if len(protocolMap) == 0 {
		return 0, nil
	}

	protocolIDs := make([]int64, 0, len(protocolMap))
	for pr, count := range protocolMap {
		id, err := r.upsertProtocol(ctx, exec, pr, count)
		if err != nil {
			return 0, err
		}

		protocolIDs = append(protocolIDs, int64(id))
	}

	return r.upsertProtocolSet(ctx, exec, protocolIDs)
}

// upsertProtocolSet upserts a protocol set record and returns the protocol set ID
func (r *PeerRepository) upsertProtocolSet(ctx context.Context, exec boil.ContextExecutor, protocolIDs []int64) (int, error) {
	sort.Slice(protocolIDs, func(i, j int) bool { return protocolIDs[i] < protocolIDs[j] })

	ps := &models.ProtocolsSet{
		Hash:        protocolsSetHash(protocolIDs),
		ProtocolIds: types.Int64Array(protocolIDs),
	}

	err := ps.Upsert(ctx, exec, true, []string{models.ProtocolsSetColumns.Hash},
		boil.Whitelist(models.ProtocolsSetColumns.Hash), boil.Infer())
	if err != nil {
		return 0, fmt.Errorf("insert protocol set: %w", err)
	}

	return ps.ID, nil
}

// upsertProtocol upserts a protocol record and returns the protocol ID
func (r *PeerRepository) upsertProtocol(ctx context.Context, exec boil.ContextExecutor, protocol string, count int) (int, error) {
	pr := &models.Protocol{
		Protocol:  protocol,
		Count:     count,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	columns := []string{models.ProtocolColumns.Protocol}

	if count > 0 {
		columns = append(columns, models.ProtocolColumns.Count, models.ProtocolColumns.UpdatedAt)
	}

	err := pr.Upsert(ctx, exec, true,
		[]string{models.ProtocolColumns.Protocol},
		boil.Whitelist(columns...), boil.Infer())
	if err != nil {
		return 0, fmt.Errorf("upsert protocol: %w", err)
	}

	return pr.ID, nil
}

// protocolsSetHash returns a hash of the protocol IDs
func protocolsSetHash(protocolIDs []int64) []byte {
	h := sha256.New()
	var err error
	for _, id := range protocolIDs {
		err = binary.Write(h, binary.LittleEndian, id)
		if err != nil {
			panic(fmt.Sprintf("failed to write protocol ID to hash: %v", err))
		}
	}
	return h.Sum(nil)
}
