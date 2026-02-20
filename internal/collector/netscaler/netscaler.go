// Package netscaler implements the Collector interface for Citrix NetScaler
// using the NITRO REST API.
package netscaler

import (
	"context"
	"fmt"

	"github.com/qrunner/arch/internal/collector"
	"github.com/qrunner/arch/internal/model"
)

// Collector implements collector.Collector for Citrix NetScaler.
type Collector struct{}

// New creates a new NetScaler collector.
func New() *Collector {
	return &Collector{}
}

// Name returns the collector identifier.
func (c *Collector) Name() string {
	return "netscaler"
}

// Collect connects to the NetScaler NITRO API and discovers virtual servers,
// service groups, backends, and SSL certificates.
func (c *Collector) Collect(ctx context.Context, cfg model.CollectorConfig) (*collector.CollectResult, error) {
	nitroURL := cfg.Settings["url"]
	if nitroURL == "" {
		return nil, fmt.Errorf("netscaler collector requires url setting")
	}

	// TODO: implement NITRO API client
	// 1. Authenticate to NITRO REST API
	// 2. Fetch virtual servers (lbvserver) -> asset_type: "vserver"
	// 3. Fetch service groups -> asset_type: "service_group"
	// 4. Fetch backend servers -> asset_type: "backend"
	// 5. Fetch SSL certificates -> asset_type: "ssl_cert"
	// 6. Build LOAD_BALANCES relationships (vserver -> backend)
	// 7. Build MEMBER_OF relationships (backend -> service_group)

	result := &collector.CollectResult{}
	return result, nil
}
