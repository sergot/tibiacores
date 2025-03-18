-- name: GetUserCharacters :many
SELECT id, name, world
FROM characters
WHERE user_id = $1
ORDER BY name;


-- name: GetCharacter :one
SELECT * FROM characters
WHERE id = $1;

-- name: GetCharactersByUserId :many
SELECT * FROM characters
WHERE user_id = $1;