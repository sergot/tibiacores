-- name: GetUserCharacters :many
SELECT id, name, world
FROM characters
WHERE user_id = $1
ORDER BY name;


-- name: GetCharacter :one
SELECT * FROM characters
WHERE id = $1;

-- name: GetCharactersByUserID :many
SELECT * FROM characters
WHERE user_id = $1;

-- name: CreateCharacter :one
INSERT INTO characters (user_id, name, world)
VALUES ($1, $2, $3)
RETURNING *;