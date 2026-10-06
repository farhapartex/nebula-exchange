CREATE TABLE tool_mastery_events (
    id UUID PRIMARY KEY,
    tool_id UUID NOT NULL REFERENCES tools (id) ON DELETE CASCADE,
    fight_session_id UUID NOT NULL REFERENCES fight_sessions (id) ON DELETE RESTRICT,
    points_gained INTEGER NOT NULL CHECK (points_gained >= 0),
    level_after SMALLINT NOT NULL CHECK (level_after >= 1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT tool_mastery_events_once_per_fight UNIQUE (tool_id, fight_session_id)
);

CREATE INDEX tool_mastery_events_fight_session_id_index ON tool_mastery_events (fight_session_id);
