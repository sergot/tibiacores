package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a registered user with authentication details
type User struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email             string             `json:"email" bson:"email"`
	Password          string             `json:"-" bson:"password"` // Password is never returned in JSON
	PlayerID          primitive.ObjectID `json:"player_id" bson:"player_id"`
	EmailVerified     bool               `json:"email_verified" bson:"email_verified"`
	VerificationToken string             `json:"-" bson:"verification_token,omitempty"`
	GoogleID          string             `json:"-" bson:"google_id,omitempty"`
	DiscordID         string             `json:"-" bson:"discord_id,omitempty"`
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" bson:"updated_at"`
}

// AuthResponse is the response sent after successful authentication
type AuthResponse struct {
	Token    string `json:"token"`
	PlayerID string `json:"player_id"`
	Username string `json:"username"`
}

// RegistrationRequest represents the data needed to register a new user
type RegistrationRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	SessionID   string `json:"session_id,omitempty"`   // Optional, for converting anonymous accounts
	RedirectURL string `json:"redirect_url,omitempty"` // For email verification
}

// LoginRequest represents the data needed to log in
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// OAuthRequest represents the data needed for OAuth login
type OAuthRequest struct {
	Provider    string `json:"provider"` // "google" or "discord"
	AccessToken string `json:"access_token"`
	SessionID   string `json:"session_id,omitempty"` // Optional, for converting anonymous accounts
}

// VerifyEmailRequest represents the data needed to verify an email
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// ResetPasswordRequest represents the data needed to reset a password
type ResetPasswordRequest struct {
	Email string `json:"email"`
}

// UpdatePasswordRequest represents the data needed to update a password
type UpdatePasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}
