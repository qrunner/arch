// Package api implements the HTTP REST API server.
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/qrunner/arch/internal/store/postgres"
	neostore "github.com/qrunner/arch/internal/store/neo4j"
	"go.uber.org/zap"
)

// Server holds the dependencies for the HTTP API.
type Server struct {
	router   chi.Router
	logger   *zap.Logger
	pgStore  *postgres.Store
	neoStore *neostore.Store
}

// NewServer creates a new API server with all routes configured.
func NewServer(logger *zap.Logger, pgStore *postgres.Store, neoStore *neostore.Store) *Server {
	s := &Server{
		logger:   logger,
		pgStore:  pgStore,
		neoStore: neoStore,
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(corsMiddleware)
	r.Use(jsonContentType)

	// Health check
	r.Get("/healthz", s.handleHealthCheck)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Assets
		r.Route("/assets", func(r chi.Router) {
			r.Get("/", s.handleListAssets)
			r.Post("/", s.handleCreateAsset)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", s.handleGetAsset)
				r.Put("/", s.handleUpdateAsset)
				r.Delete("/", s.handleDeleteAsset)
				r.Get("/history", s.handleGetAssetHistory)
				r.Get("/relationships", s.handleGetAssetRelationships)
			})
		})

		// Graph
		r.Route("/graph", func(r chi.Router) {
			r.Get("/dependencies/{id}", s.handleGetDependencyGraph)
			r.Get("/impact/{id}", s.handleGetImpactGraph)
		})

		// Collectors
		r.Route("/collectors", func(r chi.Router) {
			r.Get("/", s.handleListCollectors)
			r.Post("/{name}/run", s.handleTriggerCollector)
		})

		// Changes
		r.Get("/changes", s.handleListChanges)

		// Dashboard
		r.Get("/dashboard/stats", s.handleDashboardStats)

		// SSE events stream
		r.Get("/events", s.handleSSEEvents)
	})

	s.router = r
	return s
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Router returns the underlying chi router for testing.
func (s *Server) Router() chi.Router {
	return s.router
}
