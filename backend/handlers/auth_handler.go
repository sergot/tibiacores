package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fiendlist/backend/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	DB *mongo.Database
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(db *mongo.Database) *AuthHandler {
	return &AuthHandler{DB: db}
}

// Register handles user registration
func (h *AuthHandler) Register(c echo.Context) error {
	var req models.RegistrationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Validate email and password
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email and password are required"})
	}

	if len(req.Password) < 8 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password must be at least 8 characters"})
	}

	// Check if email already exists
	var existingUser models.User
	err := h.DB.Collection("users").FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Email already registered"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	// Generate verification token
	token, err := generateRandomToken(32)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate verification token"})
	}

	// Create user
	now := time.Now()
	user := models.User{
		ID:                primitive.NewObjectID(),
		Email:             req.Email,
		Password:          string(hashedPassword),
		EmailVerified:     false,
		VerificationToken: token,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	// If session ID is provided, find the player and link it
	var playerID primitive.ObjectID
	if req.SessionID != "" {
		var player models.Player
		err := h.DB.Collection("players").FindOne(context.Background(), bson.M{"session_id": req.SessionID}).Decode(&player)
		if err == nil {
			// Update player to mark as not anonymous
			update := bson.M{
				"$set": bson.M{
					"is_anonymous": false,
					"updated_at":   now,
				},
			}
			_, err = h.DB.Collection("players").UpdateOne(context.Background(), bson.M{"_id": player.ID}, update)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update player"})
			}

			playerID = player.ID
			user.PlayerID = playerID
		}
	}

	// If no player was found or linked, create a new one
	if playerID.IsZero() {
		// Create a new player
		player := models.Player{
			ID:          primitive.NewObjectID(),
			Username:    strings.Split(req.Email, "@")[0], // Use part of email as username
			Characters:  []models.Character{},
			IsAnonymous: false,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		_, err = h.DB.Collection("players").InsertOne(context.Background(), player)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create player"})
		}

		playerID = player.ID
		user.PlayerID = playerID
	}

	// Insert user
	_, err = h.DB.Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}

	// Send verification email
	err = h.sendVerificationEmail(req.Email, token, req.RedirectURL)
	if err != nil {
		// Log the error but don't fail the request
		fmt.Printf("Failed to send verification email: %v\n", err)
	}

	// Generate JWT token
	jwtToken, err := generateJWT(user.ID.Hex(), playerID.Hex())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	// Get player username
	var player models.Player
	err = h.DB.Collection("players").FindOne(context.Background(), bson.M{"_id": playerID}).Decode(&player)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
	}

	return c.JSON(http.StatusCreated, models.AuthResponse{
		Token:    jwtToken,
		PlayerID: playerID.Hex(),
		Username: player.Username,
	})
}

// Login handles user login
func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Find user by email
	var user models.User
	err := h.DB.Collection("users").FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid email or password"})
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid email or password"})
	}

	// Generate JWT token
	token, err := generateJWT(user.ID.Hex(), user.PlayerID.Hex())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	// Get player username
	var player models.Player
	err = h.DB.Collection("players").FindOne(context.Background(), bson.M{"_id": user.PlayerID}).Decode(&player)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
	}

	return c.JSON(http.StatusOK, models.AuthResponse{
		Token:    token,
		PlayerID: user.PlayerID.Hex(),
		Username: player.Username,
	})
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Token is required"})
	}

	// Find user by verification token
	var user models.User
	err := h.DB.Collection("users").FindOne(context.Background(), bson.M{"verification_token": token}).Decode(&user)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Invalid or expired token"})
	}

	// Update user to mark email as verified
	update := bson.M{
		"$set": bson.M{
			"email_verified":     true,
			"verification_token": "",
			"updated_at":         time.Now(),
		},
	}
	_, err = h.DB.Collection("users").UpdateOne(context.Background(), bson.M{"_id": user.ID}, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to verify email"})
	}

	// Redirect to frontend if redirect URL is provided
	redirectURL := c.QueryParam("redirect")
	if redirectURL != "" {
		return c.Redirect(http.StatusFound, redirectURL)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Email verified successfully"})
}

// GoogleLogin handles Google OAuth login
func (h *AuthHandler) GoogleLogin(c echo.Context) error {
	var req models.OAuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if req.Provider != "google" || req.AccessToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid provider or access token"})
	}

	// TODO: Verify Google token and get user info
	// This would typically involve calling Google's API to verify the token
	// For now, we'll just simulate this with a placeholder

	// For demonstration purposes, we'll assume we got the email from Google
	googleEmail := "google_user@example.com" // This would come from Google's API
	googleID := "google_123456789"           // This would be the Google user ID

	// Check if user exists with this Google ID
	var user models.User
	err := h.DB.Collection("users").FindOne(context.Background(), bson.M{"google_id": googleID}).Decode(&user)

	now := time.Now()
	var playerID primitive.ObjectID

	if err == mongo.ErrNoDocuments {
		// User doesn't exist, create a new one

		// Check if the email is already registered
		err = h.DB.Collection("users").FindOne(context.Background(), bson.M{"email": googleEmail}).Decode(&user)
		if err == nil {
			// Email exists but not linked to Google, update the user
			update := bson.M{
				"$set": bson.M{
					"google_id":  googleID,
					"updated_at": now,
				},
			}
			_, err = h.DB.Collection("users").UpdateOne(context.Background(), bson.M{"_id": user.ID}, update)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user"})
			}

			playerID = user.PlayerID
		} else {
			// Handle session ID if provided (convert anonymous account)
			if req.SessionID != "" {
				var player models.Player
				err := h.DB.Collection("players").FindOne(context.Background(), bson.M{"session_id": req.SessionID}).Decode(&player)
				if err == nil {
					// Update player to mark as not anonymous
					update := bson.M{
						"$set": bson.M{
							"is_anonymous": false,
							"updated_at":   now,
						},
					}
					_, err = h.DB.Collection("players").UpdateOne(context.Background(), bson.M{"_id": player.ID}, update)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update player"})
					}

					playerID = player.ID
				}
			}

			// If no player was found or linked, create a new one
			if playerID.IsZero() {
				// Create a new player
				player := models.Player{
					ID:          primitive.NewObjectID(),
					Username:    strings.Split(googleEmail, "@")[0], // Use part of email as username
					Characters:  []models.Character{},
					IsAnonymous: false,
					CreatedAt:   now,
					UpdatedAt:   now,
				}

				_, err = h.DB.Collection("players").InsertOne(context.Background(), player)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create player"})
				}

				playerID = player.ID
			}

			// Create new user
			user = models.User{
				ID:            primitive.NewObjectID(),
				Email:         googleEmail,
				GoogleID:      googleID,
				PlayerID:      playerID,
				EmailVerified: true, // Google emails are verified
				CreatedAt:     now,
				UpdatedAt:     now,
			}

			_, err = h.DB.Collection("users").InsertOne(context.Background(), user)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
			}
		}
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	} else {
		// User exists, get player ID
		playerID = user.PlayerID
	}

	// Generate JWT token
	token, err := generateJWT(user.ID.Hex(), playerID.Hex())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	// Get player username
	var player models.Player
	err = h.DB.Collection("players").FindOne(context.Background(), bson.M{"_id": playerID}).Decode(&player)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
	}

	return c.JSON(http.StatusOK, models.AuthResponse{
		Token:    token,
		PlayerID: playerID.Hex(),
		Username: player.Username,
	})
}

// DiscordLogin handles Discord OAuth login
func (h *AuthHandler) DiscordLogin(c echo.Context) error {
	var req models.OAuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if req.Provider != "discord" || req.AccessToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid provider or access token"})
	}

	// TODO: Verify Discord token and get user info
	// This would typically involve calling Discord's API to verify the token
	// For now, we'll just simulate this with a placeholder

	// For demonstration purposes, we'll assume we got the email from Discord
	discordEmail := "discord_user@example.com" // This would come from Discord's API
	discordID := "discord_123456789"           // This would be the Discord user ID

	// Check if user exists with this Discord ID
	var user models.User
	err := h.DB.Collection("users").FindOne(context.Background(), bson.M{"discord_id": discordID}).Decode(&user)

	now := time.Now()
	var playerID primitive.ObjectID

	if err == mongo.ErrNoDocuments {
		// User doesn't exist, create a new one

		// Check if the email is already registered
		err = h.DB.Collection("users").FindOne(context.Background(), bson.M{"email": discordEmail}).Decode(&user)
		if err == nil {
			// Email exists but not linked to Discord, update the user
			update := bson.M{
				"$set": bson.M{
					"discord_id": discordID,
					"updated_at": now,
				},
			}
			_, err = h.DB.Collection("users").UpdateOne(context.Background(), bson.M{"_id": user.ID}, update)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user"})
			}

			playerID = user.PlayerID
		} else {
			// Handle session ID if provided (convert anonymous account)
			if req.SessionID != "" {
				var player models.Player
				err := h.DB.Collection("players").FindOne(context.Background(), bson.M{"session_id": req.SessionID}).Decode(&player)
				if err == nil {
					// Update player to mark as not anonymous
					update := bson.M{
						"$set": bson.M{
							"is_anonymous": false,
							"updated_at":   now,
						},
					}
					_, err = h.DB.Collection("players").UpdateOne(context.Background(), bson.M{"_id": player.ID}, update)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update player"})
					}

					playerID = player.ID
				}
			}

			// If no player was found or linked, create a new one
			if playerID.IsZero() {
				// Create a new player
				player := models.Player{
					ID:          primitive.NewObjectID(),
					Username:    strings.Split(discordEmail, "@")[0], // Use part of email as username
					Characters:  []models.Character{},
					IsAnonymous: false,
					CreatedAt:   now,
					UpdatedAt:   now,
				}

				_, err = h.DB.Collection("players").InsertOne(context.Background(), player)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create player"})
				}

				playerID = player.ID
			}

			// Create new user
			user = models.User{
				ID:            primitive.NewObjectID(),
				Email:         discordEmail,
				DiscordID:     discordID,
				PlayerID:      playerID,
				EmailVerified: true, // Discord emails are verified
				CreatedAt:     now,
				UpdatedAt:     now,
			}

			_, err = h.DB.Collection("users").InsertOne(context.Background(), user)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
			}
		}
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	} else {
		// User exists, get player ID
		playerID = user.PlayerID
	}

	// Generate JWT token
	token, err := generateJWT(user.ID.Hex(), playerID.Hex())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	// Get player username
	var player models.Player
	err = h.DB.Collection("players").FindOne(context.Background(), bson.M{"_id": playerID}).Decode(&player)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
	}

	return c.JSON(http.StatusOK, models.AuthResponse{
		Token:    token,
		PlayerID: playerID.Hex(),
		Username: player.Username,
	})
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c echo.Context) error {
	// Get JWT token from context
	userToken := c.Get("user")
	if userToken == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	// Extract token
	token, ok := userToken.(*jwt.Token)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}

	// Extract claims
	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		// Try with regular MapClaims
		regularClaims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid claims"})
		}

		// Get user ID from claims
		userIDStr, ok := regularClaims["user_id"].(string)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user ID in token"})
		}

		// Convert user ID to ObjectID
		userObjID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID format"})
		}

		// Find user by ID
		var userData models.User
		err = h.DB.Collection("users").FindOne(context.Background(), bson.M{"_id": userObjID}).Decode(&userData)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		}

		// Find player by ID
		var player models.Player
		err = h.DB.Collection("players").FindOne(context.Background(), bson.M{"_id": userData.PlayerID}).Decode(&player)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
		}

		// Set is_anonymous to false for authenticated users
		player.IsAnonymous = false

		// Return player data
		return c.JSON(http.StatusOK, player)
	}

	// Get user ID from claims
	userIDStr, ok := (*claims)["user_id"].(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user ID in token"})
	}

	// Convert user ID to ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID format"})
	}

	// Find user by ID
	var userData models.User
	err = h.DB.Collection("users").FindOne(context.Background(), bson.M{"_id": userObjID}).Decode(&userData)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	// Find player by ID
	var player models.Player
	err = h.DB.Collection("players").FindOne(context.Background(), bson.M{"_id": userData.PlayerID}).Decode(&player)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
	}

	// Set is_anonymous to false for authenticated users
	player.IsAnonymous = false

	// Return player data
	return c.JSON(http.StatusOK, player)
}

// Helper functions

// generateRandomToken generates a random token for email verification
func generateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// generateJWT generates a JWT token
func generateJWT(userID, playerID string) (string, error) {
	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default_jwt_secret_change_in_production"
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   userID,
		"player_id": playerID,
		"exp":       time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	})

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// sendVerificationEmail sends an email with a verification link
func (h *AuthHandler) sendVerificationEmail(email, token, redirectURL string) error {
	// TODO: Implement email sending
	// For now, we'll just log the verification link
	verificationLink := fmt.Sprintf("http://localhost:8080/api/auth/verify-email?token=%s", token)
	if redirectURL != "" {
		verificationLink += fmt.Sprintf("&redirect=%s", redirectURL)
	}
	fmt.Printf("Verification link for %s: %s\n", email, verificationLink)
	return nil
}
