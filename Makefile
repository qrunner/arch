.PHONY: build run test lint clean migrate-up migrate-down docker-build docker-up docker-down

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
BINARY_SERVER=bin/server
BINARY_COLLECTOR=bin/collector
BINARY_MCP=bin/mcp

# Build all binaries
build:
	$(GOBUILD) -o $(BINARY_SERVER) ./cmd/server
	$(GOBUILD) -o $(BINARY_COLLECTOR) ./cmd/collector
	$(GOBUILD) -o $(BINARY_MCP) ./cmd/mcp

# Build individual binaries
build-server:
	$(GOBUILD) -o $(BINARY_SERVER) ./cmd/server

build-collector:
	$(GOBUILD) -o $(BINARY_COLLECTOR) ./cmd/collector

build-mcp:
	$(GOBUILD) -o $(BINARY_MCP) ./cmd/mcp

# Run the API server
run-server:
	$(GOBUILD) -o $(BINARY_SERVER) ./cmd/server && ./$(BINARY_SERVER)

# Run the collector worker
run-collector:
	$(GOBUILD) -o $(BINARY_COLLECTOR) ./cmd/collector && ./$(BINARY_COLLECTOR)

# Run all tests
test:
	$(GOTEST) -v -race ./...

# Run tests with coverage
test-cover:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	$(GOVET) ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Database migrations
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

# Docker operations
docker-build:
	docker build -f deploy/docker/Dockerfile.server -t arch-server .
	docker build -f deploy/docker/Dockerfile.collector -t arch-collector .
	docker build -f deploy/docker/Dockerfile.web -t arch-web ./web

docker-up:
	docker compose -f deploy/docker-compose.yml up -d

docker-down:
	docker compose -f deploy/docker-compose.yml down

# Frontend
web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

# Full stack development setup
dev-infra:
	docker compose -f deploy/docker-compose.yml up -d postgres neo4j nats

# Generate Go mocks (for testing)
generate:
	$(GOCMD) generate ./...
