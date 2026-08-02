package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"

	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func (r *PeerRepository) CreateCrawl(ctx context.Context, crawl *coretypes.Crawl, crawlStatus coretypes.CrawlStatus) (int, error) {
	var crawlID int
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		cr := &models.Crawl{
			State:           null.StringFrom(crawl.State),
			StartedAt:       crawl.StartedAt,
			FinishedAt:      crawl.FinishedAt,
			CrawledPeers:    null.IntFrom(crawl.CrawledPeers),
			DialablePeers:   null.IntFrom(crawl.DialablePeers),
			UndialablePeers: null.IntFrom(crawl.UndialablePeers),
			RemainingPeers:  null.IntFrom(crawl.RemainingPeers),
			Status:          string(crawlStatus),
			Version:         null.StringFrom(crawl.Version),
			ReferenceCCV:    null.NewString(crawl.ReferenceCCV, crawl.ReferenceCCV != ""),
		}

		err := cr.Insert(ctx, r.db, boil.Infer())
		if err != nil {
			return fmt.Errorf("insert crawl: %w", err)
		}

		crawlID = cr.ID
		return nil
	})

	return crawlID, err
}

func (r *PeerRepository) UpdateCrawlStatus(ctx context.Context, crawlID int, crawlStatus coretypes.CrawlStatus) error {
	cr := &models.Crawl{
		ID:     crawlID,
		Status: string(crawlStatus),
	}

	_, err := cr.Update(ctx, r.db, boil.Whitelist(models.CrawlColumns.Status))
	return err
}

func (r *PeerRepository) UpdateCrawlPeerCount(ctx context.Context, crawlID, peerCount int) error {
	cr := &models.Crawl{
		ID:                   crawlID,
		CCVFilteredPeerCount: null.IntFrom(peerCount),
	}

	rowsAff, err := cr.Update(ctx, r.db, boil.Whitelist(models.CrawlColumns.CCVFilteredPeerCount))
	if err != nil {
		return fmt.Errorf("failed to update crawl peer count: %w", err)
	}
	if rowsAff == 0 {
		return fmt.Errorf("crawl %d not found", crawlID)
	}
	return nil
}

func (r *PeerRepository) GetLastFinishedCrawlTimestamp(ctx context.Context) (time.Time, error) {
	crawl, err := models.Crawls(
		qm.Where("status = ?", "finished"),
		qm.OrderBy("id DESC"),
		qm.Limit(1),
	).One(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("get last finished crawl: %w", err)
	}
	return crawl.FinishedAt, nil
}
