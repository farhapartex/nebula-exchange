CREATE TYPE tool_status AS ENUM ('OWNED', 'LISTED');
CREATE TYPE tool_source AS ENUM ('SHOP', 'REWARD', 'MARKET');

CREATE TABLE tools (
    id UUID PRIMARY KEY,
    tool_type_id TEXT NOT NULL REFERENCES tool_types (id) ON DELETE RESTRICT,
    owner_user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    mastery_level SMALLINT NOT NULL DEFAULT 1,
    mastery_points INTEGER NOT NULL DEFAULT 0,
    fights_used INTEGER NOT NULL DEFAULT 0,
    hits_landed INTEGER NOT NULL DEFAULT 0,
    status tool_status NOT NULL DEFAULT 'OWNED',
    acquired_from tool_source NOT NULL,
    acquired_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT tools_progress_not_negative CHECK (
        mastery_level >= 1 AND mastery_points >= 0 AND fights_used >= 0 AND hits_landed >= 0
    )
);

CREATE INDEX tools_owner_user_id_index ON tools (owner_user_id, acquired_at DESC, id DESC);
CREATE INDEX tools_tool_type_id_index ON tools (tool_type_id);
