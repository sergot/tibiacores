-- name: CreateSoulcoreSuggestion :exec
INSERT INTO character_soulcore_suggestions (character_id, creature_id, list_id)
VALUES ($1, $2, $3)
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