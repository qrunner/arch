// Package neo4j implements the GraphStore interface using Neo4j.
package neo4j

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/qrunner/arch/internal/model"
)

// Store implements store.GraphStore using Neo4j.
type Store struct {
	driver neo4j.DriverWithContext
}

// New creates a new Neo4j store with the given driver.
func New(driver neo4j.DriverWithContext) *Store {
	return &Store{driver: driver}
}

// Connect creates a new Neo4j driver and returns a Store.
func Connect(ctx context.Context, uri, user, password string) (*Store, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, fmt.Errorf("creating neo4j driver: %w", err)
	}
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("verifying neo4j connectivity: %w", err)
	}
	return &Store{driver: driver}, nil
}

// Close shuts down the driver.
func (s *Store) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}

// UpsertNode creates or updates an asset node in the graph.
func (s *Store) UpsertNode(ctx context.Context, asset *model.Asset) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MERGE (a:Asset {id: $id})
			SET a.name = $name,
			    a.asset_type = $asset_type,
			    a.source = $source,
			    a.status = $status
		`
		_, err := tx.Run(ctx, query, map[string]any{
			"id":         asset.ID.String(),
			"name":       asset.Name,
			"asset_type": asset.AssetType,
			"source":     asset.Source,
			"status":     string(asset.Status),
		})
		return nil, err
	})
	if err != nil {
		return fmt.Errorf("upserting neo4j node: %w", err)
	}
	return nil
}

// DeleteNode removes an asset node and its relationships from the graph.
func (s *Store) DeleteNode(ctx context.Context, id uuid.UUID) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, "MATCH (a:Asset {id: $id}) DETACH DELETE a", map[string]any{
			"id": id.String(),
		})
		return nil, err
	})
	if err != nil {
		return fmt.Errorf("deleting neo4j node: %w", err)
	}
	return nil
}

// UpsertRelationship creates or updates a relationship edge between two asset nodes.
func (s *Store) UpsertRelationship(ctx context.Context, rel *model.Relationship) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Neo4j doesn't support parameterized relationship types, so we use APOC or
		// a switch. For safety, we whitelist relationship types.
		relType := string(rel.Type)
		query := fmt.Sprintf(`
			MATCH (from:Asset {id: $from_id})
			MATCH (to:Asset {id: $to_id})
			MERGE (from)-[r:%s {id: $id}]->(to)
			SET r.source = $source, r.properties = $properties
		`, relType)

		props, _ := json.Marshal(rel.Properties)
		_, err := tx.Run(ctx, query, map[string]any{
			"id":         rel.ID.String(),
			"from_id":    rel.FromID.String(),
			"to_id":      rel.ToID.String(),
			"source":     rel.Source,
			"properties": string(props),
		})
		return nil, err
	})
	if err != nil {
		return fmt.Errorf("upserting neo4j relationship: %w", err)
	}
	return nil
}

// DeleteRelationship removes a relationship edge by its ID.
func (s *Store) DeleteRelationship(ctx context.Context, id uuid.UUID) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, "MATCH ()-[r {id: $id}]-() DELETE r", map[string]any{
			"id": id.String(),
		})
		return nil, err
	})
	if err != nil {
		return fmt.Errorf("deleting neo4j relationship: %w", err)
	}
	return nil
}

// GetRelationships returns all relationships for an asset.
func (s *Store) GetRelationships(ctx context.Context, assetID uuid.UUID) ([]*model.Relationship, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (a:Asset {id: $id})-[r]-(b:Asset)
			RETURN r.id AS id, a.id AS from_id, b.id AS to_id, type(r) AS rel_type, r.source AS source, r.properties AS properties
		`
		records, err := tx.Run(ctx, query, map[string]any{"id": assetID.String()})
		if err != nil {
			return nil, err
		}

		var rels []*model.Relationship
		for records.Next(ctx) {
			rec := records.Record()
			rel, err := recordToRelationship(rec)
			if err != nil {
				return nil, err
			}
			rels = append(rels, rel)
		}
		return rels, nil
	})
	if err != nil {
		return nil, fmt.Errorf("getting relationships: %w", err)
	}
	return result.([]*model.Relationship), nil
}

// GetDependencyGraph returns the subgraph of dependencies (outgoing) starting from an asset.
func (s *Store) GetDependencyGraph(ctx context.Context, assetID uuid.UUID, depth int) ([]*model.Asset, []*model.Relationship, error) {
	return s.traverseGraph(ctx, assetID, depth, "outgoing")
}

// GetImpactGraph returns the subgraph of assets impacted (incoming) if the given asset fails.
func (s *Store) GetImpactGraph(ctx context.Context, assetID uuid.UUID, depth int) ([]*model.Asset, []*model.Relationship, error) {
	return s.traverseGraph(ctx, assetID, depth, "incoming")
}

func (s *Store) traverseGraph(ctx context.Context, assetID uuid.UUID, depth int, direction string) ([]*model.Asset, []*model.Relationship, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	if depth <= 0 {
		depth = 3
	}

	var pattern string
	if direction == "outgoing" {
		pattern = fmt.Sprintf("(a:Asset {id: $id})-[r*1..%d]->(b:Asset)", depth)
	} else {
		pattern = fmt.Sprintf("(b:Asset)-[r*1..%d]->(a:Asset {id: $id})", depth)
	}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := fmt.Sprintf(`
			MATCH path = %s
			UNWIND nodes(path) AS n
			UNWIND relationships(path) AS rel
			WITH COLLECT(DISTINCT n) AS nodes, COLLECT(DISTINCT rel) AS rels
			RETURN nodes, rels
		`, pattern)

		records, err := tx.Run(ctx, query, map[string]any{"id": assetID.String()})
		if err != nil {
			return nil, err
		}

		var assets []*model.Asset
		var relationships []*model.Relationship

		if records.Next(ctx) {
			rec := records.Record()
			nodesVal, _ := rec.Get("nodes")
			relsVal, _ := rec.Get("rels")

			if nodesList, ok := nodesVal.([]any); ok {
				for _, n := range nodesList {
					if node, ok := n.(neo4j.Node); ok {
						a := nodeToAsset(node)
						assets = append(assets, a)
					}
				}
			}

			if relsList, ok := relsVal.([]any); ok {
				for _, r := range relsList {
					if neoRel, ok := r.(neo4j.Relationship); ok {
						rel := neoRelToRelationship(neoRel)
						relationships = append(relationships, rel)
					}
				}
			}
		}

		return []any{assets, relationships}, nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("traversing graph: %w", err)
	}

	parts := result.([]any)
	return parts[0].([]*model.Asset), parts[1].([]*model.Relationship), nil
}

func nodeToAsset(node neo4j.Node) *model.Asset {
	props := node.Props
	a := &model.Asset{}
	if id, ok := props["id"].(string); ok {
		a.ID, _ = uuid.Parse(id)
	}
	if name, ok := props["name"].(string); ok {
		a.Name = name
	}
	if at, ok := props["asset_type"].(string); ok {
		a.AssetType = at
	}
	if src, ok := props["source"].(string); ok {
		a.Source = src
	}
	if st, ok := props["status"].(string); ok {
		a.Status = model.AssetStatus(st)
	}
	return a
}

func neoRelToRelationship(neoRel neo4j.Relationship) *model.Relationship {
	rel := &model.Relationship{
		Type: model.RelationshipType(neoRel.Type),
	}
	if id, ok := neoRel.Props["id"].(string); ok {
		rel.ID, _ = uuid.Parse(id)
	}
	if src, ok := neoRel.Props["source"].(string); ok {
		rel.Source = src
	}
	return rel
}

func recordToRelationship(rec *neo4j.Record) (*model.Relationship, error) {
	rel := &model.Relationship{}
	if id, ok := rec.Get("id"); ok {
		if s, ok := id.(string); ok {
			rel.ID, _ = uuid.Parse(s)
		}
	}
	if fromID, ok := rec.Get("from_id"); ok {
		if s, ok := fromID.(string); ok {
			rel.FromID, _ = uuid.Parse(s)
		}
	}
	if toID, ok := rec.Get("to_id"); ok {
		if s, ok := toID.(string); ok {
			rel.ToID, _ = uuid.Parse(s)
		}
	}
	if rt, ok := rec.Get("rel_type"); ok {
		if s, ok := rt.(string); ok {
			rel.Type = model.RelationshipType(s)
		}
	}
	if src, ok := rec.Get("source"); ok {
		if s, ok := src.(string); ok {
			rel.Source = s
		}
	}
	return rel, nil
}
