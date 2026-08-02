package cmd

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/NethermindEth/aztec-p2p-explorer/core"
	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database"
	"github.com/NethermindEth/aztec-p2p-explorer/maxmind"
	"github.com/NethermindEth/aztec-p2p-explorer/repo"
	"github.com/NethermindEth/aztec-p2p-explorer/server"
)

func runServer(
	bindAddr string,
	logger *slog.Logger,
	r *repo.PeerRepository,
	wg *sync.WaitGroup,
	stop chan struct{},
	onCrawlProcessedChan chan *coretypes.Crawl,
	useIndexTable bool,
	aztecFeeder *core.AztecFeederClient,
) {
	defer wg.Done()

	s := server.New(bindAddr, logger, r, onCrawlProcessedChan, useIndexTable, aztecFeeder)

	go func() {
		err := s.Start()
		cobra.CheckErr(err)
	}()

	slog.Info("Aztec P2P explorer API ready",
		"port", bindAddr)

	<-stop

	slog.Info("Shutting down Aztec P2P explorer API...")

	// give ourself 5 seconds to shutdown gracefully
	ctx, cancel := context.WithTimeout(context.Background(), GracefulShutdownTimeout)
	defer cancel()

	err := s.Stop(ctx)
	if err != nil {
		slog.Error("Server stopped with errors.", "error", err)
	}

	slog.Info("Server stopped")
}

func serverMain(cmd *cobra.Command, args []string) {
	printBanner()
	logger := configureLogging(cmd, args)

	bindAddr := viper.GetString(bindAddrF)
	maxmindPath := viper.GetString(maxmindDirF)
	maxMindClient, err := maxmind.NewClient(maxmindPath)
	cobra.CheckErr(err)

	dbConfig := &database.Config{
		DatabaseSourceName: viper.GetString(dbDSNF),
		ApplyDownMigration: viper.GetBool(migrateDownF),
	}

	db, err := database.NewWithMigrate(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialise database: %v", err)
	}

	r := repo.NewPeerRepository(db.DB, logger)

	stop := make(chan struct{})
	onCrawlProcessedChan := make(chan *coretypes.Crawl)
	waitGroup := &sync.WaitGroup{}

	ctx := context.Background()
	rpcURL := viper.GetString(rpcURLF)
	if rpcURL == "" {
		log.Fatal("RPC URL is required: set --rpc-url flag or AZTEC_P2P_EXPLORER_RPC_URL environment variable")
	}

	maxBlockDiff := viper.GetUint64(maxBlockDiffF)
	if maxBlockDiff == 0 {
		maxBlockDiff = 30
	}
	const historyBuffer = 10
	historySize := int(maxBlockDiff*2 + historyBuffer) //nolint:gosec // Safe conversion for small values

	aztecFeeder := core.NewAztecFeederClient(logger, rpcURL,
		core.WithHistorySize(historySize))
	if err := aztecFeeder.Start(ctx); err != nil {
		log.Fatalf("Failed to start Aztec feeder client: %v", err)
	}

	waitGroup.Add(1)
	go runServer(bindAddr, logger, r, waitGroup, stop, onCrawlProcessedChan, viper.GetBool(useIndexTableF), aztecFeeder)

	if !viper.GetBool(disableCrawlerF) {
		// create executor
		executorOpts := []func(*core.NebulaExecutor){
			core.WithBootstrapPeers(viper.GetString(nebulaBootstrapPeersF)),
			core.WithWorkers(viper.GetInt(nebulaWorkersF)),
			core.WithNetwork(viper.GetString(networkF)),
		}
		if limit := viper.GetInt(nebulaLimitF); limit > 0 {
			executorOpts = append(executorOpts, core.WithLimit(limit))
		}
		executor := core.NewNebulaExecutor(logger, executorOpts...)

		// start feeder client

		// create repository
		peerRepo := repo.NewPeerRepository(db.DB, logger)

		// create processor

		processor := core.NewNebulaJSONProcessor(
			logger,
			peerRepo,
			maxMindClient,
			aztecFeeder,
			maxBlockDiff,
		)

		// create engine
		engine := core.NewEngine(executor, processor, logger, onCrawlProcessedChan, core.WithInterval(viper.GetDuration(crawlIntervalF)))

		// start engine
		go engine.Start(ctx)
	} else {
		go func() {
			ticker := time.NewTicker(viper.GetDuration(flushIntervalF))
			defer ticker.Stop()

			for range ticker.C {
				onCrawlProcessedChan <- &coretypes.Crawl{FinishedAt: time.Now()}
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	sig := <-quit
	slog.Info("Received signal", "signal", sig)

	// signal the API server to stop
	close(stop)
	waitGroup.Wait()
}
