CREATE TYPE loadout_slot AS ENUM ('WEAPON', 'GUARD');

CREATE TABLE loadout_slots (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    slot loadout_slot NOT NULL,
    tool_id UUID REFERENCES tools (id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, slot),
    CONSTRAINT loadout_slots_tool_id_unique UNIQUE (tool_id)
);
