package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type NewsletterServiceInterface interface {
	Subscribe(ctx context.Context, email string) error
}

type NewsletterService struct {
	apiKey  string
	listID  string
	baseURL string
}

type EmailOctopusSubscribeRequest struct {
	APIKey string `json:"api_key"`
	Email  string `json:"email_address"`
}

type EmailOctopusErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewNewsletterService() (*NewsletterService, error) {
	apiKey := os.Getenv("EMAILOCTOPUS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("EMAILOCTOPUS_API_KEY environment variable is required")
	}

	listID := os.Getenv("EMAILOCTOPUS_LIST_ID")
	if listID == "" {
		return nil, fmt.Errorf("EMAILOCTOPUS_LIST_ID environment variable is required")
	}

	return &NewsletterService{
		apiKey:  apiKey,
		listID:  listID,
		baseURL: "https://emailoctopus.com/api/1.6",
	}, nil
}

func (s *NewsletterService) Subscribe(ctx context.Context, email string) error {
	subscribeReq := EmailOctopusSubscribeRequest{
		APIKey: s.apiKey,
		Email:  email,
	}

	jsonData, err := json.Marshal(subscribeReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/lists/%s/contacts", s.baseURL, s.listID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // Ignore close error as it's a cleanup operation
	}()

	// EmailOctopus returns 200 for success, 409 for already subscribed
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusConflict {
		return nil
	}

	// Handle error response
	var errorResp EmailOctopusErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
		return fmt.Errorf("failed to decode error response: %w", err)
	}

	return fmt.Errorf("emailoctopus error: %s - %s", errorResp.Error.Code, errorResp.Error.Message)
}
