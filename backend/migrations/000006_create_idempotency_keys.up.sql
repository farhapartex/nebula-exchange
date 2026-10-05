CREATE TYPE idempotency_key_status AS ENUM ('IN_PROGRESS', 'COMPLETED');

CREATE TABLE idempotency_keys (
    scope TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    request_hash BYTEA NOT NULL,
    status idempotency_key_status NOT NULL,
    response_status_code INTEGER,
    response_content_type TEXT NOT NULL DEFAULT '',
    response_body BYTEA,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    PRIMARY KEY (scope, idempotency_key),
    CONSTRAINT idempotency_keys_completed_has_response CHECK (status = 'IN_PROGRESS' OR response_status_code IS NOT NULL)
);

CREATE INDEX idempotency_keys_created_at_index ON idempotency_keys (created_at);
