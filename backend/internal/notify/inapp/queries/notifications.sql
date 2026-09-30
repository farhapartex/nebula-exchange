-- name: InsertNotification :one
INSERT INTO notifications (id, user_id, kind, title, body, link)
VALUES (@id, @user_id, @kind, @title, @body, @link)
RETURNING *;

-- name: ListNotifications :many
SELECT * FROM notifications
WHERE user_id = @user_id
  AND (sqlc.narg(before_id)::uuid IS NULL OR id < sqlc.narg(before_id)::uuid)
  AND (NOT sqlc.arg(unread_only)::boolean OR read_at IS NULL)
ORDER BY id DESC
LIMIT @row_limit;

-- name: CountUnreadNotifications :one
SELECT count(*)::integer FROM notifications WHERE user_id = @user_id AND read_at IS NULL;

-- name: MarkNotificationsRead :one
WITH marked AS (
    UPDATE notifications
    SET read_at = @read_at
    WHERE user_id = @user_id
      AND read_at IS NULL
      AND (sqlc.arg(mark_all)::boolean OR id = ANY(@notification_ids::uuid[]))
    RETURNING id
)
SELECT count(*)::integer FROM marked;
