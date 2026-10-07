DROP INDEX IF EXISTS tool_types_shop_order_index;

ALTER TABLE tool_types
    DROP CONSTRAINT IF EXISTS tool_types_image_svg_size,
    DROP CONSTRAINT IF EXISTS tool_types_minimum_fighter_level_positive,
    DROP COLUMN IF EXISTS minimum_fighter_level,
    DROP COLUMN IF EXISTS image_svg,
    ADD COLUMN image_key TEXT;
