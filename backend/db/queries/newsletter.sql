-- name: CreateNewsletterSubscriber :one
INSERT INTO newsletter_subscribers (
    email,
    emailoctopus_contact_id
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetNewsletterSubscriberByEmail :one
SELECT * FROM newsletter_subscribers
WHERE email = $1;

-- name: ConfirmNewsletterSubscription :one
UPDATE newsletter_subscribers
SET confirmed = true,
    confirmed_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE email = $1
RETURNING *;

-- name: UnsubscribeFromNewsletter :one
UPDATE newsletter_subscribers
SET unsubscribed_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE email = $1
RETURNING *;

-- name: GetActiveNewsletterSubscribers :many
SELECT * FROM newsletter_subscribers
WHERE confirmed = true
AND unsubscribed_at IS NULL
ORDER BY subscribed_at DESC;

-- name: GetNewsletterSubscriberStats :one
SELECT 
    COUNT(*) as total_subscribers,
    COUNT(*) FILTER (WHERE confirmed = true AND unsubscribed_at IS NULL) as active_subscribers,
    COUNT(*) FILTER (WHERE confirmed = false) as pending_confirmation,
    COUNT(*) FILTER (WHERE unsubscribed_at IS NOT NULL) as unsubscribed
FROM newsletter_subscribers;
