CREATE TABLE enemies (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    title TEXT NOT NULL,
    stats JSONB NOT NULL,
    brain JSONB NOT NULL,
    look JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT enemies_settings_are_objects CHECK (
        jsonb_typeof(stats) = 'object' AND jsonb_typeof(brain) = 'object' AND jsonb_typeof(look) = 'object'
    )
);
