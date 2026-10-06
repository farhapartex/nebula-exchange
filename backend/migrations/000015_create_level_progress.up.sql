CREATE TYPE progress_status AS ENUM ('STARTED', 'COMPLETED');

CREATE TABLE level_progress (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    level_id TEXT NOT NULL REFERENCES levels (id) ON DELETE RESTRICT,
    status progress_status NOT NULL DEFAULT 'STARTED',
    best_stars SMALLINT,
    attempts INTEGER NOT NULL DEFAULT 0,
    first_started_at TIMESTAMPTZ NOT NULL,
    first_completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, level_id),
    CONSTRAINT level_progress_attempts_not_negative CHECK (attempts >= 0),
    CONSTRAINT level_progress_best_stars_range CHECK (best_stars IS NULL OR best_stars BETWEEN 0 AND 3),
    CONSTRAINT level_progress_completion_consistency CHECK ((status = 'COMPLETED') = (first_completed_at IS NOT NULL))
);

CREATE INDEX level_progress_level_id_index ON level_progress (level_id);
