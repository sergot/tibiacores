package services

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type TibiaDataService struct {
	baseURL string
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
	}
}

func (s *TibiaDataService) GetCharacter(name string) (*TibiaCharacter, error) {
	resp, err := http.Get(fmt.Sprintf("%s/character/%s", s.baseURL, name))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("character not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch character data: %d", resp.StatusCode)
	}

	var response TibiaDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
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
