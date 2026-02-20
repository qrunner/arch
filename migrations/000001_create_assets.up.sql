-- Create the assets table for storing discovered IT assets.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE asset_status AS ENUM ('active', 'stale', 'removed');

CREATE TABLE IF NOT EXISTS assets (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id   TEXT NOT NULL,
    source        TEXT NOT NULL,
    asset_type    TEXT NOT NULL,
    name          TEXT NOT NULL,
    fqdn          TEXT,
    ip_addresses  TEXT[] DEFAULT '{}',
    attributes    JSONB DEFAULT '{}',
    first_seen    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status        asset_status NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique constraint on source + external_id for reconciliation.
CREATE UNIQUE INDEX idx_assets_source_external_id ON assets (source, external_id);

-- Indexes for common query patterns.
CREATE INDEX idx_assets_source ON assets (source);
CREATE INDEX idx_assets_asset_type ON assets (asset_type);
CREATE INDEX idx_assets_status ON assets (status);
CREATE INDEX idx_assets_ip_addresses ON assets USING GIN (ip_addresses);
CREATE INDEX idx_assets_fqdn ON assets (fqdn) WHERE fqdn IS NOT NULL;
CREATE INDEX idx_assets_name_trgm ON assets USING GIN (name gin_trgm_ops);
CREATE INDEX idx_assets_updated_at ON assets (updated_at DESC);
