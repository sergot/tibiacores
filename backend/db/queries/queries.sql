-- TODO: divide into smaller chunks

-- name: GetCreatures :many
SELECT id, name
FROM creatures
ORDER BY name;

-- name: CreateCreature :one
INSERT INTO creatures (name)
VALUES ($1)
RETURNING id, name;

-- name: CreateAnonymousUser :one
INSERT INTO users (session_token)
VALUES ($1)
RETURNING *;

-- name: CreateUser :one
INSERT INTO users (email, password, email_verification_token, email_verification_expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;


-- name: MigrateAnonymousUser :one
UPDATE users
SET email = $1,
    password = $2,
    email_verification_token = $3,
    email_verification_expires_at = $4,
    is_anonymous = false
WHERE id = $5
RETURNING *;

-- name: VerifyEmail :exec
UPDATE users
SET email_verified = true,
    email_verification_token = NULL,
    email_verification_expires_at = NULL
WHERE id = $1 AND email_verification_token = $2;


-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;


-- name: GetUserCharacters :many
SELECT id, name, world
FROM characters
WHERE user_id = $1
ORDER BY name;

-- name: CreateList :one
INSERT INTO lists (author_id, name, world)
VALUES ($1, $2, $3)
RETURNING *;

-- name: AddListCharacter :exec
INSERT INTO lists_users (list_id, user_id, character_id)
VALUES ($1, $2, $3);


-- name: GetUserLists :many
SELECT DISTINCT
    l.*,
    CASE WHEN l.author_id = $1 THEN TRUE ELSE FALSE END AS is_author
FROM lists l
LEFT JOIN lists_users lu ON l.id = lu.list_id AND lu.user_id = $1
WHERE l.author_id = $1 OR lu.user_id = $1
ORDER BY l.created_at DESC;

-- name: GetList :one
SELECT * FROM lists
WHERE id = $1;

-- name: GetMembers :one
SELECT * FROM lists_users
WHERE list_id = $1;

-- name: GetCharacter :one
SELECT * FROM characters
WHERE id = $1;

-- name: GetCharactersByUserId :many
SELECT * FROM characters
WHERE user_id = $1;