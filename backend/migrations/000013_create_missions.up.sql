CREATE TABLE missions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    zone_id TEXT NOT NULL REFERENCES zones (id) ON DELETE RESTRICT,
    ship_item_id INTEGER NOT NULL REFERENCES items (id) ON DELETE RESTRICT,
    drill_item_id INTEGER NOT NULL REFERENCES items (id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'RUNNING' CHECK (status IN ('RUNNING', 'COMPLETED', 'COLLECTED', 'ABORTED')),
    fuel_spent INTEGER NOT NULL CHECK (fuel_spent >= 0),
    ship_hold_id UUID NOT NULL REFERENCES ledger_holds (id) ON DELETE RESTRICT,
    drill_hold_id UUID NOT NULL REFERENCES ledger_holds (id) ON DELETE RESTRICT,
    zone_snapshot JSONB NOT NULL,
    loot JSONB,
    started_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    resolved_at TIMESTAMPTZ,
    collected_at TIMESTAMPTZ,
    aborted_at TIMESTAMPTZ,
    CONSTRAINT missions_loot_after_resolution CHECK ((loot IS NOT NULL) = (resolved_at IS NOT NULL)),
    CONSTRAINT missions_ends_after_start CHECK (ends_at > started_at)
);

CREATE INDEX missions_user_idx ON missions (user_id, id DESC);
CREATE INDEX missions_due_idx ON missions (ends_at) WHERE status = 'RUNNING';
CREATE INDEX missions_active_user_idx ON missions (user_id) WHERE status IN ('RUNNING', 'COMPLETED');
