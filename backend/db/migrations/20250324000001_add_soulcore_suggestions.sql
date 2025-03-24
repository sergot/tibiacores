-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS character_soulcore_suggestions (
    character_id UUID NOT NULL REFERENCES characters(id),
    creature_id UUID NOT NULL REFERENCES creatures(id),
    list_id UUID NOT NULL REFERENCES lists(id),
    suggested_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (character_id, creature_id)
);

CREATE INDEX character_soulcore_suggestions_character_id_idx ON character_soulcore_suggestions(character_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS character_soulcore_suggestions;
-- +goose StatementEnd