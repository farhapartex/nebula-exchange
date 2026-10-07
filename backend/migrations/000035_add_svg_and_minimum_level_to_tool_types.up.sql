ALTER TABLE tool_types
    DROP COLUMN image_key,
    ADD COLUMN image_svg TEXT,
    ADD COLUMN minimum_fighter_level SMALLINT NOT NULL DEFAULT 1,
    ADD CONSTRAINT tool_types_minimum_fighter_level_positive CHECK (minimum_fighter_level >= 1),
    ADD CONSTRAINT tool_types_image_svg_size CHECK (image_svg IS NULL OR octet_length(image_svg) <= 32768);

CREATE INDEX tool_types_shop_order_index ON tool_types (minimum_fighter_level, id) WHERE shop_price_coins IS NOT NULL;
