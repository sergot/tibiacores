-- +goose Up
-- +goose StatementBegin

-- "Monk (Creature)" was added as a new entry but "Monk" is the same creature.
-- Delete the duplicate "Monk (Creature)" and rename "Monk" → "Monk (Creature)"
-- so all existing soul core data is preserved.
DELETE FROM lists_soulcores
WHERE creature_id = (SELECT id FROM creatures WHERE name = 'Monk (Creature)');

DELETE FROM characters_soulcores
WHERE creature_id = (SELECT id FROM creatures WHERE name = 'Monk (Creature)');

DELETE FROM creatures WHERE name = 'Monk (Creature)';

UPDATE creatures SET name = 'Monk (Creature)' WHERE name = 'Monk';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

UPDATE creatures SET name = 'Monk' WHERE name = 'Monk (Creature)';
INSERT INTO creatures (name) VALUES ('Monk (Creature)') ON CONFLICT (name) DO NOTHING;

-- +goose StatementEnd
