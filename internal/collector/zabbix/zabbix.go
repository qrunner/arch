// Package zabbix implements the Collector interface for Zabbix monitoring.
package zabbix

import (
	"context"
	"fmt"

	"github.com/qrunner/arch/internal/collector"
	"github.com/qrunner/arch/internal/model"
)

// Collector implements collector.Collector for Zabbix.
type Collector struct{}

// New creates a new Zabbix collector.
func New() *Collector {
	return &Collector{}
}

// Name returns the collector identifier.
func (c *Collector) Name() string {
	return "zabbix"
}

// Collect connects to the Zabbix API and discovers monitored hosts and groups.
func (c *Collector) Collect(ctx context.Context, cfg model.CollectorConfig) (*collector.CollectResult, error) {
	apiURL := cfg.Settings["url"]
	if apiURL == "" {
		return nil, fmt.Errorf("zabbix collector requires url setting")
	}

	// TODO: implement Zabbix REST API client
	// 1. Authenticate to Zabbix API
	// 2. Fetch hosts -> asset_type: "host"
	// 3. Fetch host groups -> asset_type: "host_group"
	// 4. Build MEMBER_OF relationships (host -> host_group)
	// 5. Build MONITORS relationships (zabbix host -> target asset by IP match)

	result := &collector.CollectResult{}
	return result, nil
}
