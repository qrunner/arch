// Package model defines the core domain types for the IT Asset Inventory System.
package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AssetStatus represents the lifecycle state of an asset.
type AssetStatus string

const (
	AssetStatusActive  AssetStatus = "active"
	AssetStatusStale   AssetStatus = "stale"
	AssetStatusRemoved AssetStatus = "removed"
)

// Asset represents a discovered IT asset from any source.
type Asset struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	ExternalID  string          `json:"external_id" db:"external_id"`
	Source      string          `json:"source" db:"source"`
	AssetType   string          `json:"asset_type" db:"asset_type"`
	Name        string          `json:"name" db:"name"`
	FQDN        *string         `json:"fqdn,omitempty" db:"fqdn"`
	IPAddresses []string        `json:"ip_addresses" db:"ip_addresses"`
	Attributes  json.RawMessage `json:"attributes" db:"attributes"`
	FirstSeen   time.Time       `json:"first_seen" db:"first_seen"`
	LastSeen    time.Time       `json:"last_seen" db:"last_seen"`
	Status      AssetStatus     `json:"status" db:"status"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// NewAsset creates a new Asset with generated UUID and timestamps.
func NewAsset(externalID, source, assetType, name string) *Asset {
	now := time.Now().UTC()
	return &Asset{
		ID:          uuid.New(),
		ExternalID:  externalID,
		Source:      source,
		AssetType:   assetType,
		Name:        name,
		IPAddresses: []string{},
		Attributes:  json.RawMessage("{}"),
		FirstSeen:   now,
		LastSeen:    now,
		Status:      AssetStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
