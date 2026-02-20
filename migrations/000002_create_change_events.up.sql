-- Create the change_events table for tracking asset change history.
CREATE TABLE IF NOT EXISTS change_events (
    id        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    asset_id  UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    action    TEXT NOT NULL,
    source    TEXT NOT NULL,
    diff      JSONB,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_change_events_asset_id ON change_events (asset_id);
CREATE INDEX idx_change_events_timestamp ON change_events (timestamp DESC);
CREATE INDEX idx_change_events_action ON change_events (action);
