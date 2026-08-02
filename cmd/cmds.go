package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Flag Names
const (
	bindAddrF             = "bind-addr"
	maxmindDirF           = "maxmind-dir"
	dbDSNF                = "db-dsn"
	nebulaProtocolsF      = "protocols"
	nebulaBootstrapPeersF = "bootstrap-peers"
	crawlIntervalF        = "interval"
	flushIntervalF        = "flush-interval"
	disableCrawlerF       = "disable-crawler"
	networkF              = "network"
	rpcURLF               = "rpc-url"
	nebulaWorkersF        = "workers"
	nebulaLimitF          = "limit"
	migrateDownF          = "migrate-down"
	maxBlockDiffF         = "max-block-diff"
	useIndexTableF        = "use-index-table"
	sourceDSNF            = "source-dsn"
	destDSNF              = "dest-dsn"
	startDateF            = "start-date"
	endDateF              = "end-date"
	dryRunF               = "dry-run"
	batchSizeF            = "batch-size"
)

// Shared Constants
const (
	NetworkSimulationInterval = 5 * time.Second
	GracefulShutdownTimeout   = 5 * time.Second
	DefaultCrawlInterval      = 30 * time.Minute
	DefaultNetwork            = "aztec-testnet"
	DefaultFlushInterval      = 5 * time.Minute
	DefaultWorkerCount        = 500
	defaultBatchSize          = 1000
	percentageMultiplier      = 100
)

// Default Values
const (
	//nolint:gosec // no secrets here
	defaultDBDSN = "postgres://explorer:explorer@localhost:15432/explorer?sslmode=disable"
)

var rootCmd = &cobra.Command{
	Use:   "aztec-p2p-explorer",
	Short: "Aztec P2P explorer",
}

func addSourceDSNFlag(cmd *cobra.Command, usage string) {
	cmd.Flags().String(sourceDSNF, "", usage)
}

func addDestDSNFlag(cmd *cobra.Command, usage string) {
	cmd.Flags().String(destDSNF, defaultDBDSN, usage)
}

func addDryRunFlag(cmd *cobra.Command, usage string) {
	cmd.Flags().Bool(dryRunF, false, usage)
}

func addBatchSizeFlag(cmd *cobra.Command, usage string) {
	cmd.Flags().Int(batchSizeF, defaultBatchSize, usage)
}

func addDateRangeFlags(cmd *cobra.Command, required bool) {
	cmd.Flags().String(
		startDateF,
		"",
		"Start date in YYYY-MM-DD format (required)",
	)
	cmd.Flags().String(
		endDateF,
		"",
		"End date in YYYY-MM-DD format (required)",
	)

	if required {
		_ = cmd.MarkFlagRequired(startDateF)
		_ = cmd.MarkFlagRequired(endDateF)
	}
}

func makeServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start API server",
		Run:   serverMain,
	}

	cmd.Flags().String(
		bindAddrF,
		":8080",
		"Address to bind to for the API server (e.g. :8080) or AZTEC_P2P_EXPLORER_BIND_ADDR in the environment",
	)
	cmd.Flags().String(
		maxmindDirF,
		"/var/lib/GeoIP",
		"Directory of the Maxmind database. It should contain both ASN and City databases",
	)
	cmd.Flags().String(
		dbDSNF,
		defaultDBDSN,
		"The database source name used when connecting to the database. It should be in the format of URL",
	)
	cmd.Flags().String(
		nebulaProtocolsF,
		"/aztec/kad/1.0.0",
		"Protocols to use for Nebula",
	)
	cmd.Flags().String(
		nebulaBootstrapPeersF,
		"",
		"Bootstrap peers for Nebula",
	)
	cmd.Flags().Bool(
		disableCrawlerF,
		false,
		"Disable the feature to crawl the network",
	)
	cmd.Flags().Duration(
		crawlIntervalF,
		DefaultCrawlInterval,
		"Interval to crawl the network",
	)
	cmd.Flags().Duration(
		flushIntervalF,
		DefaultFlushInterval,
		"Interval to flush the cache",
	)
	cmd.Flags().String(
		networkF,
		DefaultNetwork,
		"Network to crawl (aztec-testnet, aztec-mainnet)",
	)
	cmd.Flags().String(
		rpcURLF,
		"",
		"RPC URL for the Aztec node (overrides network default). Env: AZTEC_P2P_EXPLORER_RPC_URL",
	)
	cmd.Flags().Bool(
		migrateDownF,
		false,
		"Apply down migration",
	)
	cmd.Flags().Int(
		nebulaWorkersF,
		DefaultWorkerCount,
		"Number of workers to use for Nebula",
	)
	cmd.Flags().Int(
		nebulaLimitF,
		0,
		"Limit the number of peers to crawl (0 = unlimited)",
	)
	cmd.Flags().Uint64(
		maxBlockDiffF,
		10, //nolint:mnd
		"Maximum block difference allowed for a node to be considered synced (Aztec network only)",
	)
	cmd.Flags().Bool(
		useIndexTableF,
		true,
		"Use optimised index table for /peers and /peers/map endpoints (experimental)",
	)

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		log.Fatalf("bind flags: %v", err)
	}

	return cmd
}

func makeMigrateHistoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate-history",
		Short: "Migrate historical peer count data to daily_peer_history table",
		Long: `Migrate historical peer count data from peer_visits_index or crawls table to daily_peer_history table.
This command supports cross-database migration and automatically falls back to crawls.dialable_peers when
peer_visits_index data is not available.`,
		Run: migrateHistoryMain,
	}

	addSourceDSNFlag(cmd,
		"Source database DSN (if different from dest). If not provided, dest-dsn will be used for both source and dest")
	addDestDSNFlag(cmd,
		"Destination database DSN where daily_peer_history table will be populated")
	addDateRangeFlags(cmd, true)
	addDryRunFlag(cmd,
		"Dry run mode - show what would be migrated without inserting into database")

	return cmd
}

func makeMigrateFirstSeenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate-first-seen",
		Short: "Migrate and fix 'first seen' timestamps for peers",
		Long: `Migrate and fix the peers.created_at field (displayed as "First seen" in the frontend).
This command finds the earliest timestamp for each peer from historical data and updates the destination database.

The command processes all peers in the database and uses the following priority:
1. visits.visit_started_at (most reliable historical data)
2. peer_visits_index.created_at (fallback if visits not available)
3. peers.created_at (final fallback - current value)`,
		Run: migrateFirstSeenMain,
	}

	addSourceDSNFlag(cmd,
		"Source database DSN (if different from dest). If not provided, dest-dsn will be used for both source and dest")
	addDestDSNFlag(cmd,
		"Destination database DSN where peers.created_at will be updated")
	addDryRunFlag(cmd,
		"Dry run mode - show what would be migrated without updating database")
	addBatchSizeFlag(cmd,
		"Number of peers to process per batch")

	return cmd
}

func makeVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Returns the current version",
		Run:   versionMain,
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("debug", false, "Enables debug output.")
	rootCmd.PersistentFlags().Bool("json", false, "Enables structured logging in JSON format.")

	rootCmd.AddCommand(makeServerCommand())
	rootCmd.AddCommand(makeMigrateHistoryCommand())
	rootCmd.AddCommand(makeMigrateFirstSeenCommand())
	rootCmd.AddCommand(makeVersionCommand())
}
