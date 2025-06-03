-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS list_chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id UUID NOT NULL REFERENCES lists(id),
    user_id UUID NOT NULL REFERENCES users(id),
    character_id UUID NOT NULL REFERENCES characters(id),
    message TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index for faster retrieval of messages by list
CREATE INDEX IF NOT EXISTS idx_list_chat_messages_list_id ON list_chat_messages(list_id);
-- Create index for ordering by created_at
CREATE INDEX IF NOT EXISTS idx_list_chat_messages_created_at ON list_chat_messages(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_list_chat_messages_created_at;
DROP INDEX IF EXISTS idx_list_chat_messages_list_id;
DROP TABLE IF EXISTS list_chat_messages;
-- +goose StatementEnd
