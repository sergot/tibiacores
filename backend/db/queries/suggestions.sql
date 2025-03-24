-- name: CreateSoulcoreSuggestions :exec
WITH list_members AS (
    SELECT DISTINCT c.id as character_id, ls.creature_id, l.id as list_id
    FROM lists_users lu
    JOIN characters c ON c.user_id = lu.user_id
    JOIN lists l ON l.id = lu.list_id
    JOIN lists_soulcores ls ON ls.list_id = l.id
    WHERE l.id = $1 AND ls.creature_id = $2 AND ls.status = 'unlocked'
    AND NOT EXISTS (
        SELECT 1 FROM characters_soulcores cs 
        WHERE cs.character_id = c.id AND cs.creature_id = ls.creature_id
    )
)
INSERT INTO character_soulcore_suggestions (character_id, creature_id, list_id)
SELECT character_id, creature_id, list_id
FROM list_members
ON CONFLICT DO NOTHING;

-- name: GetCharacterSuggestions :many
SELECT cs.character_id, cs.creature_id, cs.list_id, cs.suggested_at, c.name as creature_name
FROM character_soulcore_suggestions cs
JOIN creatures c ON c.id = cs.creature_id
WHERE cs.character_id = $1
ORDER BY cs.suggested_at DESC;

-- name: DeleteSoulcoreSuggestion :exec
DELETE FROM character_soulcore_suggestions
WHERE character_id = $1 AND creature_id = $2;

-- name: GetPendingSuggestionsForUser :many
SELECT 
    c.id as character_id,
    c.name as character_name,
    COUNT(cs.creature_id) as suggestion_count
FROM characters c
JOIN character_soulcore_suggestions cs ON cs.character_id = c.id
WHERE c.user_id = $1
GROUP BY c.id, c.name;