package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/NethermindEth/aztec-p2p-explorer/core/types"
)

const (
	defaultLogLevel  = "6"
	defaultAddrDial  = "any"
	defaultAddrTrack = "any"
	dataDirPerm      = 0o755
)

type Executor interface {
	Execute(ctx context.Context) error
}

type NebulaExecutor struct {
	logger         *slog.Logger
	nebulaLogger   *nebulaLogger
	protocols      string
	bootstrapPeers string
	workers        int
	limit          int
	network        string
}

func NewNebulaExecutor(logger *slog.Logger, options ...func(*NebulaExecutor)) *NebulaExecutor {
	executor := &NebulaExecutor{
		logger:       logger,
		nebulaLogger: &nebulaLogger{logger: logger},
	}

	for _, option := range options {
		option(executor)
	}

	return executor
}

func WithProtocols(protocols string) func(*NebulaExecutor) {
	return func(n *NebulaExecutor) {
		n.protocols = protocols
	}
}

func WithBootstrapPeers(bootstrapPeers string) func(*NebulaExecutor) {
	return func(n *NebulaExecutor) {
		n.bootstrapPeers = bootstrapPeers
	}
}

func WithWorkers(workers int) func(*NebulaExecutor) {
	return func(n *NebulaExecutor) {
		n.workers = workers
	}
}

func WithLimit(limit int) func(*NebulaExecutor) {
	return func(n *NebulaExecutor) {
		n.limit = limit
	}
}

func WithNetwork(network string) func(*NebulaExecutor) {
	return func(n *NebulaExecutor) {
		n.network = network
	}
}

//nolint:misspell
func (n *NebulaExecutor) Execute(ctx context.Context) error {
	safeBootstrapPeers := sanitizeInput(n.bootstrapPeers)

	// Ensure the Nebula data directory exists
	if err := os.MkdirAll(types.NebulaDataDirPath, dataDirPerm); err != nil {
		return fmt.Errorf("create nebula data dir: %w", err)
	}

	// Map network names to nebula network IDs
	networkID := n.network
	switch n.network {
	case "aztec-testnet", "aztec":
		networkID = "AZTEC_TESTNET"
	case "aztec-mainnet":
		networkID = "AZTEC_MAINNET"
	}

	// Build nebula command arguments
	args := []string{
		"--json-out", types.NebulaDataDirPath,
		"--log-level", defaultLogLevel,
		"crawl",
		"--keep-enr",
		"--addr-dial-type", defaultAddrDial,
		"--addr-track-type", defaultAddrTrack,
		"--network", networkID,
		"--workers", fmt.Sprintf("%d", n.workers),
		"--bootstrap-peers", safeBootstrapPeers,
	}

	// Add limit if specified
	if n.limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", n.limit))
	}

	// Add neighbors flag
	args = append(args, "--neighbors")

	cmd := exec.CommandContext(ctx, "/usr/local/bin/nebula", args...)

	cmd.Stdout = n.nebulaLogger
	cmd.Stderr = n.nebulaLogger

	n.logger.Info("Running nebula...", slog.String("args", strings.Join(cmd.Args, " ")))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("run nebula: %w", err)
	}

	return nil
}

func sanitizeInput(input string) string {
	return strings.TrimSpace(input)
}

// isValidInput checks if the input string contains only allowed characters
// func isValidInput(input string) bool {
// 	// Define a regex pattern for allowed characters (alphanumeric and some special characters)
// 	pattern := `^[a-zA-Z0-9,.\-_/]+$`
// 	matched, _ := regexp.MatchString(pattern, input)
// 	return matched
// }

type nebulaLogger struct {
	logger *slog.Logger
}

func (w *nebulaLogger) Write(p []byte) (n int, err error) {
	w.logger.Info(string(p))
	return len(p), nil
}
