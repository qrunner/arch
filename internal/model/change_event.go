package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ChangeAction describes what kind of change occurred.
type ChangeAction string

const (
	ChangeActionCreated ChangeAction = "asset.created"
	ChangeActionUpdated ChangeAction = "asset.updated"
	ChangeActionRemoved ChangeAction = "asset.removed"
	ChangeActionRelChanged ChangeAction = "relationship.changed"
)

// ChangeEvent records a change to an asset or relationship.
type ChangeEvent struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	AssetID   uuid.UUID       `json:"asset_id" db:"asset_id"`
	Action    ChangeAction    `json:"action" db:"action"`
	Source    string          `json:"source" db:"source"`
	Diff      json.RawMessage `json:"diff,omitempty" db:"diff"`
	Timestamp time.Time       `json:"timestamp" db:"timestamp"`
}

// NewChangeEvent creates a new ChangeEvent with generated UUID and current timestamp.
func NewChangeEvent(assetID uuid.UUID, action ChangeAction, source string, diff json.RawMessage) *ChangeEvent {
	return &ChangeEvent{
		ID:        uuid.New(),
		AssetID:   assetID,
		Action:    action,
		Source:    source,
		Diff:      diff,
		Timestamp: time.Now().UTC(),
	}
}
