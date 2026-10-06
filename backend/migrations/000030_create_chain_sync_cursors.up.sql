CREATE TABLE chain_sync_cursors (
    name TEXT PRIMARY KEY,
    chain_id BIGINT NOT NULL,
    block_number BIGINT NOT NULL,
    block_hash CHAR(66) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
