package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/auth"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/services"
)

type UsersHandler struct {
	store        db.Store
	emailService services.EmailServiceInterface
}

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	UserID   string `json:"user_id,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewUsersHandler(store db.Store, emailService services.EmailServiceInterface) *UsersHandler {
	return &UsersHandler{store: store, emailService: emailService}
}

// Login authenticates a user with email and password
func (h *UsersHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
	}

	ctx := c.Request().Context()

	var email pgtype.Text
	email.String = req.Email
	email.Valid = true

	user, err := h.store.GetUserByEmail(ctx, email)
	if err != nil {
		log.Printf("Failed to get user by email: %v", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
	}

	// Check password
	if !user.Password.Valid || !auth.CheckPasswordHash(req.Password, user.Password.String) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID.String(), true)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
	}

	// Set token in X-Auth-Token header
	c.Response().Header().Set("X-Auth-Token", token)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":        user.ID,
		"has_email": true,
	})
}

// Signup adds email/password to an account (new or existing)
func (h *UsersHandler) Signup(c echo.Context) error {
	var req SignupRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
	}

	ctx := c.Request().Context()

	// Check if user exists with this email
	var email pgtype.Text
	email.String = req.Email
	email.Valid = true

	existingUser, getUserErr := h.store.GetUserByEmail(ctx, email)
	if getUserErr == nil && !existingUser.IsAnonymous {
		return echo.NewHTTPError(http.StatusConflict, "email already in use")
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to process password")
	}

	var password pgtype.Text
	password.String = hashedPassword
	password.Valid = true

	verificationToken := uuid.New()
	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	}

	var user db.User

	// Check for existing session token
	authHeader := c.Request().Header.Get("Authorization")
	var existingUserID *string
	if authHeader != "" {
		if claims, err := auth.ValidateToken(strings.TrimPrefix(authHeader, "Bearer ")); err == nil {
			existingUserID = &claims.UserID
		}
	}

	if existingUserID != nil {
		// Migrate existing anonymous user
		userID, parseErr := uuid.Parse(*existingUserID)
		if parseErr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID format")
		}

		user, err = h.store.MigrateAnonymousUser(ctx, db.MigrateAnonymousUserParams{
			Email:                      email,
			Password:                   password,
			EmailVerificationToken:     verificationToken,
			EmailVerificationExpiresAt: expiresAt,
			ID:                         userID,
		})
		if err != nil {
			log.Printf("Failed to migrate anonymous user: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to migrate user")
		}
	} else if getUserErr == nil && existingUser.IsAnonymous {
		// Update existing anonymous user found by email
		user, err = h.store.MigrateAnonymousUser(ctx, db.MigrateAnonymousUserParams{
			Email:                      email,
			Password:                   password,
			EmailVerificationToken:     verificationToken,
			EmailVerificationExpiresAt: expiresAt,
			ID:                         existingUser.ID,
		})
		if err != nil {
			log.Printf("Failed to migrate existing user: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update user")
		}
	} else {
		// Create new user
		user, err = h.store.CreateUser(ctx, db.CreateUserParams{
			Email:                      email,
			Password:                   password,
			EmailVerificationToken:     verificationToken,
			EmailVerificationExpiresAt: expiresAt,
		})
		if err != nil {
			log.Printf("Failed to create user: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
		}
	}

	// Generate new token
	token, err := auth.GenerateToken(user.ID.String(), true)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
	}

	// Send verification email
	if err := h.emailService.SendVerificationEmail(ctx, email.String, verificationToken.String(), user.ID.String()); err != nil {
		log.Printf("Failed to send verification email: %v", err)
		// Don't return error to client, as the account was created successfully
	}

	// Set token in X-Auth-Token header
	c.Response().Header().Set("X-Auth-Token", token)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":        user.ID,
		"has_email": true,
	})
}

func (h *UsersHandler) GetCharactersByUserId(c echo.Context) error {
	ctx := c.Request().Context()

	requestedUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	// Get authenticated user ID from context
	authedUserIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user authentication")
	}

	authedUserID, err := uuid.Parse(authedUserIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID format")
	}

	// Only allow users to view their own characters
	if requestedUserID != authedUserID {
		return echo.NewHTTPError(http.StatusForbidden, "cannot access other users' characters")
	}

	characters, err := h.store.GetCharactersByUserID(ctx, requestedUserID)
	if err != nil {
		log.Printf("Error getting characters for user %s: %v", requestedUserID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get characters")
	}

	return c.JSON(http.StatusOK, characters)
}

// GetUserLists returns all lists where the user is either an author or a member
func (h *UsersHandler) GetUserLists(c echo.Context) error {

	requestedUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	// Get authenticated user ID from context
	authedUserIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user authentication")
	}

	authedUserID, err := uuid.Parse(authedUserIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID format")
	}

	// Only allow users to view their own lists
	if requestedUserID != authedUserID {
		return echo.NewHTTPError(http.StatusForbidden, "cannot access other users' lists")
	}

	ctx := c.Request().Context()
	lists, err := h.store.GetUserLists(ctx, requestedUserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get lists")
	}

	return c.JSON(http.StatusOK, lists)
}

// GetCharacter returns details about a specific character
func (h *UsersHandler) GetCharacter(c echo.Context) error {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid character ID")
	}

	// Get authenticated user ID from context
	userIDStr := c.Get("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID format")
	}

	ctx := c.Request().Context()

	// Get character details
	character, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get character")
	}

	// Verify character belongs to user
	if character.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "character does not belong to user")
	}

	return c.JSON(http.StatusOK, character)
}

// GetCharacterSoulcores returns all unlocked soulcores for a character
func (h *UsersHandler) GetCharacterSoulcores(c echo.Context) error {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid character ID")
	}

	// Get authenticated user ID from context
	userIDStr := c.Get("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID format")
	}

	ctx := c.Request().Context()

	// Verify character belongs to user
	character, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get character")
	}
	if character.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "character does not belong to user")
	}

	// Get unlocked soulcores
	soulcores, err := h.store.GetCharacterSoulcores(ctx, characterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get character soulcores")
	}

	return c.JSON(http.StatusOK, soulcores)
}

// RemoveCharacterSoulcore removes a soulcore from a character
func (h *UsersHandler) RemoveCharacterSoulcore(c echo.Context) error {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid character ID")
	}

	creatureID, err := uuid.Parse(c.Param("creature_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid creature ID")
	}

	// Get authenticated user ID from context
	userIDStr := c.Get("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID format")
	}

	ctx := c.Request().Context()

	// Verify character belongs to user
	character, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get character")
	}
	if character.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "character does not belong to user")
	}

	// Remove the soulcore
	err = h.store.RemoveCharacterSoulcore(ctx, db.RemoveCharacterSoulcoreParams{
		CharacterID: characterID,
		CreatureID:  creatureID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove soul core")
	}

	return c.NoContent(http.StatusOK)
}

// GetPendingSuggestions returns all characters with pending soulcore suggestions
func (h *UsersHandler) GetPendingSuggestions(c echo.Context) error {
	// Get authenticated user ID from context
	userIDStr := c.Get("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID format")
	}

	ctx := c.Request().Context()

	suggestions, err := h.store.GetPendingSuggestionsForUser(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get pending suggestions")
	}

	return c.JSON(http.StatusOK, suggestions)
}

// VerifyEmail verifies a user's email address using the verification token
func (h *UsersHandler) VerifyEmail(c echo.Context) error {
	userID, err := uuid.Parse(c.QueryParam("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	token, err := uuid.Parse(c.QueryParam("token"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid verification token")
	}

	ctx := c.Request().Context()

	err = h.store.VerifyEmail(ctx, db.VerifyEmailParams{
		ID:                     userID,
		EmailVerificationToken: token,
	})
	if err != nil {
		log.Printf("Failed to verify email: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid or expired verification token")
	}

	return c.NoContent(http.StatusOK)
}

// GetUser returns details about a specific user
func (h *UsersHandler) GetUser(c echo.Context) error {
	requestedUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	// Get authenticated user ID from context
	authedUserIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user authentication")
	}

	authedUserID, err := uuid.Parse(authedUserIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID format")
	}

	// Only allow users to view their own details
	if requestedUserID != authedUserID {
		return echo.NewHTTPError(http.StatusForbidden, "cannot access other users' details")
	}

	ctx := c.Request().Context()

	// Get user details using the queries object
	user, err := h.store.GetUserByID(ctx, requestedUserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user details")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"email":          user.Email.String,
		"email_verified": user.EmailVerified,
	})
}
