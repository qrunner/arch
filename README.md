# IT Asset Inventory System

A system for automatic discovery, inventory, and relationship tracking of IT assets across multiple infrastructure sources. Built with Go backend, React frontend, Neo4j for relationship graphs, and PostgreSQL for structured data.

## Architecture

- **Backend**: Go with chi router, PostgreSQL (pgx), Neo4j, NATS event bus
- **Frontend**: React 18 + TypeScript + Vite + TailwindCSS + Cytoscape.js
- **Storage**: PostgreSQL (structured data), Neo4j (relationship graph), NATS (events)

## Data Sources

| Source | Data Collected | Protocol/API |
|--------|---------------|-------------|
| Network Scanner | Hosts, open ports, OS fingerprints | Nmap XML output |
| VMware vCenter | VMs, ESXi hosts, clusters, datastores | govmomi (vSphere API) |
| Zabbix | Monitored hosts, metrics, triggers | Zabbix REST API |
| Ansible | Inventory hosts, groups, facts | Ansible inventory JSON |
| Kubernetes | Nodes, pods, services, deployments | client-go (K8s API) |
| Citrix NetScaler | Virtual servers, service groups, backends | NITRO REST API |

## Project Structure

```
cmd/
  server/          API server entrypoint
  collector/       Collector worker entrypoint
  mcp/             MCP server entrypoint (Phase 6)
internal/
  api/             HTTP handlers, routes, middleware
  collector/       Collector engine and source implementations
  reconciler/      Asset matching and change detection
  store/           Database access layer (postgres + neo4j)
  scheduler/       Job scheduling
  notifier/        Alerts and notifications
  config/          Configuration loading
  model/           Domain models
web/               React frontend
migrations/        PostgreSQL migrations
deploy/            Docker and Kubernetes deployment configs
configs/           Default configuration files
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.22+ (for local development)
- Node.js 20+ (for frontend development)

### Start Infrastructure

```bash
# Start PostgreSQL, Neo4j, and NATS
make dev-infra
```

### Run Migrations

```bash
export DATABASE_URL="postgres://arch:arch@localhost:5432/arch?sslmode=disable"
make migrate-up
```

### Run Backend

```bash
make run-server      # API server
make run-collector   # Collector worker
```

### Run Frontend

```bash
make web-install
make web-dev
```

### Full Stack (Docker)

```bash
make docker-build
make docker-up
```

The API server runs on `http://localhost:8080` and the frontend on `http://localhost:3000`.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /healthz | Health check |
| GET | /api/v1/assets | List assets (filter, paginate) |
| GET | /api/v1/assets/:id | Get asset details |
| POST | /api/v1/assets | Create asset |
| PUT | /api/v1/assets/:id | Update asset |
| DELETE | /api/v1/assets/:id | Delete asset |
| GET | /api/v1/assets/:id/history | Change history |
| GET | /api/v1/assets/:id/relationships | Relationships |
| GET | /api/v1/graph/dependencies/:id | Dependency subgraph |
| GET | /api/v1/graph/impact/:id | Impact/blast radius |
| GET | /api/v1/collectors | List collectors |
| POST | /api/v1/collectors/:name/run | Trigger collection |
| GET | /api/v1/changes | Recent changes |
| GET | /api/v1/dashboard/stats | Dashboard stats |
| GET | /api/v1/events | SSE event stream |

## Configuration

Configuration is loaded from YAML files and environment variables:

- File locations: `./configs/config.yaml`, `/etc/arch/config.yaml`
- Environment prefix: `ARCH_` (e.g., `ARCH_DATABASE_HOST=postgres`)

See [configs/config.yaml](configs/config.yaml) for the full configuration reference.

## Technology Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go 1.22+, chi router |
| PostgreSQL driver | pgx v5 |
| Neo4j driver | neo4j-go-driver v5 |
| NATS client | nats.go |
| Config | Viper |
| Frontend | React 18, TypeScript, Vite |
| CSS | TailwindCSS |
| Graph visualization | Cytoscape.js |
| Deployment | Docker, Docker Compose |
