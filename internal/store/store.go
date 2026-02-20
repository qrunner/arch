// Package store defines the data access interfaces for the application.
package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/qrunner/arch/internal/model"
)

// AssetFilter holds query parameters for listing assets.
type AssetFilter struct {
	Source    string
	AssetType string
	Status    string
	Search    string
	Limit     int
	Offset    int
}

// AssetStore defines operations on the canonical asset records in PostgreSQL.
type AssetStore interface {
	// Create inserts a new asset.
	Create(ctx context.Context, asset *model.Asset) error
	// GetByID retrieves an asset by its UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*model.Asset, error)
	// GetByExternalID retrieves an asset by source and external ID.
	GetByExternalID(ctx context.Context, source, externalID string) (*model.Asset, error)
	// List returns assets matching the given filter.
	List(ctx context.Context, filter AssetFilter) ([]*model.Asset, int, error)
	// Update persists changes to an existing asset.
	Update(ctx context.Context, asset *model.Asset) error
	// Delete removes an asset by ID.
	Delete(ctx context.Context, id uuid.UUID) error
	// FindByIP looks up assets that have the given IP address.
	FindByIP(ctx context.Context, ip string) ([]*model.Asset, error)
	// FindByFQDN looks up assets that match the given FQDN.
	FindByFQDN(ctx context.Context, fqdn string) ([]*model.Asset, error)
}

// ChangeEventStore defines operations on change history records.
type ChangeEventStore interface {
	// Create inserts a new change event.
	Create(ctx context.Context, event *model.ChangeEvent) error
	// ListByAssetID returns change events for a given asset.
	ListByAssetID(ctx context.Context, assetID uuid.UUID, limit, offset int) ([]*model.ChangeEvent, int, error)
	// ListRecent returns the most recent change events across all assets.
	ListRecent(ctx context.Context, limit, offset int) ([]*model.ChangeEvent, int, error)
}

// GraphStore defines operations on the Neo4j relationship graph.
type GraphStore interface {
	// UpsertNode creates or updates an asset node in the graph.
	UpsertNode(ctx context.Context, asset *model.Asset) error
	// DeleteNode removes an asset node from the graph.
	DeleteNode(ctx context.Context, id uuid.UUID) error
	// UpsertRelationship creates or updates a relationship edge.
	UpsertRelationship(ctx context.Context, rel *model.Relationship) error
	// DeleteRelationship removes a relationship edge.
	DeleteRelationship(ctx context.Context, id uuid.UUID) error
	// GetRelationships returns all relationships for an asset.
	GetRelationships(ctx context.Context, assetID uuid.UUID) ([]*model.Relationship, error)
	// GetDependencyGraph returns the subgraph of dependencies starting from an asset.
	GetDependencyGraph(ctx context.Context, assetID uuid.UUID, depth int) ([]*model.Asset, []*model.Relationship, error)
	// GetImpactGraph returns the subgraph of assets impacted if the given asset fails.
	GetImpactGraph(ctx context.Context, assetID uuid.UUID, depth int) ([]*model.Asset, []*model.Relationship, error)
}
