-- +goose Up
-- +goose StatementBegin
CREATE TABLE newsletter_subscribers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    subscribed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    confirmed_at TIMESTAMP WITH TIME ZONE,
    unsubscribed_at TIMESTAMP WITH TIME ZONE,
    emailoctopus_contact_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add index for email lookups
CREATE INDEX idx_newsletter_subscribers_email ON newsletter_subscribers(email);

-- Add index for active subscribers
CREATE INDEX idx_newsletter_subscribers_active ON newsletter_subscribers(confirmed, unsubscribed_at) WHERE confirmed = true AND unsubscribed_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE newsletter_subscribers;
-- +goose StatementEnd
