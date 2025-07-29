package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
	"github.com/sergot/tibiacores/backend/services"
)

type NewsletterHandler struct {
	newsletterService services.NewsletterServiceInterface
}

type NewsletterSubscribeRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type NewsletterSubscribeResponse struct {
	Message string `json:"message"`
}

func NewNewsletterHandler(newsletterService services.NewsletterServiceInterface) *NewsletterHandler {
	return &NewsletterHandler{
		newsletterService: newsletterService,
	}
}

// Subscribe subscribes an email to the newsletter
func (h *NewsletterHandler) Subscribe(c echo.Context) error {
	var req NewsletterSubscribeRequest
	if err := c.Bind(&req); err != nil {
		return apperror.ValidationError("Invalid request body", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "body",
				Value:  "",
				Reason: "must be valid JSON",
			})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return apperror.ValidationError("Invalid email address", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "email",
				Value:  req.Email,
				Reason: "must be a valid email address",
			})
	}

	// Subscribe to newsletter
	if err := h.newsletterService.Subscribe(c.Request().Context(), req.Email); err != nil {
		return apperror.ExternalServiceError("Failed to subscribe to newsletter", err).
			WithDetails(&apperror.ExternalServiceErrorDetails{
				Service:   "EmailOctopus",
				Operation: "subscribe",
				Endpoint:  "newsletter subscription",
			})
	}

	return c.JSON(http.StatusOK, NewsletterSubscribeResponse{
		Message: "Successfully subscribed to newsletter",
	})
}
