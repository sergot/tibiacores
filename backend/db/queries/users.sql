-- name: CreateAnonymousUser :one
INSERT INTO users (id, is_anonymous)
VALUES ($1, TRUE)
RETURNING *;

-- name: CreateUser :one
INSERT INTO users (email, password, email_verification_token, email_verification_expires_at, is_anonymous)
VALUES ($1, $2, $3, $4, FALSE)
RETURNING *;

-- name: MigrateAnonymousUser :one
UPDATE users
SET email = $1,
    password = $2,
    email_verification_token = $3,
    email_verification_expires_at = $4,
    is_anonymous = false
WHERE id = $5 AND is_anonymous = true
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

-- name: GetUserLists :many
WITH user_lists AS (
    -- Get lists where user is the author
    SELECT l.*, lu.character_id, TRUE as is_author
    FROM lists l
    LEFT JOIN lists_users lu ON l.id = lu.list_id AND lu.user_id = l.author_id
    WHERE l.author_id = $1
    
    UNION ALL
    
    -- Get lists where user is a member
    SELECT l.*, lu.character_id, FALSE as is_author
    FROM lists l
    JOIN lists_users lu ON l.id = lu.list_id
    WHERE lu.user_id = $1 AND l.author_id != $1
)
SELECT DISTINCT
    ul.id,
    ul.author_id,
    ul.name,
    ul.share_code,
    ul.world,
    ul.created_at,
    ul.updated_at,
    ul.is_author,
    c.name as character_name
FROM user_lists ul
LEFT JOIN characters c ON ul.character_id = c.id
ORDER BY ul.created_at DESC;