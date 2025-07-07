package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
	"github.com/sergot/tibiacores/backend/services"
)

type NewsletterHandler struct {
	store             db.Store
	newsletterService services.NewsletterServiceInterface
}

func NewNewsletterHandler(store db.Store, newsletterService services.NewsletterServiceInterface) *NewsletterHandler {
	return &NewsletterHandler{
		store:             store,
		newsletterService: newsletterService,
	}
}

type SubscribeRequest struct {
	Email string `json:"email"`
}

type UnsubscribeRequest struct {
	Email string `json:"email"`
}

// Subscribe handles newsletter subscription
func (h *NewsletterHandler) Subscribe(c echo.Context) error {
	var req SubscribeRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return apperror.ValidationError("Invalid request body", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "body",
				Reason: "Invalid JSON format",
			})
	}

	if req.Email == "" {
		return apperror.ValidationError("Email is required", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "email",
				Reason: "Email cannot be empty",
			})
	}

	ctx := c.Request().Context()

	// Check if already subscribed
	existing, err := h.store.GetNewsletterSubscriberByEmail(ctx, req.Email)
	if err == nil {
		// Already exists
		if existing.UnsubscribedAt.Valid {
			// Was unsubscribed, resubscribe
			_, err := h.newsletterService.SubscribeToNewsletter(ctx, req.Email)
			if err != nil {
				return apperror.ExternalServiceError("Failed to resubscribe to newsletter", err).
					WithDetails(&apperror.ExternalServiceErrorDetails{
						Service:   "EmailOctopus",
						Operation: "Subscribe",
						Endpoint:  "/contacts",
					}).
					Wrap(err)
			}

			// Update database
			_, err = h.store.ConfirmNewsletterSubscription(ctx, req.Email)
			if err != nil {
				return apperror.DatabaseError("Failed to update subscription", err).
					WithDetails(&apperror.DatabaseErrorDetails{
						Operation: "ConfirmNewsletterSubscription",
						Table:     "newsletter_subscribers",
					}).
					Wrap(err)
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"message": "Successfully resubscribed to newsletter",
				"status":  "resubscribed",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Already subscribed to newsletter",
			"status":  "already_subscribed",
		})
	}

	// Subscribe to EmailOctopus
	contactID, err := h.newsletterService.SubscribeToNewsletter(ctx, req.Email)
	if err != nil {
		return apperror.ExternalServiceError("Failed to subscribe to newsletter", err).
			WithDetails(&apperror.ExternalServiceErrorDetails{
				Service:   "EmailOctopus",
				Operation: "Subscribe",
				Endpoint:  "/contacts",
			}).
			Wrap(err)
	}

	// Store in database
	var contactIDText pgtype.Text
	contactIDText.String = contactID
	contactIDText.Valid = true

	_, err = h.store.CreateNewsletterSubscriber(ctx, db.CreateNewsletterSubscriberParams{
		Email:                 req.Email,
		EmailoctopusContactID: contactIDText,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to store subscription", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "CreateNewsletterSubscriber",
				Table:     "newsletter_subscribers",
			}).
			Wrap(err)
	}

	// Auto-confirm the subscription since EmailOctopus handles double opt-in
	_, err = h.store.ConfirmNewsletterSubscription(ctx, req.Email)
	if err != nil {
		return apperror.DatabaseError("Failed to confirm subscription", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "ConfirmNewsletterSubscription",
				Table:     "newsletter_subscribers",
			}).
			Wrap(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Successfully subscribed to newsletter",
		"status":  "subscribed",
	})
}

// Unsubscribe handles newsletter unsubscription
func (h *NewsletterHandler) Unsubscribe(c echo.Context) error {
	var req UnsubscribeRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return apperror.ValidationError("Invalid request body", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "body",
				Reason: "Invalid JSON format",
			})
	}

	if req.Email == "" {
		return apperror.ValidationError("Email is required", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "email",
				Reason: "Email cannot be empty",
			})
	}

	ctx := c.Request().Context()

	// Check if subscribed
	subscriber, err := h.store.GetNewsletterSubscriberByEmail(ctx, req.Email)
	if err != nil {
		return apperror.ValidationError("Email not found in newsletter", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "email",
				Value:  req.Email,
				Reason: "Email is not subscribed to newsletter",
			})
	}

	// Unsubscribe from EmailOctopus using stored contact ID
	if subscriber.EmailoctopusContactID.Valid && subscriber.EmailoctopusContactID.String != "" {
		err = h.newsletterService.UnsubscribeFromNewsletterByID(ctx, subscriber.EmailoctopusContactID.String)
		if err != nil {
			return apperror.ExternalServiceError("Failed to unsubscribe from newsletter", err).
				WithDetails(&apperror.ExternalServiceErrorDetails{
					Service:   "EmailOctopus",
					Operation: "Unsubscribe",
					Endpoint:  "/contacts",
				}).
				Wrap(err)
		}
	}

	// Update database
	_, err = h.store.UnsubscribeFromNewsletter(ctx, req.Email)
	if err != nil {
		return apperror.DatabaseError("Failed to update unsubscription", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "UnsubscribeFromNewsletter",
				Table:     "newsletter_subscribers",
			}).
			Wrap(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Successfully unsubscribed from newsletter",
		"status":  "unsubscribed",
	})
}

// GetStats returns newsletter subscription statistics (admin only)
func (h *NewsletterHandler) GetStats(c echo.Context) error {
	ctx := c.Request().Context()

	stats, err := h.store.GetNewsletterSubscriberStats(ctx)
	if err != nil {
		return apperror.DatabaseError("Failed to get newsletter stats", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetNewsletterSubscriberStats",
				Table:     "newsletter_subscribers",
			}).
			Wrap(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total_subscribers":    stats.TotalSubscribers,
		"active_subscribers":   stats.ActiveSubscribers,
		"pending_confirmation": stats.PendingConfirmation,
		"unsubscribed":         stats.Unsubscribed,
	})
}

// CheckSubscriptionStatus checks if an email is subscribed to the newsletter
func (h *NewsletterHandler) CheckSubscriptionStatus(c echo.Context) error {
	email := c.QueryParam("email")
	if email == "" {
		return apperror.ValidationError("Email is required", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "email",
				Reason: "Email query parameter is required",
			})
	}

	ctx := c.Request().Context()

	subscriber, err := h.store.GetNewsletterSubscriberByEmail(ctx, email)
	if err != nil {
		// Not subscribed
		return c.JSON(http.StatusOK, map[string]interface{}{
			"subscribed": false,
		})
	}

	isSubscribed := subscriber.Confirmed && !subscriber.UnsubscribedAt.Valid

	return c.JSON(http.StatusOK, map[string]interface{}{
		"subscribed": isSubscribed,
	})
}
