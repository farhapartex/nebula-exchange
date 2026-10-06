CREATE TABLE stripe_webhook_events (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    payload JSONB NOT NULL,
    received_at TIMESTAMPTZ NOT NULL,
    processed_at TIMESTAMPTZ
);

CREATE INDEX stripe_webhook_events_type_index ON stripe_webhook_events (type);
