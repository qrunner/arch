package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// RelationshipType defines the kind of relationship between two assets.
type RelationshipType string

const (
	RelRunsOn        RelationshipType = "RUNS_ON"
	RelDependsOn     RelationshipType = "DEPENDS_ON"
	RelMemberOf      RelationshipType = "MEMBER_OF"
	RelLoadBalances  RelationshipType = "LOAD_BALANCES"
	RelMonitors      RelationshipType = "MONITORS"
	RelConnectsTo    RelationshipType = "CONNECTS_TO"
	RelContains      RelationshipType = "CONTAINS"
)

// Relationship represents a directed edge between two assets in the graph.
type Relationship struct {
	ID         uuid.UUID        `json:"id"`
	FromID     uuid.UUID        `json:"from_id"`
	ToID       uuid.UUID        `json:"to_id"`
	Type       RelationshipType `json:"type"`
	Source     string           `json:"source"`
	Properties json.RawMessage  `json:"properties,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

// NewRelationship creates a new Relationship with a generated UUID.
func NewRelationship(fromID, toID uuid.UUID, relType RelationshipType, source string) *Relationship {
	now := time.Now().UTC()
	return &Relationship{
		ID:         uuid.New(),
		FromID:     fromID,
		ToID:       toID,
		Type:       relType,
		Source:     source,
		Properties: json.RawMessage("{}"),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
