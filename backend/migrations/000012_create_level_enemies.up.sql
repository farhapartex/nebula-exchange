CREATE TABLE level_enemies (
    level_id TEXT NOT NULL REFERENCES levels (id) ON DELETE CASCADE,
    wave SMALLINT NOT NULL,
    enemy_id TEXT NOT NULL REFERENCES enemies (id) ON DELETE RESTRICT,
    modifiers JSONB NOT NULL DEFAULT '{}',
    intro_line TEXT,
    PRIMARY KEY (level_id, wave),
    CONSTRAINT level_enemies_wave_positive CHECK (wave > 0),
    CONSTRAINT level_enemies_modifiers_is_object CHECK (jsonb_typeof(modifiers) = 'object')
);

CREATE INDEX level_enemies_enemy_id_index ON level_enemies (enemy_id);
