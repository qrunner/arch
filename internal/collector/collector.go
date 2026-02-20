// Package collector defines the Collector interface and the collector registry
// that manages all source collectors.
package collector

import (
	"context"
	"fmt"
	"sync"

	"github.com/qrunner/arch/internal/model"
	"go.uber.org/zap"
)

// CollectResult holds the assets and relationships discovered by a collector.
type CollectResult struct {
	Assets        []model.Asset
	Relationships []model.Relationship
}

// Collector is the interface that all source collectors must implement.
type Collector interface {
	// Name returns the unique identifier for this collector (e.g., "vmware", "k8s").
	Name() string
	// Collect connects to the source and returns discovered assets and relationships.
	Collect(ctx context.Context, cfg model.CollectorConfig) (*CollectResult, error)
}

// Registry manages registered collectors and their configurations.
type Registry struct {
	mu         sync.RWMutex
	collectors map[string]Collector
	configs    map[string]model.CollectorConfig
	statuses   map[string]*model.CollectorStatus
	logger     *zap.Logger
}

// NewRegistry creates a new collector registry.
func NewRegistry(logger *zap.Logger) *Registry {
	return &Registry{
		collectors: make(map[string]Collector),
		configs:    make(map[string]model.CollectorConfig),
		statuses:   make(map[string]*model.CollectorStatus),
		logger:     logger,
	}
}

// Register adds a collector to the registry.
func (r *Registry) Register(c Collector, cfg model.CollectorConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.collectors[c.Name()] = c
	r.configs[c.Name()] = cfg
	r.statuses[c.Name()] = &model.CollectorStatus{Name: c.Name()}
	r.logger.Info("registered collector", zap.String("name", c.Name()))
}

// Get returns a collector by name.
func (r *Registry) Get(name string) (Collector, model.CollectorConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.collectors[name]
	if !ok {
		return nil, model.CollectorConfig{}, fmt.Errorf("collector %q not found", name)
	}
	return c, r.configs[name], nil
}

// List returns all registered collector names and their statuses.
func (r *Registry) List() []*model.CollectorStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*model.CollectorStatus, 0, len(r.statuses))
	for _, s := range r.statuses {
		result = append(result, s)
	}
	return result
}

// Names returns all registered collector names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.collectors))
	for name := range r.collectors {
		names = append(names, name)
	}
	return names
}
