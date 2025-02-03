-- name: CreateFeed :one
INSERT INTO feeds(id, name, url, user_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetFeed :one
SELECT * FROM feeds WHERE id = $1;

-- name: GetFeedByUser :many
SELECT * FROM feeds WHERE user_id = $1;

-- name: GetFeeds :many
SELECT * FROM feeds;