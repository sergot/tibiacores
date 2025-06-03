-- +goose Up
CREATE TABLE IF NOT EXISTS list_user_read_status (
    user_id UUID NOT NULL REFERENCES users(id),
    list_id UUID NOT NULL REFERENCES lists(id),
    last_read_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, list_id)
);

-- +goose Down
DROP TABLE IF EXISTS list_user_read_status;
