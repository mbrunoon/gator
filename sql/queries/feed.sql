-- name: CreateFeed :one
INSERT INTO feeds(id, name, url, user_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetFeed :one
SELECT * FROM feeds WHERE id = $1;

-- name: GetFeedByUser :many
SELECT * FROM feeds WHERE user_id = $1;

-- name: GetFeedByURL :one
SELECT * FROM feeds WHERE url = $1;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: CreateFeedFollow :one
WITH inserted_feed_follow as (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
) SELECT
    inserted_feed_follow.*,
    feeds.name as feed_name,
    users.name as user_name
FROM inserted_feed_follow
INNER JOIN users ON users.id = user_id
INNER JOIN feeds ON feeds.id = feed_id;

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.*, feeds.name AS feed_name, users.name AS user_name
FROM feed_follows
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
INNER JOIN users ON feed_follows.user_id = users.id
WHERE feed_follows.user_id = $1;

-- name: GetFeedFollowByUser :one
SELECT * 
FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows WHERE id = $1;