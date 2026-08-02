package core

import (
	"context"
	"log/slog"
	"time"

	"github.com/NethermindEth/aztec-p2p-explorer/core/types"
)

type Engine struct {
	executor          Executor
	processor         Processor
	logger            *slog.Logger
	interval          time.Duration
	crawlFinishedChan chan *types.Crawl
}

func NewEngine(
	executor Executor,
	processor Processor,
	logger *slog.Logger,
	crawlFinishedChan chan *types.Crawl,
	options ...func(*Engine),
) *Engine {
	engine := &Engine{
		executor:          executor,
		processor:         processor,
		logger:            logger,
		interval:          1 * time.Minute,
		crawlFinishedChan: crawlFinishedChan,
	}

	for _, option := range options {
		option(engine)
	}

	return engine
}

func WithInterval(interval time.Duration) func(*Engine) {
	return func(e *Engine) {
		e.interval = interval
	}
}

func (e *Engine) Start(ctx context.Context) {
	e.logger.Info("Starting engine")

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("Stopping engine")
			return
		default:
			err := e.executor.Execute(ctx)
			if err != nil {
				e.logger.Error("Failed to execute", "err", err)
				return
			}

			crawl, err := e.processor.Process(ctx)
			if err != nil {
				e.logger.Error("Failed to process", "err", err)
			}
			if crawl != nil {
				e.crawlFinishedChan <- crawl
			}

			timer := time.NewTimer(e.interval)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				// Continue to the next iteration
			}
		}
	}
}
