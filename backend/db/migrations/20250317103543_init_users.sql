-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    is_anonymous BOOLEAN NOT NULL DEFAULT TRUE,
    session_token UUID,
    email TEXT,
    password TEXT,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verification_token UUID,
    email_verification_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
-- XXX: add email index for lookup?
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
