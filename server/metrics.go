package server

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/NethermindEth/aztec-p2p-explorer/core"
	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	"github.com/NethermindEth/aztec-p2p-explorer/repo"
)

const (
	metricsPrefix     = "aztec_p2p_explorer_"
	ageUpdateInterval = 10 * time.Second
)

type BusinessMetrics struct {
	ActiveNodeCount   prometheus.Gauge
	SyncStatusTotal   *prometheus.GaugeVec
	LastCrawlAge      prometheus.Gauge
	CrawlSuccessTotal prometheus.Counter
	logger            *slog.Logger
	repo              *repo.PeerRepository

	mu            sync.RWMutex
	lastCrawlTime time.Time
}

func NewBusinessMetrics(logger *slog.Logger, r *repo.PeerRepository, feeder *core.AztecFeederClient) *BusinessMetrics {
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: metricsPrefix + "aztec_feeder_tips_age_seconds",
		Help: "Seconds since the last successful chain-tips poll (node_getChainTips, or node_getL2Tips on pre-v5 nodes). -1 if never succeeded.",
	}, func() float64 {
		age := feeder.GetTipsAge()
		if age < 0 {
			return -1
		}
		return age.Seconds()
	})

	return &BusinessMetrics{
		ActiveNodeCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name: metricsPrefix + "active_node_count",
			Help: "Number of currently active nodes discovered in the latest crawl",
		}),
		SyncStatusTotal: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: metricsPrefix + "sync_status_total",
			Help: "Number of nodes by sync status in the latest crawl",
		}, []string{"status"}),
		LastCrawlAge: promauto.NewGauge(prometheus.GaugeOpts{
			Name: metricsPrefix + "last_crawl_age_seconds",
			Help: "Seconds since the last successfully finished crawl",
		}),
		CrawlSuccessTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: metricsPrefix + "crawl_success_total",
			Help: "Total number of successful crawls completed",
		}),
		logger: logger,
		repo:   r,
	}
}

// UpdateMetrics fetches latest data from the database and updates all metrics
// This is used at startup to initialise metrics from the most recent crawl
func (m *BusinessMetrics) UpdateMetrics(ctx context.Context) {
	crawlID, err := m.repo.GetLatestFinishedCrawlID(ctx)
	if err != nil {
		m.logger.Error("Failed to get latest finished crawl ID for metrics", "error", err)
		return
	}
	if crawlID == 0 {
		m.logger.Warn("No finished crawl found for metrics initialisation")
		return
	}

	m.updateActiveNodeCount(ctx, crawlID)
	m.updateSyncStatus(ctx, crawlID)
	m.updateLastCrawlAge(ctx)
}

// StartAgeUpdater starts a background goroutine that continuously updates the last crawl age
func (m *BusinessMetrics) StartAgeUpdater(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(ageUpdateInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.mu.RLock()
				lastTime := m.lastCrawlTime
				m.mu.RUnlock()

				if !lastTime.IsZero() {
					m.LastCrawlAge.Set(time.Since(lastTime).Seconds())
				}
			}
		}
	}()
}

// OnCrawlSuccess should be called after each successful crawl to update all metrics
// and increment the crawl success counter
func (m *BusinessMetrics) OnCrawlSuccess(ctx context.Context, crawl *coretypes.Crawl) {
	m.updateActiveNodeCount(ctx, crawl.ID)
	m.updateSyncStatus(ctx, crawl.ID)

	// Update last crawl time from the crawl data
	m.mu.Lock()
	m.lastCrawlTime = crawl.FinishedAt
	m.mu.Unlock()
	m.LastCrawlAge.Set(time.Since(crawl.FinishedAt).Seconds())

	m.CrawlSuccessTotal.Inc()
}

func (m *BusinessMetrics) updateActiveNodeCount(ctx context.Context, crawlID int) {
	count, err := m.repo.GetTotalCountByCrawlID(ctx, crawlID)
	if err != nil {
		m.logger.Error("Failed to get active node count for metrics", "error", err, "crawlID", crawlID)
		return
	}
	m.ActiveNodeCount.Set(float64(count))
	m.logger.Debug("Updated active_node_count metric", "value", count, "crawlID", crawlID)
}

func (m *BusinessMetrics) updateSyncStatus(ctx context.Context, crawlID int) {
	counts, err := m.repo.GetSyncStatusCountByCrawlID(ctx, crawlID)
	if err != nil {
		m.logger.Error("Failed to get sync status counts for metrics", "error", err, "crawlID", crawlID)
		return
	}
	if counts == nil {
		m.logger.Warn("No sync status counts available", "crawlID", crawlID)
		return
	}

	m.SyncStatusTotal.WithLabelValues("synced").Set(counts["synced"])
	m.SyncStatusTotal.WithLabelValues("not_synced").Set(counts["not_synced"])
	m.SyncStatusTotal.WithLabelValues("unknown").Set(counts["unknown"])
	m.logger.Debug("Updated sync_status_total metric",
		"synced", counts["synced"],
		"not_synced", counts["not_synced"],
		"unknown", counts["unknown"],
		"crawlID", crawlID)
}

func (m *BusinessMetrics) updateLastCrawlAge(ctx context.Context) {
	timestamp, err := m.repo.GetLastFinishedCrawlTimestamp(ctx)
	if err != nil {
		m.logger.Error("Failed to get last finished crawl timestamp for metrics", "error", err)
		return
	}
	if timestamp.IsZero() {
		m.logger.Warn("No finished crawl found")
		return
	}

	m.mu.Lock()
	m.lastCrawlTime = timestamp
	m.mu.Unlock()

	m.LastCrawlAge.Set(time.Since(timestamp).Seconds())
	m.logger.Debug("Updated last_crawl_age metric", "timestamp", timestamp)
}
