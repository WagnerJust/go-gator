-- name: CreateFeedFollow :one

WITH follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    ) RETURNING *
)
SELECT follow.*, feeds.name AS feed_name, users.name AS user_name
FROM follow
JOIN feeds ON follow.feed_id = feeds.id
JOIN users ON follow.user_id = users.id;



-- name: GetFeedFollowsForUser :many

SELECT follow.*, feeds.name AS feed_name, users.name AS user_name
FROM feed_follows follow
JOIN feeds ON follow.feed_id = feeds.id
JOIN users ON follow.user_id = users.id
WHERE users.id = $1;
