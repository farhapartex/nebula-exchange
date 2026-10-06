ALTER TABLE fight_sessions DROP CONSTRAINT IF EXISTS fight_sessions_reason_only_when_rejected;
ALTER TABLE fight_sessions DROP COLUMN IF EXISTS rejection_reason;
