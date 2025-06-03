-- name: CreateChatMessage :one
INSERT INTO list_chat_messages (list_id, user_id, character_id, message)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetChatMessages :many
SELECT 
    lcm.id,
    lcm.list_id,
    lcm.user_id,
    c.name as character_name,
    lcm.message,
    lcm.created_at
FROM list_chat_messages lcm
JOIN characters c ON lcm.character_id = c.id
WHERE lcm.list_id = $1
ORDER BY lcm.created_at DESC
LIMIT $2
OFFSET $3;

-- name: GetChatMessagesByTimestamp :many
SELECT 
    lcm.id,
    lcm.list_id,
    lcm.user_id,
    c.name as character_name,
    lcm.message,
    lcm.created_at
FROM list_chat_messages lcm
JOIN characters c ON lcm.character_id = c.id
WHERE lcm.list_id = $1 AND lcm.created_at > $2
ORDER BY lcm.created_at ASC;

-- name: DeleteChatMessage :exec
DELETE FROM list_chat_messages
WHERE id = $1 AND user_id = $2;

-- name: DeleteAllChatMessages :exec
DELETE FROM list_chat_messages
WHERE list_id = $1;

-- name: GetChatNotificationsForUser :many
WITH user_lists AS (
    SELECT DISTINCT l.id, l.name
    FROM lists l
    JOIN lists_users lu ON l.id = lu.list_id
    WHERE lu.user_id = $1
),
last_read_times AS (
    SELECT user_id, list_id, last_read_at
    FROM list_user_read_status
    WHERE user_id = $1
),
list_messages AS (
    SELECT 
        lcm.list_id, 
        ul.name as list_name,
        MAX(lcm.created_at) as last_message_time,
        COUNT(*) as unread_count,
        (
            SELECT c.name 
            FROM list_chat_messages lcm2
            JOIN characters c ON lcm2.character_id = c.id
            WHERE lcm2.list_id = lcm.list_id
            ORDER BY lcm2.created_at DESC
            LIMIT 1
        ) as last_character_name
    FROM list_chat_messages lcm
    JOIN user_lists ul ON lcm.list_id = ul.id
    LEFT JOIN last_read_times lrt ON lcm.list_id = lrt.list_id
    WHERE (lrt.last_read_at IS NULL OR lcm.created_at > lrt.last_read_at)
      AND lcm.user_id != $1 -- Don't count user's own messages as unread
    GROUP BY lcm.list_id, ul.name
    HAVING COUNT(*) > 0
)
SELECT 
    list_id,
    list_name,
    last_message_time,
    unread_count,
    last_character_name
FROM list_messages
ORDER BY last_message_time DESC;

-- name: MarkListMessagesAsRead :exec
INSERT INTO list_user_read_status (user_id, list_id, last_read_at)
VALUES ($1, $2, NOW())
ON CONFLICT (user_id, list_id) 
DO UPDATE SET last_read_at = NOW();
