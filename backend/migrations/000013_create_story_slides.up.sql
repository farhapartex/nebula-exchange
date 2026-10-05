CREATE TYPE slide_kind AS ENUM ('SLIDE', 'CALL_TO_ACTION');

CREATE TABLE story_slides (
    id UUID PRIMARY KEY,
    level_id TEXT NOT NULL REFERENCES levels (id) ON DELETE CASCADE,
    position SMALLINT NOT NULL,
    kind slide_kind NOT NULL,
    eyebrow TEXT,
    heading TEXT NOT NULL,
    body TEXT NOT NULL,
    image_url TEXT,
    palette JSONB NOT NULL,
    button_label TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT story_slides_level_position_unique UNIQUE (level_id, position),
    CONSTRAINT story_slides_position_positive CHECK (position > 0),
    CONSTRAINT story_slides_button_only_on_call_to_action CHECK ((kind = 'CALL_TO_ACTION') = (button_label IS NOT NULL)),
    CONSTRAINT story_slides_palette_is_object CHECK (jsonb_typeof(palette) = 'object')
);

CREATE UNIQUE INDEX story_slides_one_call_to_action_per_level ON story_slides (level_id) WHERE kind = 'CALL_TO_ACTION';
