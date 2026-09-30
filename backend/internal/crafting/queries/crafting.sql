-- name: InsertCraftJob :one
INSERT INTO craft_jobs (id, user_id, recipe_id, quantity, output_item_id, output_quantity, fee_micro, inputs, started_at, ends_at)
VALUES (@id, @user_id, @recipe_id, @quantity, @output_item_id, @output_quantity, @fee_micro, @inputs, @started_at, @ends_at)
ON CONFLICT (user_id) WHERE status = 'CRAFTING' DO NOTHING
RETURNING *;

-- name: ListCraftJobsForUser :many
SELECT * FROM craft_jobs
WHERE user_id = @user_id
  AND (sqlc.narg(before_id)::uuid IS NULL OR id < sqlc.narg(before_id)::uuid)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text)
ORDER BY id DESC
LIMIT @row_limit;

-- name: ListDueCraftJobs :many
SELECT * FROM craft_jobs
WHERE status = 'CRAFTING' AND ends_at <= @now
ORDER BY ends_at
LIMIT @row_limit;

-- name: DeliverCraftJob :one
UPDATE craft_jobs
SET status = 'DELIVERED', delivered_at = @delivered_at
WHERE id = @id AND status = 'CRAFTING'
RETURNING *;
