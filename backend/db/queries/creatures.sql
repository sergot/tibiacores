-- name: GetCreatures :many
SELECT id, name
FROM creatures
ORDER BY name;