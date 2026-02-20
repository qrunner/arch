// Package nmap implements the Collector interface for network scanning
// using Nmap XML output parsing.
package nmap

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/qrunner/arch/internal/collector"
	"github.com/qrunner/arch/internal/model"
)

// Collector implements collector.Collector for network scanning via Nmap.
type Collector struct{}

// New creates a new Nmap collector.
func New() *Collector {
	return &Collector{}
}

// Name returns the collector identifier.
func (c *Collector) Name() string {
	return "nmap"
}

// Collect reads Nmap XML output and converts discovered hosts to assets.
func (c *Collector) Collect(ctx context.Context, cfg model.CollectorConfig) (*collector.CollectResult, error) {
	xmlPath := cfg.Settings["xml_path"]
	if xmlPath == "" {
		return nil, fmt.Errorf("nmap collector requires xml_path setting")
	}

	data, err := os.ReadFile(xmlPath)
	if err != nil {
		return nil, fmt.Errorf("reading nmap xml: %w", err)
	}

	var nmapRun nmapRunXML
	if err := xml.Unmarshal(data, &nmapRun); err != nil {
		return nil, fmt.Errorf("parsing nmap xml: %w", err)
	}

	result := &collector.CollectResult{}
	for _, host := range nmapRun.Hosts {
		if host.Status.State != "up" {
			continue
		}

		var ips []string
		var hostname string
		for _, addr := range host.Addresses {
			if addr.AddrType == "ipv4" || addr.AddrType == "ipv6" {
				ips = append(ips, addr.Addr)
			}
		}
		if len(host.Hostnames) > 0 {
			hostname = host.Hostnames[0].Name
		}

		externalID := ""
		if len(ips) > 0 {
			externalID = ips[0]
		}

		asset := *model.NewAsset(externalID, "nmap", "host", hostname)
		asset.IPAddresses = ips
		if hostname != "" {
			asset.FQDN = &hostname
		}

		// Store port and OS info as attributes
		attrs := map[string]any{
			"ports":     extractPorts(host.Ports),
			"os_match":  extractOS(host.OS),
		}
		asset.Attributes, _ = json.Marshal(attrs)

		result.Assets = append(result.Assets, asset)
	}

	return result, nil
}

// --- Nmap XML structs ---

type nmapRunXML struct {
	XMLName xml.Name   `xml:"nmaprun"`
	Hosts   []hostXML  `xml:"host"`
}

type hostXML struct {
	Status    statusXML    `xml:"status"`
	Addresses []addressXML `xml:"address"`
	Hostnames []hostnameXML `xml:"hostnames>hostname"`
	Ports     []portXML    `xml:"ports>port"`
	OS        osXML        `xml:"os"`
}

type statusXML struct {
	State string `xml:"state,attr"`
}

type addressXML struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

type hostnameXML struct {
	Name string `xml:"name,attr"`
}

type portXML struct {
	Protocol string     `xml:"protocol,attr"`
	PortID   string     `xml:"portid,attr"`
	State    statusXML  `xml:"state"`
	Service  serviceXML `xml:"service"`
}

type serviceXML struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr"`
	Version string `xml:"version,attr"`
}

type osXML struct {
	Matches []osMatchXML `xml:"osmatch"`
}

type osMatchXML struct {
	Name     string `xml:"name,attr"`
	Accuracy string `xml:"accuracy,attr"`
}

func extractPorts(ports []portXML) []map[string]string {
	var result []map[string]string
	for _, p := range ports {
		if p.State.State == "open" {
			result = append(result, map[string]string{
				"port":     p.PortID,
				"protocol": p.Protocol,
				"service":  p.Service.Name,
				"product":  p.Service.Product,
				"version":  p.Service.Version,
			})
		}
	}
	return result
}

func extractOS(os osXML) string {
	if len(os.Matches) > 0 {
		return os.Matches[0].Name
	}
	return ""
}
