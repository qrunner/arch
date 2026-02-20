-- Create the alert_rules table for notification configuration.
CREATE TABLE IF NOT EXISTS alert_rules (
    id        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name      TEXT NOT NULL,
    actions   TEXT[] NOT NULL DEFAULT '{}',
    sources   TEXT[] DEFAULT '{}',
    channels  TEXT[] NOT NULL DEFAULT '{}',
    enabled   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
