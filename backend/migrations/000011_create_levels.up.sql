CREATE TYPE level_kind AS ENUM ('STORY', 'TRAINING', 'BOSS');

CREATE TABLE levels (
    id TEXT PRIMARY KEY,
    chapter_id TEXT REFERENCES chapters (id) ON DELETE RESTRICT,
    number INTEGER,
    kind level_kind NOT NULL,
    title TEXT NOT NULL,
    teaser TEXT NOT NULL,
    arena_id TEXT NOT NULL REFERENCES arenas (id) ON DELETE RESTRICT,
    time_limit_seconds INTEGER NOT NULL,
    difficulty JSONB NOT NULL DEFAULT '{}',
    star_rules JSONB NOT NULL DEFAULT '[]',
    first_clear_coins BIGINT NOT NULL DEFAULT 0,
    first_clear_experience INTEGER NOT NULL DEFAULT 0,
    replay_experience INTEGER NOT NULL DEFAULT 0,
    is_published BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT levels_chapter_number_unique UNIQUE (chapter_id, number),
    CONSTRAINT levels_training_has_no_chapter CHECK ((kind = 'TRAINING') = (chapter_id IS NULL)),
    CONSTRAINT levels_number_matches_chapter CHECK ((chapter_id IS NULL) = (number IS NULL)),
    CONSTRAINT levels_number_positive CHECK (number IS NULL OR number > 0),
    CONSTRAINT levels_time_limit_positive CHECK (time_limit_seconds > 0),
    CONSTRAINT levels_rewards_not_negative CHECK (first_clear_coins >= 0 AND first_clear_experience >= 0 AND replay_experience >= 0),
    CONSTRAINT levels_difficulty_is_object CHECK (jsonb_typeof(difficulty) = 'object'),
    CONSTRAINT levels_star_rules_is_array CHECK (jsonb_typeof(star_rules) = 'array')
);

CREATE INDEX levels_arena_id_index ON levels (arena_id);
