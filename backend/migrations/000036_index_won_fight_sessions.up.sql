CREATE INDEX fight_sessions_user_wins_index ON fight_sessions (user_id, level_id, finished_at DESC, id DESC) WHERE outcome = 'WON';
