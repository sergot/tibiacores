-- +goose Up
-- +goose StatementBegin
CREATE TYPE soulcore_status AS ENUM ('obtained', 'unlocked');

CREATE TABLE IF NOT EXISTS lists_soulcores (
    list_id UUID NOT NULL REFERENCES lists(id),
    creature_id UUID NOT NULL REFERENCES creatures(id),
    added_by_user_id UUID NOT NULL REFERENCES users(id),
    status soulcore_status NOT NULL,
    PRIMARY KEY (list_id, creature_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS lists_soulcores;
DROP TYPE soulcore_status;
-- +goose StatementEnd
