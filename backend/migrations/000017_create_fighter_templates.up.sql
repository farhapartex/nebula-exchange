CREATE TABLE fighter_templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    title TEXT NOT NULL,
    starting_level INTEGER NOT NULL DEFAULT 1,
    stats JSONB NOT NULL,
    look JSONB NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fighter_templates_starting_level_positive CHECK (starting_level > 0),
    CONSTRAINT fighter_templates_settings_are_objects CHECK (jsonb_typeof(stats) = 'object' AND jsonb_typeof(look) = 'object')
);

CREATE UNIQUE INDEX fighter_templates_one_default ON fighter_templates (is_default) WHERE is_default;
