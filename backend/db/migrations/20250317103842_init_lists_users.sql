-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS lists_users (
    list_id UUID NOT NULL REFERENCES lists(id),
    user_id UUID NOT NULL REFERENCES users(id),
    character_id UUID NOT NULL REFERENCES characters(id),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (list_id, user_id, character_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS lists_users;
-- +goose StatementEnd
