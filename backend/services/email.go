package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/mailgun/mailgun-go/v5"
)

type EmailServiceInterface interface {
	SendVerificationEmail(ctx context.Context, email string, verificationToken string, userID string) error
}

type NewsletterServiceInterface interface {
	SubscribeToNewsletter(ctx context.Context, email string) (string, error)
	UnsubscribeFromNewsletter(ctx context.Context, email string) error
	UnsubscribeFromNewsletterByID(ctx context.Context, contactID string) error
	GetSubscriberStatus(ctx context.Context, email string) (bool, error)
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

	// Create new Mailgun client with API key only (v5)
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

// EmailOctopus API structures
type EmailOctopusContact struct {
	ID           string      `json:"id"`
	EmailAddress string      `json:"email_address"`
	Fields       interface{} `json:"fields"` // Can be array or map
	Tags         []string    `json:"tags"`
	Status       string      `json:"status"`
	CreatedAt    string      `json:"created_at"`
	UpdatedAt    string      `json:"updated_at"`
}

type EmailOctopusSubscribeRequest struct {
	EmailAddress string      `json:"email_address"`
	Fields       interface{} `json:"fields,omitempty"` // Can be array or map
	Tags         []string    `json:"tags,omitempty"`
	Status       string      `json:"status,omitempty"`
}

type EmailOctopusService struct {
	apiKey string
	listID string
	client *http.Client
}

func NewEmailOctopusService() (*EmailOctopusService, error) {
	apiKey := os.Getenv("EMAILOCTOPUS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("EMAILOCTOPUS_API_KEY environment variable is required")
	}

	listID := os.Getenv("EMAILOCTOPUS_LIST_ID")
	if listID == "" {
		return nil, fmt.Errorf("EMAILOCTOPUS_LIST_ID environment variable is required")
	}

	return &EmailOctopusService{
		apiKey: apiKey,
		listID: listID,
		client: &http.Client{},
	}, nil
}

func (s *EmailOctopusService) SubscribeToNewsletter(ctx context.Context, email string) (string, error) {
	url := fmt.Sprintf("https://emailoctopus.com/api/1.6/lists/%s/contacts", s.listID)

	reqBody := EmailOctopusSubscribeRequest{
		EmailAddress: email,
		Status:       "SUBSCRIBED",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.URL.RawQuery = fmt.Sprintf("api_key=%s", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusConflict {
		// Contact already exists, return existing contact ID
		existingContact, err := s.getContactByEmail(ctx, email)
		if err != nil {
			return "", fmt.Errorf("contact already exists but failed to retrieve: %w", err)
		}
		return existingContact.ID, nil
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var contact EmailOctopusContact
	if err := json.NewDecoder(resp.Body).Decode(&contact); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return contact.ID, nil
}

func (s *EmailOctopusService) UnsubscribeFromNewsletter(ctx context.Context, email string) error {
	contact, err := s.getContactByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to find contact: %w", err)
	}

	return s.UnsubscribeFromNewsletterByID(ctx, contact.ID)
}

func (s *EmailOctopusService) UnsubscribeFromNewsletterByID(ctx context.Context, contactID string) error {
	url := fmt.Sprintf("https://emailoctopus.com/api/1.6/lists/%s/contacts/%s", s.listID, contactID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.URL.RawQuery = fmt.Sprintf("api_key=%s", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (s *EmailOctopusService) GetSubscriberStatus(ctx context.Context, email string) (bool, error) {
	contact, err := s.getContactByEmail(ctx, email)
	if err != nil {
		return false, err
	}

	return contact.Status == "SUBSCRIBED", nil
}

func (s *EmailOctopusService) getContactByEmail(ctx context.Context, email string) (*EmailOctopusContact, error) {
	// EmailOctopus doesn't have a direct "get by email" endpoint, so we need to search
	url := fmt.Sprintf("https://emailoctopus.com/api/1.6/lists/%s/contacts", s.listID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Set("api_key", s.apiKey)
	q.Set("limit", "100") // Limit to 100 contacts per page
	req.URL.RawQuery = q.Encode()

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Data []EmailOctopusContact `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Find contact by email
	for _, contact := range result.Data {
		if contact.EmailAddress == email {
			return &contact, nil
		}
	}

	return nil, fmt.Errorf("contact not found")
}
