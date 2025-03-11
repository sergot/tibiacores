package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}

// GenerateShareCode generates a unique share code for a list
func GenerateShareCode() string {
	return uuid.New().String()[:8]
}

// GenerateSessionID generates a unique session ID for a player
func GenerateSessionID() string {
	return uuid.New().String()
}

// GenerateJWT generates a JWT token for a player
func GenerateJWT(playerID string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["player_id"] = playerID
	claims["exp"] = time.Now().Add(time.Hour * 24 * 30).Unix() // 30 days

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your_jwt_secret_key_change_in_production"
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// LoadCreaturesFromJSON loads creatures from the JSON file
func LoadCreaturesFromJSON(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	return data, nil
}
