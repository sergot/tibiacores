package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/sergot/tibiacores/backend/pkg/apperror"
)

// httpClient is a shared HTTP client with proper timeout and transport configuration
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

// TibiaDataServiceInterface defines the methods for interacting with TibiaData API
type TibiaDataServiceInterface interface {
	GetCharacter(name string) (*TibiaCharacter, error)
	VerifyCharacterClaim(name, verificationCode string) (bool, error)
}

// Ensure TibiaDataService implements TibiaDataServiceInterface
var _ TibiaDataServiceInterface = (*TibiaDataService)(nil)

type TibiaDataService struct {
	baseURL string
	client  *http.Client
}

type TibiaCharacter struct {
	Name    string `json:"name"`
	World   string `json:"world"`
	Comment string `json:"comment"`
}

type TibiaDataResponse struct {
	Character struct {
		Character TibiaCharacter `json:"character"`
	} `json:"character"`
}

func NewTibiaDataService() *TibiaDataService {
	return &TibiaDataService{
		baseURL: "https://api.tibiadata.com/v4",
		client:  httpClient,
	}
}

func (s *TibiaDataService) GetCharacter(name string) (*TibiaCharacter, error) {
	resp, err := s.client.Get(fmt.Sprintf("%s/character/%s", s.baseURL, name))
	if err != nil {
		return nil, apperror.ExternalServiceError("failed to reach tibiadata", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, apperror.NotFoundError("character not found", nil).WithDetails(&apperror.ExternalServiceErrorDetails{
			Service:   "TibiaData",
			Operation: "GetCharacter",
			Endpoint:  name,
		})
	}
	if resp.StatusCode != http.StatusOK {
		return nil, apperror.ExternalServiceError("failed to fetch character data", fmt.Errorf("status code: %d", resp.StatusCode)).
			WithDetails(&apperror.ExternalServiceErrorDetails{
				Service:   "TibiaData",
				Operation: "GetCharacter",
				Endpoint:  name,
			})
	}

	var response TibiaDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, apperror.ExternalServiceError("failed to decode response", err)
	}

	return &response.Character.Character, nil
}

func (s *TibiaDataService) VerifyCharacterClaim(name, verificationCode string) (bool, error) {
	character, err := s.GetCharacter(name)
	if err != nil {
		return false, err
	}

	return character.Comment != "" && character.Comment == verificationCode, nil
}
