package model

import "time"

// CollectorConfig holds the configuration for a single collector source.
type CollectorConfig struct {
	Name     string            `json:"name" mapstructure:"name"`
	Type     string            `json:"type" mapstructure:"type"`
	Enabled  bool              `json:"enabled" mapstructure:"enabled"`
	Interval time.Duration     `json:"interval" mapstructure:"interval"`
	Settings map[string]string `json:"settings" mapstructure:"settings"`
}

// CollectorStatus tracks the runtime state of a collector.
type CollectorStatus struct {
	Name        string    `json:"name"`
	LastRun     time.Time `json:"last_run"`
	LastSuccess time.Time `json:"last_success"`
	LastError   string    `json:"last_error,omitempty"`
	Running     bool      `json:"running"`
	AssetCount  int       `json:"asset_count"`
}
