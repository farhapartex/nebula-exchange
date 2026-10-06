CREATE TYPE tool_category AS ENUM ('WEAPON', 'GUARD');
CREATE TYPE tool_rarity AS ENUM ('COMMON', 'UNCOMMON', 'RARE', 'EPIC', 'LEGENDARY');

CREATE TABLE tool_types (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    category tool_category NOT NULL,
    rarity tool_rarity NOT NULL,
    base_stats JSONB NOT NULL,
    mastery_curve JSONB NOT NULL,
    max_mastery_level SMALLINT NOT NULL,
    shop_price_coins BIGINT,
    is_tradeable BOOLEAN NOT NULL DEFAULT true,
    max_supply INTEGER,
    image_key TEXT,
    introduced_in_level_id TEXT REFERENCES levels (id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT tool_types_settings_are_objects CHECK (
        jsonb_typeof(base_stats) = 'object' AND jsonb_typeof(mastery_curve) = 'object'
    ),
    CONSTRAINT tool_types_max_mastery_level_range CHECK (max_mastery_level BETWEEN 1 AND 10),
    CONSTRAINT tool_types_shop_price_positive CHECK (shop_price_coins IS NULL OR shop_price_coins > 0),
    CONSTRAINT tool_types_max_supply_positive CHECK (max_supply IS NULL OR max_supply > 0)
);

CREATE INDEX tool_types_introduced_in_level_id_index ON tool_types (introduced_in_level_id);
