package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	"github.com/NethermindEth/aztec-p2p-explorer/repo"
)

func migrateSinglePeer(
	ctx context.Context,
	destRepo *repo.PeerRepository,
	multiHash string,
	firstSeen sql.NullTime,
	source string,
	dryRun bool,
	logger *slog.Logger,
) error {
	if !firstSeen.Valid {
		logger.Warn("Peer has null timestamp, skipping",
			"multi_hash", multiHash,
		)
		return fmt.Errorf("null timestamp")
	}

	if dryRun {
		logger.Debug("Would update peer",
			"multi_hash", multiHash,
			"first_seen", firstSeen.Time.Format(time.RFC3339),
			"source", source,
		)
		return nil
	}

	if err := destRepo.UpdatePeerCreatedAt(ctx, multiHash, firstSeen); err != nil {
		return fmt.Errorf("failed to update peer: %w", err)
	}

	logger.Debug("Updated peer",
		"multi_hash", multiHash,
		"first_seen", firstSeen.Time.Format(time.RFC3339),
		"source", source,
	)
	return nil
}

//nolint:gocyclo
func migrateFirstSeenMain(cmd *cobra.Command, _ []string) {
	logger := configureLogging(cmd, nil)

	sourceDSN, _ := cmd.Flags().GetString(sourceDSNF)
	destDSN, _ := cmd.Flags().GetString(destDSNF)
	dryRun, _ := cmd.Flags().GetBool(dryRunF)
	batchSize, _ := cmd.Flags().GetInt(batchSizeF)

	if sourceDSN == "" {
		sourceDSN = destDSN
		logger.Info("Source DSN not provided, using destination DSN for both")
	}

	logger.Info("Starting 'first seen' timestamp migration",
		"batch_size", batchSize,
		"dry_run", dryRun,
	)

	logger.Info("Connecting to source database")
	sourceDB, err := sql.Open("postgres", sourceDSN)
	if err != nil {
		logger.Error("Failed to connect to source database", "error", err)
		return
	}
	defer sourceDB.Close()

	sourceRepo := repo.NewPeerRepository(sourceDB, logger)

	totalPeers, err := sourceRepo.CountPeers(context.Background())
	if err != nil {
		logger.Error("Failed to count peers", "error", err)
		return
	}

	logger.Info("Total peers to process", "count", totalPeers)

	destRepo, destDB, err := setupDestinationRepo(sourceDSN, destDSN, sourceRepo, dryRun, logger)
	if err != nil {
		logger.Error("Failed to setup destination", "error", err)
		return
	}
	if destDB != nil {
		defer destDB.Close()
	}

	ctx := context.Background()

	offset := 0
	totalUpdated := 0
	totalSkipped := 0
	totalErrors := 0

	visitsSources := 0
	indexSources := 0
	peersSources := 0

	startTime := time.Now()

	for offset < totalPeers {
		logger.Info("Processing batch",
			"offset", offset,
			"batch_size", batchSize,
			"progress", fmt.Sprintf("%d/%d (%.1f%%)", offset, totalPeers, float64(offset)/float64(totalPeers)*percentageMultiplier),
		)

		timestamps, err := sourceRepo.GetEarliestPeerTimestamps(ctx, batchSize, offset)
		if err != nil {
			logger.Error("Failed to get peer timestamps", "error", err, "offset", offset)
			totalErrors++
			break
		}

		if len(timestamps) == 0 {
			logger.Info("No more peers to process")
			break
		}

		for _, pt := range timestamps {
			switch pt.Source {
			case "visits":
				visitsSources++
			case "peer_visits_index":
				indexSources++
			case "peers":
				peersSources++
			}

			err = migrateSinglePeer(ctx, destRepo, pt.MultiHash, pt.FirstSeen, pt.Source, dryRun, logger)
			if err != nil {
				if err.Error() == "null timestamp" {
					totalSkipped++
				} else {
					logger.Error("Failed to migrate peer",
						"multi_hash", pt.MultiHash,
						"error", err,
					)
					totalErrors++
				}
			} else {
				totalUpdated++
			}
		}

		if offset += len(timestamps); (offset/batchSize)%10 == 0 {
			elapsed := time.Since(startTime)
			peersPerSecond := float64(offset) / elapsed.Seconds()
			remainingPeers := totalPeers - offset
			estimatedTimeRemaining := time.Duration(float64(remainingPeers)/peersPerSecond) * time.Second

			logger.Info("Progress update",
				"processed", offset,
				"total", totalPeers,
				"updated", totalUpdated,
				"skipped", totalSkipped,
				"errors", totalErrors,
				"elapsed", elapsed.Round(time.Second),
				"peers_per_second", fmt.Sprintf("%.1f", peersPerSecond),
				"estimated_time_remaining", estimatedTimeRemaining.Round(time.Second),
			)
		}
	}

	elapsed := time.Since(startTime)
	logger.Info("Migration completed",
		"total_peers", totalPeers,
		"updated", totalUpdated,
		"skipped", totalSkipped,
		"errors", totalErrors,
		"visits_source", visitsSources,
		"peer_visits_index_source", indexSources,
		"peers_source", peersSources,
		"elapsed_time", elapsed.Round(time.Second),
		"dry_run", dryRun,
	)

	if dryRun {
		logger.Info("Dry run completed - no data was actually migrated")
	}

	if totalErrors > 0 {
		logger.Warn(fmt.Sprintf("Migration completed with %d errors", totalErrors))
	}
}
