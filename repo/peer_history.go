package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aarondl/sqlboiler/v4/queries/qm"

	"github.com/NethermindEth/aztec-p2p-explorer/database/models"
)

func (r *PeerRepository) UpsertDailyPeerHistory(ctx context.Context, date time.Time, peerCount int) error {
	query := `
		INSERT INTO daily_peer_history (date, peer_count, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (date) DO UPDATE
		SET peer_count = GREATEST(daily_peer_history.peer_count, EXCLUDED.peer_count),
		    updated_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query, date, peerCount)
	if err != nil {
		return fmt.Errorf("failed to upsert daily peer history: %w", err)
	}

	return nil
}

func (r *PeerRepository) GetDailyPeerHistory(ctx context.Context, start, end time.Time) ([]*PeerHistoryPoint, error) {
	if end.IsZero() {
		end = time.Now()
	}

	if start.After(end) {
		return nil, fmt.Errorf("start date must be before end date")
	}

	var results []*PeerHistoryPoint
	err := models.NewQuery(
		qm.Select("date, peer_count"),
		qm.From("daily_peer_history"),
		qm.Where("date >= ?", start),
		qm.Where("date < ?", end),
		qm.OrderBy("date ASC"),
	).Bind(ctx, r.db, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily peer history: %w", err)
	}

	return results, nil
}

func (r *PeerRepository) CalculateAndStorePeerHistory(ctx context.Context, crawlID int) error {
	var peerCount int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(DISTINCT peer_multi_hash) FROM peer_visits_index WHERE crawl_id = $1",
		crawlID,
	).Scan(&peerCount)
	if err != nil {
		return fmt.Errorf("failed to count peers for crawl %d: %w", crawlID, err)
	}

	err = r.UpdateCrawlPeerCount(ctx, crawlID, peerCount)
	if err != nil {
		return fmt.Errorf("failed to update crawl peer count: %w", err)
	}

	crawl, err := models.FindCrawl(ctx, r.db, crawlID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("crawl %d not found", crawlID)
		}
		return fmt.Errorf("failed to get crawl date: %w", err)
	}

	crawlDate := time.Date(
		crawl.FinishedAt.Year(),
		crawl.FinishedAt.Month(),
		crawl.FinishedAt.Day(),
		0, 0, 0, 0,
		crawl.FinishedAt.Location(),
	)

	err = r.UpsertDailyPeerHistory(ctx, crawlDate, peerCount)
	if err != nil {
		return fmt.Errorf("failed to upsert daily peer history: %w", err)
	}

	return nil
}

func (r *PeerRepository) GetDailyPeerCountForDate(ctx context.Context, date time.Time) (count int, source string, err error) {
	normalizedDate := time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		0, 0, 0, 0,
		time.UTC,
	)

	query := `
		SELECT COALESCE(MAX(peer_count), 0) as max_count
		FROM (
			SELECT crawl_id, COUNT(DISTINCT peer_multi_hash) as peer_count
			FROM peer_visits_index pvi
			INNER JOIN crawls c ON pvi.crawl_id = c.id
			WHERE c.finished_at::date = $1::date
			GROUP BY crawl_id
		) counts
	`

	var peerCount int
	err = r.db.QueryRowContext(ctx, query, normalizedDate).Scan(&peerCount)
	if err != nil {
		return 0, "", fmt.Errorf("failed to query peer_visits_index: %w", err)
	}

	if peerCount > 0 {
		return peerCount, "peer_visits_index", nil
	}

	fallbackQuery := `
		SELECT COALESCE(MAX(dialable_peers), 0)
		FROM crawls
		WHERE finished_at::date = $1::date
		  AND dialable_peers IS NOT NULL
	`

	var dialablePeers int
	err = r.db.QueryRowContext(ctx, fallbackQuery, normalizedDate).Scan(&dialablePeers)
	if err != nil {
		return 0, "", fmt.Errorf("failed to query crawls.dialable_peers: %w", err)
	}

	return dialablePeers, "crawls.dialable_peers", nil
}
