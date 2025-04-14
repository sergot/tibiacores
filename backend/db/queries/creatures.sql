-- name: GetCreatures :many
SELECT id, name, difficulty
FROM creatures
ORDER BY name;