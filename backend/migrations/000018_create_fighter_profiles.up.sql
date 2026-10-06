CREATE TABLE fighter_profiles (
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    template_id TEXT NOT NULL REFERENCES fighter_templates (id) ON DELETE RESTRICT,
    fighter_level INTEGER NOT NULL DEFAULT 1,
    experience INTEGER NOT NULL DEFAULT 0,
    wins INTEGER NOT NULL DEFAULT 0,
    losses INTEGER NOT NULL DEFAULT 0,
    stats JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fighter_profiles_level_positive CHECK (fighter_level > 0),
    CONSTRAINT fighter_profiles_counters_not_negative CHECK (experience >= 0 AND wins >= 0 AND losses >= 0),
    CONSTRAINT fighter_profiles_stats_is_object CHECK (jsonb_typeof(stats) = 'object')
);

CREATE INDEX fighter_profiles_template_id_index ON fighter_profiles (template_id);
