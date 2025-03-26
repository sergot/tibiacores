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

-- name: GetMembers :many
SELECT * FROM lists_users
WHERE list_id = $1;

-- name: GetListsByAuthorId :many
SELECT * FROM lists
WHERE author_id = $1;

-- name: GetListByShareCode :one
SELECT * FROM lists
WHERE share_code = $1;

-- name: IsUserListMember :one
SELECT EXISTS (
  SELECT 1
  FROM lists_users
  WHERE list_id = $1 AND user_id = $2 AND active = true
) as is_member;

-- name: GetListMembers :many
SELECT 
  u.id as user_id,
  c.name as character_name,
  COUNT(DISTINCT CASE WHEN ls.status = 'obtained' OR ls.status = 'unlocked' THEN ls.creature_id END) as obtained_count,
  COUNT(DISTINCT CASE WHEN ls.status = 'unlocked' THEN ls.creature_id END) as unlocked_count
FROM lists_users lu
JOIN users u ON lu.user_id = u.id
JOIN characters c ON lu.character_id = c.id
LEFT JOIN lists_soulcores ls ON ls.list_id = $1 AND ls.added_by_user_id = u.id
WHERE lu.list_id = $1 AND lu.active = true
GROUP BY u.id, c.name;

-- name: GetListSoulcores :many
SELECT 
  ls.list_id,
  ls.creature_id,
  ls.status,
  cr.name as creature_name,
  c.name as added_by,
  ls.added_by_user_id
FROM lists_soulcores ls
JOIN creatures cr ON ls.creature_id = cr.id
LEFT JOIN lists_users lu ON ls.list_id = lu.list_id AND ls.added_by_user_id = lu.user_id
LEFT JOIN characters c ON lu.character_id = c.id
WHERE ls.list_id = $1
ORDER BY cr.name;

-- name: GetListSoulcore :one
SELECT 
  ls.list_id,
  ls.creature_id,
  ls.status,
  cr.name as creature_name,
  c.name as added_by,
  ls.added_by_user_id
FROM lists_soulcores ls
JOIN creatures cr ON ls.creature_id = cr.id
LEFT JOIN lists_users lu ON ls.list_id = lu.list_id AND ls.added_by_user_id = lu.user_id
LEFT JOIN characters c ON lu.character_id = c.id
WHERE ls.list_id = $1 AND ls.creature_id = $2;

-- name: AddSoulcoreToList :exec
INSERT INTO lists_soulcores (list_id, creature_id, status, added_by_user_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (list_id, creature_id) DO UPDATE
SET status = EXCLUDED.status, added_by_user_id = EXCLUDED.added_by_user_id;

-- name: UpdateSoulcoreStatus :exec
UPDATE lists_soulcores
SET status = $3
WHERE list_id = $1 AND creature_id = $2;

-- name: RemoveListSoulcore :exec
DELETE FROM lists_soulcores
WHERE list_id = $1 AND creature_id = $2;

-- name: DeactivateCharacterListMemberships :exec
UPDATE lists_users
SET active = false
WHERE character_id = $1;
