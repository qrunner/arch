// Package scheduler implements a cron-like job scheduler for periodic
// collection runs per source.
package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/qrunner/arch/internal/collector"
	"github.com/qrunner/arch/internal/model"
	"github.com/qrunner/arch/internal/reconciler"
	"go.uber.org/zap"
)

// Scheduler manages periodic collection jobs for each registered collector.
type Scheduler struct {
	registry   *collector.Registry
	reconciler *reconciler.Reconciler
	logger     *zap.Logger
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// New creates a new Scheduler.
func New(registry *collector.Registry, rec *reconciler.Reconciler, logger *zap.Logger) *Scheduler {
	return &Scheduler{
		registry:   registry,
		reconciler: rec,
		logger:     logger,
	}
}

// Start begins running collection jobs on their configured intervals.
func (s *Scheduler) Start(ctx context.Context) {
	ctx, s.cancel = context.WithCancel(ctx)

	for _, name := range s.registry.Names() {
		c, cfg, err := s.registry.Get(name)
		if err != nil {
			s.logger.Error("getting collector from registry", zap.String("name", name), zap.Error(err))
			continue
		}
		if !cfg.Enabled {
			s.logger.Info("collector disabled, skipping", zap.String("name", name))
			continue
		}

		s.wg.Add(1)
		go s.runCollectorLoop(ctx, c, cfg)
	}

	s.logger.Info("scheduler started")
}

// Stop gracefully stops all collection jobs.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	s.logger.Info("scheduler stopped")
}

func (s *Scheduler) runCollectorLoop(ctx context.Context, c collector.Collector, cfg model.CollectorConfig) {
	defer s.wg.Done()

	interval := cfg.Interval
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on start
	s.runCollector(ctx, c, cfg)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runCollector(ctx, c, cfg)
		}
	}
}

func (s *Scheduler) runCollector(ctx context.Context, c collector.Collector, cfg model.CollectorConfig) {
	s.logger.Info("running collector", zap.String("name", c.Name()))

	result, err := c.Collect(ctx, cfg)
	if err != nil {
		s.logger.Error("collector failed",
			zap.String("name", c.Name()),
			zap.Error(err),
		)
		return
	}

	s.logger.Info("collector completed",
		zap.String("name", c.Name()),
		zap.Int("assets", len(result.Assets)),
		zap.Int("relationships", len(result.Relationships)),
	)

	if err := s.reconciler.Reconcile(ctx, result); err != nil {
		s.logger.Error("reconciliation failed",
			zap.String("name", c.Name()),
			zap.Error(err),
		)
	}
}

// RunCollectorNow triggers an immediate collection for the named collector.
func (s *Scheduler) RunCollectorNow(ctx context.Context, name string) error {
	c, cfg, err := s.registry.Get(name)
	if err != nil {
		return err
	}
	go s.runCollector(ctx, c, cfg)
	return nil
}
