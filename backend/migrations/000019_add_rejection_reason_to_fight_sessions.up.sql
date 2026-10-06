ALTER TABLE fight_sessions ADD COLUMN rejection_reason TEXT;
ALTER TABLE fight_sessions ADD CONSTRAINT fight_sessions_reason_only_when_rejected CHECK ((status = 'REJECTED') = (rejection_reason IS NOT NULL));
