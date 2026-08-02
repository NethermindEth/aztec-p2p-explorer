package repo

import (
	"context"
	"database/sql"
	"log/slog"
)

type PeerRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewPeerRepository(db *sql.DB, logger *slog.Logger) *PeerRepository {
	return &PeerRepository{
		db:     db,
		logger: logger,
	}
}

// withTx runs the given function in a transaction and commits or rolls back the transaction based on the error
func (r *PeerRepository) withTx(ctx context.Context, fn func(*sql.Tx) error) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		var txErr error
		if p := recover(); p != nil {
			r.logger.Debug("Rolling back transaction due to panic", "panic", p)
			txErr = tx.Rollback()
			if txErr != nil {
				r.logger.Error("failed to rollback transaction", "error", err)
			}
			panic(p)
		} else if err != nil {
			r.logger.Debug("Rolling back transaction due to error", "error", err)
			txErr = tx.Rollback()
			if txErr != nil {
				r.logger.Error("failed to rollback transaction", "error", err)
			}
		} else {
			err = tx.Commit()
		}
	}()

	return fn(tx)
}
