-- name: GetCreatures :many
SELECT id, name
FROM creatures
ORDER BY name;

-- name: CreateCreature :one
INSERT INTO creatures (name)
VALUES ($1)
RETURNING id, name;