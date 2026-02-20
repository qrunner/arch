// Package postgres implements the AssetStore and ChangeEventStore interfaces
// using PostgreSQL via pgx.
package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/qrunner/arch/internal/model"
	"github.com/qrunner/arch/internal/store"
)

// Store implements store.AssetStore and store.ChangeEventStore.
type Store struct {
	pool *pgxpool.Pool
}

// New creates a new PostgreSQL store with the given connection pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Connect creates a new connection pool and returns a Store.
func Connect(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Pool returns the underlying connection pool.
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// Close shuts down the connection pool.
func (s *Store) Close() {
	s.pool.Close()
}

// --- AssetStore implementation ---

// Create inserts a new asset into PostgreSQL.
func (s *Store) Create(ctx context.Context, asset *model.Asset) error {
	query := `
		INSERT INTO assets (id, external_id, source, asset_type, name, fqdn, ip_addresses, attributes, first_seen, last_seen, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := s.pool.Exec(ctx, query,
		asset.ID, asset.ExternalID, asset.Source, asset.AssetType, asset.Name,
		asset.FQDN, asset.IPAddresses, asset.Attributes,
		asset.FirstSeen, asset.LastSeen, asset.Status,
		asset.CreatedAt, asset.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting asset: %w", err)
	}
	return nil
}

// GetByID retrieves an asset by its UUID.
func (s *Store) GetByID(ctx context.Context, id uuid.UUID) (*model.Asset, error) {
	query := `
		SELECT id, external_id, source, asset_type, name, fqdn, ip_addresses, attributes, first_seen, last_seen, status, created_at, updated_at
		FROM assets WHERE id = $1
	`
	return s.scanAsset(s.pool.QueryRow(ctx, query, id))
}

// GetByExternalID retrieves an asset by source and external ID.
func (s *Store) GetByExternalID(ctx context.Context, source, externalID string) (*model.Asset, error) {
	query := `
		SELECT id, external_id, source, asset_type, name, fqdn, ip_addresses, attributes, first_seen, last_seen, status, created_at, updated_at
		FROM assets WHERE source = $1 AND external_id = $2
	`
	return s.scanAsset(s.pool.QueryRow(ctx, query, source, externalID))
}

// List returns assets matching the filter with total count.
func (s *Store) List(ctx context.Context, filter store.AssetFilter) ([]*model.Asset, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.Source != "" {
		where += fmt.Sprintf(" AND source = $%d", argIdx)
		args = append(args, filter.Source)
		argIdx++
	}
	if filter.AssetType != "" {
		where += fmt.Sprintf(" AND asset_type = $%d", argIdx)
		args = append(args, filter.AssetType)
		argIdx++
	}
	if filter.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Search != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR fqdn ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	// Count query
	countQuery := "SELECT COUNT(*) FROM assets " + where
	var total int
	if err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting assets: %w", err)
	}

	// Data query
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, external_id, source, asset_type, name, fqdn, ip_addresses, attributes, first_seen, last_seen, status, created_at, updated_at
		FROM assets %s ORDER BY updated_at DESC LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing assets: %w", err)
	}
	defer rows.Close()

	var assets []*model.Asset
	for rows.Next() {
		a, err := s.scanAssetFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		assets = append(assets, a)
	}
	return assets, total, nil
}

// Update persists changes to an existing asset.
func (s *Store) Update(ctx context.Context, asset *model.Asset) error {
	asset.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE assets SET
			external_id = $2, source = $3, asset_type = $4, name = $5, fqdn = $6,
			ip_addresses = $7, attributes = $8, last_seen = $9, status = $10, updated_at = $11
		WHERE id = $1
	`
	_, err := s.pool.Exec(ctx, query,
		asset.ID, asset.ExternalID, asset.Source, asset.AssetType, asset.Name,
		asset.FQDN, asset.IPAddresses, asset.Attributes,
		asset.LastSeen, asset.Status, asset.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating asset: %w", err)
	}
	return nil
}

// Delete removes an asset by ID.
func (s *Store) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM assets WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("deleting asset: %w", err)
	}
	return nil
}

// FindByIP looks up assets containing the given IP address.
func (s *Store) FindByIP(ctx context.Context, ip string) ([]*model.Asset, error) {
	query := `
		SELECT id, external_id, source, asset_type, name, fqdn, ip_addresses, attributes, first_seen, last_seen, status, created_at, updated_at
		FROM assets WHERE $1 = ANY(ip_addresses)
	`
	rows, err := s.pool.Query(ctx, query, ip)
	if err != nil {
		return nil, fmt.Errorf("finding assets by IP: %w", err)
	}
	defer rows.Close()

	var assets []*model.Asset
	for rows.Next() {
		a, err := s.scanAssetFromRows(rows)
		if err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, nil
}

// FindByFQDN looks up assets matching the given FQDN.
func (s *Store) FindByFQDN(ctx context.Context, fqdn string) ([]*model.Asset, error) {
	query := `
		SELECT id, external_id, source, asset_type, name, fqdn, ip_addresses, attributes, first_seen, last_seen, status, created_at, updated_at
		FROM assets WHERE fqdn = $1
	`
	rows, err := s.pool.Query(ctx, query, fqdn)
	if err != nil {
		return nil, fmt.Errorf("finding assets by FQDN: %w", err)
	}
	defer rows.Close()

	var assets []*model.Asset
	for rows.Next() {
		a, err := s.scanAssetFromRows(rows)
		if err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, nil
}

// --- ChangeEventStore implementation ---

// CreateChangeEvent inserts a new change event.
func (s *Store) CreateChangeEvent(ctx context.Context, event *model.ChangeEvent) error {
	query := `
		INSERT INTO change_events (id, asset_id, action, source, diff, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := s.pool.Exec(ctx, query,
		event.ID, event.AssetID, event.Action, event.Source, event.Diff, event.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("inserting change event: %w", err)
	}
	return nil
}

// ListChangeEventsByAssetID returns change events for a given asset.
func (s *Store) ListChangeEventsByAssetID(ctx context.Context, assetID uuid.UUID, limit, offset int) ([]*model.ChangeEvent, int, error) {
	var total int
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM change_events WHERE asset_id = $1", assetID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting change events: %w", err)
	}

	query := `
		SELECT id, asset_id, action, source, diff, timestamp
		FROM change_events WHERE asset_id = $1
		ORDER BY timestamp DESC LIMIT $2 OFFSET $3
	`
	rows, err := s.pool.Query(ctx, query, assetID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing change events: %w", err)
	}
	defer rows.Close()

	var events []*model.ChangeEvent
	for rows.Next() {
		e := &model.ChangeEvent{}
		if err := rows.Scan(&e.ID, &e.AssetID, &e.Action, &e.Source, &e.Diff, &e.Timestamp); err != nil {
			return nil, 0, fmt.Errorf("scanning change event: %w", err)
		}
		events = append(events, e)
	}
	return events, total, nil
}

// ListRecentChangeEvents returns the most recent change events across all assets.
func (s *Store) ListRecentChangeEvents(ctx context.Context, limit, offset int) ([]*model.ChangeEvent, int, error) {
	var total int
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM change_events").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting change events: %w", err)
	}

	query := `
		SELECT id, asset_id, action, source, diff, timestamp
		FROM change_events ORDER BY timestamp DESC LIMIT $1 OFFSET $2
	`
	rows, err := s.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing recent changes: %w", err)
	}
	defer rows.Close()

	var events []*model.ChangeEvent
	for rows.Next() {
		e := &model.ChangeEvent{}
		if err := rows.Scan(&e.ID, &e.AssetID, &e.Action, &e.Source, &e.Diff, &e.Timestamp); err != nil {
			return nil, 0, fmt.Errorf("scanning change event: %w", err)
		}
		events = append(events, e)
	}
	return events, total, nil
}

// --- helpers ---

func (s *Store) scanAsset(row pgx.Row) (*model.Asset, error) {
	a := &model.Asset{}
	var attrs []byte
	if err := row.Scan(
		&a.ID, &a.ExternalID, &a.Source, &a.AssetType, &a.Name, &a.FQDN,
		&a.IPAddresses, &attrs,
		&a.FirstSeen, &a.LastSeen, &a.Status,
		&a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning asset: %w", err)
	}
	a.Attributes = json.RawMessage(attrs)
	return a, nil
}

func (s *Store) scanAssetFromRows(rows pgx.Rows) (*model.Asset, error) {
	a := &model.Asset{}
	var attrs []byte
	if err := rows.Scan(
		&a.ID, &a.ExternalID, &a.Source, &a.AssetType, &a.Name, &a.FQDN,
		&a.IPAddresses, &attrs,
		&a.FirstSeen, &a.LastSeen, &a.Status,
		&a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scanning asset row: %w", err)
	}
	a.Attributes = json.RawMessage(attrs)
	return a, nil
}
