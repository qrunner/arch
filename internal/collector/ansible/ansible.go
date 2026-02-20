// Package ansible implements the Collector interface for Ansible inventory.
package ansible

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/qrunner/arch/internal/collector"
	"github.com/qrunner/arch/internal/model"
)

// Collector implements collector.Collector for Ansible inventory files.
type Collector struct{}

// New creates a new Ansible collector.
func New() *Collector {
	return &Collector{}
}

// Name returns the collector identifier.
func (c *Collector) Name() string {
	return "ansible"
}

// Collect reads Ansible inventory JSON and facts cache to discover hosts.
func (c *Collector) Collect(ctx context.Context, cfg model.CollectorConfig) (*collector.CollectResult, error) {
	inventoryPath := cfg.Settings["inventory_path"]
	if inventoryPath == "" {
		return nil, fmt.Errorf("ansible collector requires inventory_path setting")
	}

	data, err := os.ReadFile(inventoryPath)
	if err != nil {
		return nil, fmt.Errorf("reading ansible inventory: %w", err)
	}

	var inventory map[string]json.RawMessage
	if err := json.Unmarshal(data, &inventory); err != nil {
		return nil, fmt.Errorf("parsing ansible inventory: %w", err)
	}

	// TODO: implement full Ansible inventory parsing
	// 1. Parse inventory JSON groups and hosts
	// 2. Read facts cache for additional host data
	// 3. Create asset_type: "host" for each inventory host
	// 4. Create asset_type: "group" for each inventory group
	// 5. Build MEMBER_OF relationships (host -> group)

	result := &collector.CollectResult{}
	return result, nil
}
