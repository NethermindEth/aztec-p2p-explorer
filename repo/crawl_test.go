package repo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func TestCreateCrawl(t *testing.T) {
	t.Parallel()

	repo, tearDown := setupTestRepo(t)
	t.Cleanup(tearDown)

	crawl := &coretypes.Crawl{
		State:           "completed",
		StartedAt:       time.Now().UTC(),
		FinishedAt:      time.Now().UTC().Add(time.Hour),
		CrawledPeers:    101,
		DialablePeers:   80,
		UndialablePeers: 20,
		RemainingPeers:  2,
		Version:         "v1.0.0",
	}

	crawlID, err := repo.CreateCrawl(context.Background(), crawl, coretypes.CrawlStatusRunning)
	require.NoError(t, err)

	// Verify the crawl was inserted correctly
	dbCrawl, err := models.Crawls(models.CrawlWhere.ID.EQ(crawlID)).One(context.Background(), repo.db)
	require.NoError(t, err)
	assert.Equal(t, crawlID, dbCrawl.ID)
	assert.Equal(t, crawl.State, dbCrawl.State.String)
	assert.Equal(t, crawl.StartedAt.Unix(), dbCrawl.StartedAt.Unix())
	assert.Equal(t, crawl.FinishedAt.Unix(), dbCrawl.FinishedAt.Unix())
	assert.Equal(t, crawl.CrawledPeers, dbCrawl.CrawledPeers.Int)
	assert.Equal(t, crawl.DialablePeers, dbCrawl.DialablePeers.Int)
	assert.Equal(t, crawl.UndialablePeers, dbCrawl.UndialablePeers.Int)
	assert.Equal(t, crawl.RemainingPeers, dbCrawl.RemainingPeers.Int)
	assert.Equal(t, crawl.Version, dbCrawl.Version.String)
}
