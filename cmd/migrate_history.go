package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	"github.com/NethermindEth/aztec-p2p-explorer/database"
	"github.com/NethermindEth/aztec-p2p-explorer/repo"
)

func setupDestinationRepo(
	sourceDSN, destDSN string,
	sourceRepo *repo.PeerRepository,
	dryRun bool,
	logger *slog.Logger,
) (*repo.PeerRepository, *sql.DB, error) {
	if dryRun {
		return nil, nil, nil
	}

	if sourceDSN == destDSN {
		logger.Info("Using same database for source and destination")
		return sourceRepo, nil, nil
	}

	logger.Info("Connecting to destination database and running migrations")
	dbConfig := &database.Config{
		DatabaseSourceName: destDSN,
	}

	destDBWrapper, err := database.NewWithMigrate(dbConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to destination database: %w", err)
	}

	destRepo := repo.NewPeerRepository(destDBWrapper.DB, logger)
	return destRepo, destDBWrapper.DB, nil
}

func migrateSingleDay(
	ctx context.Context,
	destRepo *repo.PeerRepository,
	currentDate time.Time,
	count int,
	source string,
	dryRun bool,
	logger *slog.Logger,
) error {
	dateStr := currentDate.Format("2006-01-02")

	if dryRun {
		logger.Info("Would migrate data",
			"date", dateStr,
			"peer_count", count,
			"source", source,
		)
		return nil
	}

	if err := destRepo.UpsertDailyPeerHistory(ctx, currentDate, count); err != nil {
		return fmt.Errorf("failed to upsert daily peer history: %w", err)
	}

	logger.Info("Migrated data",
		"date", dateStr,
		"peer_count", count,
		"source", source,
	)
	return nil
}

//nolint:gocyclo
func migrateHistoryMain(cmd *cobra.Command, _ []string) {
	logger := configureLogging(cmd, nil)

	sourceDSN, _ := cmd.Flags().GetString(sourceDSNF)
	destDSN, _ := cmd.Flags().GetString(destDSNF)
	startDateStr, _ := cmd.Flags().GetString(startDateF)
	endDateStr, _ := cmd.Flags().GetString(endDateF)
	dryRun, _ := cmd.Flags().GetBool(dryRunF)

	if sourceDSN == "" {
		sourceDSN = destDSN
		logger.Info("Source DSN not provided, using destination DSN for both")
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		logger.Error("Invalid start date format. Use YYYY-MM-DD", "error", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		logger.Error("Invalid end date format. Use YYYY-MM-DD", "error", err)
		return
	}

	if startDate.After(endDate) {
		logger.Error("Start date must be before or equal to end date")
		return
	}

	logger.Info("Starting historical data migration",
		"start_date", startDate.Format("2006-01-02"),
		"end_date", endDate.Format("2006-01-02"),
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

	destRepo, destDB, err := setupDestinationRepo(sourceDSN, destDSN, sourceRepo, dryRun, logger)
	if err != nil {
		logger.Error("Failed to setup destination", "error", err)
		return
	}
	if destDB != nil {
		defer destDB.Close()
	}

	ctx := context.Background()

	totalDays := 0
	successDays := 0
	skippedDays := 0
	errorDays := 0

	peerIndexDays := 0
	crawlTableDays := 0

	currentDate := startDate
	for !currentDate.After(endDate) {
		totalDays++

		count, source, err := sourceRepo.GetDailyPeerCountForDate(ctx, currentDate)
		if err != nil {
			logger.Error("Failed to get peer count for date",
				"date", currentDate.Format("2006-01-02"),
				"error", err,
			)
			errorDays++
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		switch source {
		case "peer_visits_index":
			peerIndexDays++
		case "crawls.dialable_peers":
			crawlTableDays++
		}

		if count == 0 {
			logger.Debug("No data found for date, skipping",
				"date", currentDate.Format("2006-01-02"),
			)
			skippedDays++
			currentDate = currentDate.AddDate(0, 0, 1)
			continue
		}

		err = migrateSingleDay(ctx, destRepo, currentDate, count, source, dryRun, logger)
		if err != nil {
			logger.Error("Failed to migrate day",
				"date", currentDate.Format("2006-01-02"),
				"peer_count", count,
				"error", err,
			)
			errorDays++
		} else {
			successDays++
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	logger.Info("Migration completed",
		"total_days", totalDays,
		"success_days", successDays,
		"skipped_days", skippedDays,
		"error_days", errorDays,
		"peer_visits_index_days", peerIndexDays,
		"crawls_table_days", crawlTableDays,
		"dry_run", dryRun,
	)

	if dryRun {
		logger.Info("Dry run completed - no data was actually migrated")
	}

	if errorDays > 0 {
		logger.Warn(fmt.Sprintf("Migration completed with %d errors", errorDays))
	}
}
