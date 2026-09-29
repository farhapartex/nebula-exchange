CREATE TABLE items (
    id INTEGER PRIMARY KEY CHECK (id > 0),
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    category TEXT NOT NULL CHECK (category IN ('resource', 'component', 'drill', 'ship', 'consumable', 'legendary')),
    tier INTEGER CHECK (tier IS NULL OR tier > 0),
    rarity_rank INTEGER NOT NULL DEFAULT 0,
    is_tradeable BOOLEAN NOT NULL DEFAULT true,
    is_auction_only BOOLEAN NOT NULL DEFAULT false,
    max_supply BIGINT CHECK (max_supply IS NULL OR max_supply > 0),
    description TEXT NOT NULL DEFAULT '',
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE recipes (
    id TEXT PRIMARY KEY,
    output_item_id INTEGER NOT NULL REFERENCES items (id),
    output_quantity INTEGER NOT NULL DEFAULT 1 CHECK (output_quantity > 0),
    craft_seconds INTEGER NOT NULL CHECK (craft_seconds > 0),
    fee_micro BIGINT NOT NULL CHECK (fee_micro >= 0),
    is_enabled BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE recipe_inputs (
    recipe_id TEXT NOT NULL REFERENCES recipes (id) ON DELETE CASCADE,
    item_id INTEGER NOT NULL REFERENCES items (id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (recipe_id, item_id)
);

CREATE TABLE upgrades (
    id TEXT PRIMARY KEY,
    from_item_id INTEGER NOT NULL REFERENCES items (id),
    to_item_id INTEGER NOT NULL REFERENCES items (id),
    craft_fee_micro BIGINT NOT NULL CHECK (craft_fee_micro >= 0),
    buy_price_micro BIGINT CHECK (buy_price_micro IS NULL OR buy_price_micro > 0),
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    CHECK (from_item_id <> to_item_id)
);

CREATE TABLE upgrade_inputs (
    upgrade_id TEXT NOT NULL REFERENCES upgrades (id) ON DELETE CASCADE,
    item_id INTEGER NOT NULL REFERENCES items (id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (upgrade_id, item_id)
);

CREATE TABLE zones (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    duration_seconds INTEGER NOT NULL CHECK (duration_seconds > 0),
    fuel_cost INTEGER NOT NULL CHECK (fuel_cost >= 0),
    minimum_drill_tier INTEGER NOT NULL DEFAULT 1,
    allowed_ship_item_ids INTEGER[] NOT NULL DEFAULT '{}',
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE loot_tables (
    zone_id TEXT NOT NULL REFERENCES zones (id) ON DELETE CASCADE,
    item_id INTEGER NOT NULL REFERENCES items (id),
    minimum_quantity INTEGER NOT NULL CHECK (minimum_quantity >= 0),
    maximum_quantity INTEGER NOT NULL,
    chance_basis_points INTEGER NOT NULL DEFAULT 10000 CHECK (chance_basis_points BETWEEN 1 AND 10000),
    PRIMARY KEY (zone_id, item_id),
    CHECK (maximum_quantity >= minimum_quantity)
);

CREATE TABLE shop_skus (
    sku TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price_micro BIGINT NOT NULL CHECK (price_micro > 0),
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_enabled BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE shop_sku_items (
    sku TEXT NOT NULL REFERENCES shop_skus (sku) ON DELETE CASCADE,
    item_id INTEGER NOT NULL REFERENCES items (id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (sku, item_id)
);

INSERT INTO items (id, slug, name, category, tier, rarity_rank, is_tradeable, is_auction_only, max_supply, description, attributes) VALUES
    (1, 'iron-ore', 'Iron Ore', 'resource', NULL, 1, true, false, NULL, 'The most common ore in the belt. Smelted into Alloy Plates.', '{}'),
    (2, 'copper-ore', 'Copper Ore', 'resource', NULL, 2, true, false, NULL, 'Conductive ore used for Alloy Plates and Circuits.', '{}'),
    (3, 'crystal', 'Crystal', 'resource', NULL, 3, true, false, NULL, 'Resonant crystal from the Crystal Moon.', '{}'),
    (4, 'plasma', 'Plasma', 'resource', NULL, 4, true, false, NULL, 'Charged gas harvested inside the Plasma Nebula.', '{}'),
    (5, 'rare-earth', 'Rare Earth', 'resource', NULL, 5, true, false, NULL, 'Scarce elements found only in the Deep Void.', '{}'),
    (6, 'void-shard', 'Void Shard', 'resource', NULL, 6, true, false, NULL, 'A fragment of the void itself. Extremely rare.', '{}'),
    (101, 'alloy-plate', 'Alloy Plate', 'component', NULL, 0, true, false, NULL, 'Structural plating for ships and drills.', '{}'),
    (102, 'circuit', 'Circuit', 'component', NULL, 0, true, false, NULL, 'Control electronics for upgrades.', '{}'),
    (103, 'power-core', 'Power Core', 'component', NULL, 0, true, false, NULL, 'Stores enough energy to run a high-tier drill.', '{}'),
    (104, 'void-engine', 'Void Engine', 'component', NULL, 0, true, false, NULL, 'Exotic drive that powers the best drills.', '{}'),
    (201, 'drill-t1', 'Drill T1', 'drill', 1, 0, true, false, NULL, 'A basic mining drill.', '{"drill_multiplier": "1.00"}'),
    (202, 'drill-t2', 'Drill T2', 'drill', 2, 0, true, false, NULL, 'Reinforced drill that can crack Crystal Moon rock.', '{"drill_multiplier": "1.25"}'),
    (203, 'drill-t3', 'Drill T3', 'drill', 3, 0, true, false, NULL, 'Plasma-rated drill for the Nebula.', '{"drill_multiplier": "1.50"}'),
    (204, 'drill-t4', 'Drill T4', 'drill', 4, 0, true, false, NULL, 'Deep-void drill with a shielded bore.', '{"drill_multiplier": "2.00"}'),
    (205, 'drill-t5', 'Drill T5', 'drill', 5, 0, true, false, NULL, 'The finest drill a pilot can build.', '{"drill_multiplier": "2.50"}'),
    (301, 'scout', 'Scout', 'ship', 1, 0, true, false, NULL, 'Light starter ship.', '{"cargo_capacity": 20, "speed_multiplier": "1.00"}'),
    (302, 'hauler', 'Hauler', 'ship', 2, 0, true, false, NULL, 'Triple the cargo of a Scout.', '{"cargo_capacity": 60, "speed_multiplier": "1.00"}'),
    (303, 'freighter', 'Freighter', 'ship', 3, 0, true, false, NULL, 'Huge hold, slow engines.', '{"cargo_capacity": 150, "speed_multiplier": "1.20"}'),
    (304, 'interceptor', 'Interceptor', 'ship', 2, 0, true, false, NULL, 'Fast ship that finishes missions sooner.', '{"cargo_capacity": 40, "speed_multiplier": "0.60"}'),
    (401, 'fuel-cell', 'Fuel Cell', 'consumable', NULL, 0, true, false, NULL, 'Burned to launch missions.', '{}'),
    (10001, 'relic-drill-of-orion', 'Relic Drill of Orion', 'legendary', NULL, 0, true, true, 1, 'A one-of-a-kind drill recovered from a derelict station.', '{"drill_multiplier": "3.00"}');

INSERT INTO recipes (id, output_item_id, output_quantity, craft_seconds, fee_micro) VALUES
    ('alloy-plate', 101, 1, 60, 20000),
    ('circuit', 102, 1, 120, 50000),
    ('power-core', 103, 1, 300, 100000),
    ('void-engine', 104, 1, 1800, 1000000);

INSERT INTO recipe_inputs (recipe_id, item_id, quantity) VALUES
    ('alloy-plate', 1, 5), ('alloy-plate', 2, 2),
    ('circuit', 2, 3), ('circuit', 3, 1),
    ('power-core', 3, 2), ('power-core', 4, 2),
    ('void-engine', 6, 1), ('void-engine', 103, 2), ('void-engine', 5, 1);

INSERT INTO upgrades (id, from_item_id, to_item_id, craft_fee_micro, buy_price_micro) VALUES
    ('drill-t1-to-t2', 201, 202, 500000, 3000000),
    ('drill-t2-to-t3', 202, 203, 1000000, 7000000),
    ('drill-t3-to-t4', 203, 204, 2500000, 15000000),
    ('drill-t4-to-t5', 204, 205, 5000000, NULL),
    ('scout-to-hauler', 301, 302, 1000000, 6000000),
    ('hauler-to-freighter', 302, 303, 3000000, 18000000),
    ('scout-to-interceptor', 301, 304, 2000000, 12000000);

INSERT INTO upgrade_inputs (upgrade_id, item_id, quantity) VALUES
    ('drill-t1-to-t2', 101, 5), ('drill-t1-to-t2', 102, 2),
    ('drill-t2-to-t3', 101, 10), ('drill-t2-to-t3', 102, 5), ('drill-t2-to-t3', 103, 1),
    ('drill-t3-to-t4', 101, 20), ('drill-t3-to-t4', 102, 10), ('drill-t3-to-t4', 103, 5),
    ('drill-t4-to-t5', 104, 2), ('drill-t4-to-t5', 103, 10),
    ('scout-to-hauler', 101, 15), ('scout-to-hauler', 102, 5),
    ('hauler-to-freighter', 101, 40), ('hauler-to-freighter', 103, 10),
    ('scout-to-interceptor', 102, 10), ('scout-to-interceptor', 103, 5);

INSERT INTO zones (id, name, description, duration_seconds, fuel_cost, minimum_drill_tier, allowed_ship_item_ids, sort_order) VALUES
    ('asteroid-belt', 'Asteroid Belt', 'A dense field of iron-rich rock close to home.', 900, 2, 1, '{}', 1),
    ('crystal-moon', 'Crystal Moon', 'A frozen moon laced with resonant crystal.', 1800, 4, 2, '{}', 2),
    ('plasma-nebula', 'Plasma Nebula', 'A glowing storm of charged gas. Needs a bigger hold than a Scout.', 3600, 6, 3, '{302,303,304}', 3),
    ('deep-void', 'Deep Void', 'The edge of known space, where Void Shards drift.', 14400, 15, 4, '{303,304}', 4);

INSERT INTO loot_tables (zone_id, item_id, minimum_quantity, maximum_quantity, chance_basis_points) VALUES
    ('asteroid-belt', 1, 8, 14, 10000), ('asteroid-belt', 2, 3, 6, 10000),
    ('crystal-moon', 3, 4, 8, 10000), ('crystal-moon', 2, 4, 8, 10000),
    ('plasma-nebula', 4, 3, 6, 10000), ('plasma-nebula', 3, 2, 5, 10000),
    ('deep-void', 5, 2, 4, 10000), ('deep-void', 6, 1, 1, 500);

INSERT INTO shop_skus (sku, name, description, price_micro, sort_order) VALUES
    ('fuel-cell', 'Fuel Cell', 'One Fuel Cell for launching missions.', 200000, 1),
    ('scout', 'Scout', 'A fresh Tier 1 ship.', 2000000, 2),
    ('drill-t1', 'Drill T1', 'A basic mining drill.', 1000000, 3),
    ('starter-bundle', 'Starter Bundle', '1 Scout, 1 Drill T1 and 20 Fuel Cells.', 3500000, 4);

INSERT INTO shop_sku_items (sku, item_id, quantity) VALUES
    ('fuel-cell', 401, 1),
    ('scout', 301, 1),
    ('drill-t1', 201, 1),
    ('starter-bundle', 301, 1), ('starter-bundle', 201, 1), ('starter-bundle', 401, 20);
