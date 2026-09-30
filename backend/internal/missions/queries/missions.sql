-- name: LockPlayerMissions :exec
SELECT pg_advisory_xact_lock(hashtextextended('nebula-missions:' || sqlc.arg(user_id)::text, 0));

-- name: CountActiveMissions :one
SELECT count(*)::integer FROM missions WHERE user_id = @user_id AND status IN ('RUNNING', 'COMPLETED');

-- name: InsertMission :one
INSERT INTO missions (id, user_id, zone_id, ship_item_id, drill_item_id, fuel_spent, ship_hold_id, drill_hold_id, zone_snapshot, started_at, ends_at)
VALUES (@id, @user_id, @zone_id, @ship_item_id, @drill_item_id, @fuel_spent, @ship_hold_id, @drill_hold_id, @zone_snapshot, @started_at, @ends_at)
RETURNING *;

-- name: GetMissionForUser :one
SELECT * FROM missions WHERE id = @id AND user_id = @user_id;

-- name: ListMissionsForUser :many
SELECT * FROM missions
WHERE user_id = @user_id
  AND (sqlc.narg(before_id)::uuid IS NULL OR id < sqlc.narg(before_id)::uuid)
  AND (cardinality(@statuses::text[]) = 0 OR status = ANY(@statuses::text[]))
ORDER BY id DESC
LIMIT @row_limit;

-- name: ListDueMissions :many
SELECT * FROM missions
WHERE status = 'RUNNING' AND ends_at <= @now
ORDER BY ends_at
LIMIT @row_limit;

-- name: CompleteMission :one
UPDATE missions
SET status = 'COMPLETED', loot = @loot, resolved_at = @resolved_at
WHERE id = @id AND status = 'RUNNING'
RETURNING id;

-- name: CollectMission :one
UPDATE missions
SET status = 'COLLECTED', collected_at = @collected_at
WHERE id = @id AND user_id = @user_id AND status = 'COMPLETED'
RETURNING *;

-- name: AbortMission :one
UPDATE missions
SET status = 'ABORTED', aborted_at = @aborted_at
WHERE id = @id AND user_id = @user_id AND status = 'RUNNING' AND ends_at > @aborted_at
RETURNING *;
