// Package main is the entrypoint for the IT Asset Inventory API server.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/qrunner/arch/internal/api"
	"github.com/qrunner/arch/internal/config"
	"github.com/qrunner/arch/internal/store/postgres"
	neostore "github.com/qrunner/arch/internal/store/neo4j"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
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
	logger.Info("connected to PostgreSQL")

	// Connect to Neo4j
	var neo *neostore.Store
	neo, err = neostore.Connect(ctx, cfg.Neo4j.URI, cfg.Neo4j.User, cfg.Neo4j.Password)
	if err != nil {
		logger.Warn("failed to connect to Neo4j, graph features disabled", zap.Error(err))
	} else {
		defer neo.Close(ctx)
		logger.Info("connected to Neo4j")
	}

	// Create API server
	srv := api.NewServer(logger, pgStore, neo)

	httpServer := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      srv,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("starting API server", zap.String("addr", cfg.Server.Address()))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("server shutdown error", zap.Error(err))
	}

	logger.Info("server stopped")
	_ = fmt.Sprintf // silence unused import
}
