CREATE TYPE fight_status AS ENUM ('STARTED', 'FINISHED', 'ABANDONED', 'REJECTED');
CREATE TYPE fight_outcome AS ENUM ('WON', 'LOST', 'DRAW');

CREATE TABLE fight_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    level_id TEXT NOT NULL REFERENCES levels (id) ON DELETE RESTRICT,
    status fight_status NOT NULL DEFAULT 'STARTED',
    seed BIGINT NOT NULL,
    loadout JSONB NOT NULL DEFAULT '{}',
    outcome fight_outcome,
    waves_cleared SMALLINT,
    stars SMALLINT,
    duration_ms INTEGER,
    damage_dealt INTEGER,
    damage_taken INTEGER,
    moves_used JSONB,
    input_log JSONB,
    reward_coins BIGINT,
    reward_experience INTEGER,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fight_sessions_open_until_ended CHECK ((status = 'STARTED') = (finished_at IS NULL)),
    CONSTRAINT fight_sessions_outcome_only_when_finished CHECK (outcome IS NULL OR status = 'FINISHED'),
    CONSTRAINT fight_sessions_stars_range CHECK (stars IS NULL OR stars BETWEEN 0 AND 3)
);

CREATE UNIQUE INDEX fight_sessions_one_open_per_level ON fight_sessions (user_id, level_id) WHERE status = 'STARTED';
CREATE INDEX fight_sessions_user_started_at_index ON fight_sessions (user_id, started_at DESC);
CREATE INDEX fight_sessions_level_id_index ON fight_sessions (level_id);
