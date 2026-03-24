-- +goose Up
-- +goose StatementBegin

-- TibiaWiki canonical name is "Horse (Taupe)", not "Horse (Dark Brown)"
UPDATE creatures SET name = 'Horse (Taupe)' WHERE name = 'Horse (Dark Brown)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

UPDATE creatures SET name = 'Horse (Dark Brown)' WHERE name = 'Horse (Taupe)';

-- +goose StatementEnd
