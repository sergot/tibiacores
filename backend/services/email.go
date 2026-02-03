package services

import (
	"context"
	"fmt"
	"os"

	"github.com/mailgun/mailgun-go/v5"
)

type EmailServiceInterface interface {
	SendVerificationEmail(ctx context.Context, email string, verificationToken string, userID string) error
}

type EmailService struct {
	mg          *mailgun.Client
	domain      string
	fromAddress string
}

func NewEmailService() (*EmailService, error) {
	domain := os.Getenv("MAILGUN_DOMAIN")
	if domain == "" {
		return nil, fmt.Errorf("MAILGUN_DOMAIN environment variable is required")
	}

	apiKey := os.Getenv("MAILGUN_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("MAILGUN_API_KEY environment variable is required")
	}

	fromAddress := os.Getenv("EMAIL_FROM_ADDRESS")
	if fromAddress == "" {
		fromAddress = fmt.Sprintf("noreply@%s", domain)
	}

	// Create new Mailgun client using only API key (v5 standard)
	mg := mailgun.NewMailgun(apiKey)

	// Set EU endpoint for non-production environments
	err := mg.SetAPIBase(mailgun.APIBaseEU)
	if err != nil {
		return nil, fmt.Errorf("failed to set API base: %w", err)
	}

	return &EmailService{
		mg:          mg,
		domain:      domain,
		fromAddress: fromAddress,
	}, nil
}

func (s *EmailService) SendVerificationEmail(ctx context.Context, email string, verificationToken string, userID string) error {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // Default for development
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s&user_id=%s",
		frontendURL, verificationToken, userID)

	message := mailgun.NewMessage(s.domain, s.fromAddress, "Verify your email address", fmt.Sprintf("Click the link below to verify your email address:\n\n%s\n\nThis link will expire in 24 hours.", verificationURL), email)

	_, err := s.mg.Send(ctx, message)
	if err != nil {
		return err
	}

	return nil
}
