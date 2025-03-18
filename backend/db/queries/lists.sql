-- name: CreateList :one
INSERT INTO lists (author_id, name, world)
VALUES ($1, $2, $3)
RETURNING *;

-- name: AddListCharacter :exec
INSERT INTO lists_users (list_id, user_id, character_id)
VALUES ($1, $2, $3);

-- name: GetList :one
SELECT * FROM lists
WHERE id = $1;

-- name: GetMembers :one
SELECT * FROM lists_users
WHERE list_id = $1;

-- name: GetListsByAuthorId :many
SELECT * FROM lists
WHERE author_id = $1;