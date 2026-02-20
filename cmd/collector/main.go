// Package main is the entrypoint for the collector worker process.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/qrunner/arch/internal/collector"
	"github.com/qrunner/arch/internal/collector/nmap"
	"github.com/qrunner/arch/internal/collector/vmware"
	"github.com/qrunner/arch/internal/collector/zabbix"
	"github.com/qrunner/arch/internal/collector/ansible"
	"github.com/qrunner/arch/internal/collector/k8s"
	"github.com/qrunner/arch/internal/collector/netscaler"
	"github.com/qrunner/arch/internal/config"
	"github.com/qrunner/arch/internal/model"
	"github.com/qrunner/arch/internal/reconciler"
	"github.com/qrunner/arch/internal/scheduler"
	"github.com/qrunner/arch/internal/store/postgres"
	neostore "github.com/qrunner/arch/internal/store/neo4j"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.Load("")
	if err != nil {
		logger.Fatal("loading config", zap.Error(err))
	}

	ctx := context.Background()

	// Connect to PostgreSQL
	pgStore, err := postgres.Connect(ctx, cfg.Database.DSN())
	if err != nil {
		logger.Fatal("connecting to postgres", zap.Error(err))
	}
	defer pgStore.Close()

	// Connect to Neo4j
	var neo *neostore.Store
	neo, err = neostore.Connect(ctx, cfg.Neo4j.URI, cfg.Neo4j.User, cfg.Neo4j.Password)
	if err != nil {
		logger.Warn("neo4j unavailable", zap.Error(err))
	} else {
		defer neo.Close(ctx)
	}

	// Create reconciler (no event publisher for now)
	rec := reconciler.New(pgStore, neo, nil, logger)

	// Register collectors
	registry := collector.NewRegistry(logger)
	collectorMap := map[string]collector.Collector{
		"nmap":      nmap.New(),
		"vmware":    vmware.New(),
		"zabbix":    zabbix.New(),
		"ansible":   ansible.New(),
		"k8s":       k8s.New(),
		"netscaler": netscaler.New(),
	}

	for _, entry := range cfg.Collectors {
		c, ok := collectorMap[entry.Type]
		if !ok {
			logger.Warn("unknown collector type", zap.String("type", entry.Type))
			continue
		}

		interval, _ := time.ParseDuration(entry.Interval)
		collectorCfg := model.CollectorConfig{
			Name:     entry.Name,
			Type:     entry.Type,
			Enabled:  entry.Enabled,
			Interval: interval,
			Settings: entry.Settings,
		}
		registry.Register(c, collectorCfg)
	}

	// Start scheduler
	sched := scheduler.New(registry, rec, logger)
	sched.Start(ctx)

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down collector worker...")
	sched.Stop()
	logger.Info("collector worker stopped")
}
