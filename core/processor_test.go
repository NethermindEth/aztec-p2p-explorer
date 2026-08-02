//nolint:misspell
package core

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/maxmind"
	"github.com/NethermindEth/aztec-p2p-explorer/repo"
	"github.com/NethermindEth/aztec-p2p-explorer/testutil"
)

func TestClassifyOutcome(t *testing.T) {
	cases := []struct {
		name     string
		synced   bool
		obtained bool
		want     string
	}{
		{"synced and obtained", true, true, "synced"},
		{"not synced but obtained", false, true, "not_synced"},
		{"unknown — obtained false, synced false", false, false, "unknown"},
		{"unknown — obtained false even if synced is true", true, false, "unknown"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := classifyOutcome(c.synced, c.obtained)
			assert.Equal(t, c.want, got)
		})
	}
}

func TestNebulaJSONProcessor(t *testing.T) {
	tempDir := t.TempDir()

	r, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)
	logger := slog.Default()
	processor := NewNebulaJSONProcessor(logger, r, maxmind.NewMockClient(), nil, 30, WithDataDir(tempDir))

	// Create test files
	timestamp := time.Now().Format("2006-01-02T15:04")
	createTestFile(t, tempDir, timestamp+"_"+crawlSuffix, &types.Crawl{
		State:           "completed",
		StartedAt:       time.Now().UTC(),
		FinishedAt:      time.Now().UTC().Add(time.Hour),
		CrawledPeers:    101,
		DialablePeers:   80,
		UndialablePeers: 20,
		RemainingPeers:  2,
		Version:         "v1.0.0",
	})
	createTestFile(t, tempDir, timestamp+"_"+crawlPropSuffix, &CrawlProperties{
		AgentVersion: map[string]int{"v1.0.0": 10},
		Protocol:     map[string]int{"p2p": 10},
	})
	createTestFile(t, tempDir, timestamp+"_"+visitSuffix, &RawVisit{
		PeerID:       "QmPeer123",
		MultiAddrs:   []string{"/ip4/10.20.30.40/tcp/443"},
		Protocols:    []string{"p2p"},
		AgentVersion: "v1.0.0",
		Visit: types.Visit{
			ConnectDuration: "1s",
			CrawlDuration:   "1s",
			VisitStartedAt:  time.Now().UTC(),
			VisitEndedAt:    time.Now().UTC().Add(time.Second),
			ConnectErrorStr: "",
			CrawlErrorStr:   "",
		},
	})
	createTestFile(t, tempDir, timestamp+"_"+neighborSuffix, &types.Neighbors{
		PeerID:    "QmPeer123",
		Neighbors: []string{"QmNeighbor123", "QmNeighbor456"},
	})

	// Run the processor
	crawl, err := processor.Process(context.Background())
	require.NoError(t, err)
	require.NotNil(t, crawl)

	// Check if files were deleted after processing
	files, err := filepath.Glob(filepath.Join(tempDir, "*"))
	require.NoError(t, err)
	require.Empty(t, files, "All files should have been deleted after processing")
}

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		// Plain semver
		{"4.1.3", true},
		{"4.1.2", true},
		{"0.87.2", true},

		// Pre-release suffixes (observed in testnet crawler logs)
		{"4.1.0-rc.2", true},
		{"4.1.0-rc.4", true},
		{"4.1.2-rc.1", true},
		{"4.2.0-nightly.20260408-1", true},
		{"4.2.0-nightly.20260407", true},
		{"4.2.0-aztecnr-rc.2", true},
		{"3.0.0-rc.6", true},
		{"2.1.9-rc.4-ag", true},
		{"2.1.11-rc.3", true},
		{"0.0.0-test", true},

		// Build metadata
		{"4.1.3+build.5", true},
		{"4.1.3-rc.1+build.5", true},

		// Invalid: raw git SHAs (no major.minor.patch prefix)
		{"b7b427f83a35ba8adfb525fd12de95815549a109", false},
		{"90928a47aa6dce7421caf759c453cc6e092f0081", false},

		// Invalid: missing components
		{"4.1", false},
		{"4", false},
		{"", false},

		// Invalid: leading "v" prefix
		{"v4.1.3", false},

		// Invalid: garbage
		{"not-a-version", false},
		{"4.1.3.4", false},
	}

	for _, tc := range tests {
		t.Run(tc.version, func(t *testing.T) {
			require.Equal(t, tc.valid, isValidVersion(tc.version))
		})
	}
}

func createTestFile(t *testing.T, dir, name string, data interface{}) {
	content, err := json.Marshal(data)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(dir, name), content, 0o600)
	require.NoError(t, err)
}

func setupTestRepo(t *testing.T) (*repo.PeerRepository, func()) {
	db, dbTearDown := testutil.PrepareTestDatabase(t)
	logger := slog.Default()
	r := repo.NewPeerRepository(db, logger)
	return r, dbTearDown
}
