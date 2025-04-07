-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS characters_soulcores (
    character_id UUID NOT NULL REFERENCES characters(id),
    creature_id UUID NOT NULL REFERENCES creatures(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (character_id, creature_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS characters_soulcores;
-- +goose StatementEnd
