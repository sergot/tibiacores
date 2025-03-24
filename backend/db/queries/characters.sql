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

-- name: AddCharacterSoulcore :exec
INSERT INTO characters_soulcores (character_id, creature_id)
VALUES ($1, $2);

-- name: RemoveCharacterSoulcore :exec
DELETE FROM characters_soulcores
WHERE character_id = $1 AND creature_id = $2;

-- name: CreateCharacterClaim :one
INSERT INTO character_claims (character_id, claimer_id, verification_code, status)
VALUES ($1, $2, $3, 'pending')
RETURNING *;

-- name: GetCharacterClaim :one
SELECT * FROM character_claims
WHERE character_id = $1 AND claimer_id = $2;

-- name: GetPendingClaims :many
SELECT c.*, ch.name as character_name
FROM character_claims c
JOIN characters ch ON c.character_id = ch.id
WHERE c.status = 'pending' AND c.last_checked_at < NOW() - INTERVAL '1 hour';

-- name: UpdateClaimStatus :one
UPDATE character_claims
SET status = $3,
    last_checked_at = NOW(),
    updated_at = NOW()
WHERE character_id = $1 AND claimer_id = $2
RETURNING *;

-- name: UpdateCharacterOwner :one
UPDATE characters
SET user_id = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetCharacterByName :one
SELECT * FROM characters
WHERE name = $1;

-- name: GetClaimByID :one
SELECT c.id, c.character_id, c.claimer_id, c.verification_code, 
       c.status, c.last_checked_at, c.created_at, c.updated_at,
       ch.name as character_name
FROM character_claims c
JOIN characters ch ON c.character_id = ch.id
WHERE c.id = $1;

-- name: GetPendingClaimsToCheck :many
SELECT c.*, ch.name as character_name
FROM character_claims c
JOIN characters ch ON c.character_id = ch.id
WHERE c.status = 'pending' 
  AND c.last_checked_at < NOW() - INTERVAL '1 minute'
  AND c.created_at > NOW() - INTERVAL '24 hours';