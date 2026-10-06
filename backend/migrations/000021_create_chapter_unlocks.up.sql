CREATE TYPE chapter_unlock_source AS ENUM ('PURCHASE', 'GRANT');

CREATE TABLE chapter_unlocks (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    chapter_id TEXT NOT NULL REFERENCES chapters (id) ON DELETE RESTRICT,
    source chapter_unlock_source NOT NULL,
    unlocked_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, chapter_id)
);

CREATE INDEX chapter_unlocks_chapter_id_index ON chapter_unlocks (chapter_id);
