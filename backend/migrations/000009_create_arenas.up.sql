CREATE TABLE arenas (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    width INTEGER NOT NULL,
    floor_y INTEGER NOT NULL,
    stage JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT arenas_size_positive CHECK (width > 0 AND floor_y > 0),
    CONSTRAINT arenas_stage_is_object CHECK (jsonb_typeof(stage) = 'object')
);
