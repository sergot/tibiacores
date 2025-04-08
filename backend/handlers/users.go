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
	"github.com/sergot/tibiacores/backend/pkg/errors"
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
		return errors.NewInvalidRequestError(errors.ErrInvalidRequest)
	}

	if req.Email == "" || req.Password == "" {
		return errors.NewValidationError(errors.ErrInvalidRequest)
	}

	ctx := c.Request().Context()

	var email pgtype.Text
	email.String = req.Email
	email.Valid = true

	user, err := h.store.GetUserByEmail(ctx, email)
	if err != nil {
		// Log the error but return a generic message to the client
		return errors.NewUnauthorizedError(errors.ErrUnauthorized)
	}

	// Check password
	if !user.Password.Valid || !auth.CheckPasswordHash(req.Password, user.Password.String) {
		return errors.NewUnauthorizedError(errors.ErrUnauthorized)
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID.String(), true)
	if err != nil {
		return errors.NewInternalError(err)
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
		return errors.NewInvalidRequestError(err)
	}

	if req.Email == "" || req.Password == "" {
		return errors.NewValidationError(errors.ErrInvalidRequest)
	}

	ctx := c.Request().Context()

	// Check if user exists with this email
	var email pgtype.Text
	email.String = req.Email
	email.Valid = true

	existingUser, getUserErr := h.store.GetUserByEmail(ctx, email)
	if getUserErr == nil && !existingUser.IsAnonymous {
		return errors.NewConflictError(errors.ErrConflict)
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return errors.NewInternalError(err)
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
			return errors.NewInvalidRequestError(parseErr)
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
			return errors.NewDatabaseError(err)
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
			return errors.NewDatabaseError(err)
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
			return errors.NewDatabaseError(err)
		}
	}

	// Generate new token
	token, err := auth.GenerateToken(user.ID.String(), true)
	if err != nil {
		return errors.NewInternalError(err)
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
		return errors.NewInvalidRequestError(err).
			WithOperation("parse_user_id").
			WithResource("user")
	}

	// Get authenticated user ID from context
	authedUserIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return errors.NewUnauthorizedError(errors.ErrUnauthorized).
			WithOperation("get_user_id").
			WithResource("user")
	}

	authedUserID, err := uuid.Parse(authedUserIDStr)
	if err != nil {
		return errors.NewUnauthorizedError(err).
			WithOperation("parse_user_id").
			WithResource("user")
	}

	// Only allow users to view their own characters
	if requestedUserID != authedUserID {
		return errors.NewForbiddenError(errors.ErrForbidden).
			WithOperation("verify_user_ownership").
			WithResource("user")
	}

	characters, err := h.store.GetCharactersByUserID(ctx, requestedUserID)
	if err != nil {
		return errors.NewDatabaseError(err).
			WithOperation("get_characters").
			WithResource("character")
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
		return errors.NewInvalidRequestError(err).
			WithOperation("parse_character_id").
			WithResource("character")
	}

	// Get authenticated user ID from context
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return errors.NewUnauthorizedError(errors.ErrUnauthorized).
			WithOperation("get_user_id").
			WithResource("user")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.NewUnauthorizedError(err).
			WithOperation("parse_user_id").
			WithResource("user")
	}

	ctx := c.Request().Context()

	// Get character details
	character, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		return errors.NewDatabaseError(err).
			WithOperation("get_character").
			WithResource("character")
	}

	// Verify character belongs to user
	if character.UserID != userID {
		return errors.NewForbiddenError(errors.ErrForbidden).
			WithOperation("verify_character_ownership").
			WithResource("character")
	}

	return c.JSON(http.StatusOK, character)
}

// GetCharacterSoulcores returns all unlocked soulcores for a character
func (h *UsersHandler) GetCharacterSoulcores(c echo.Context) error {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.NewInvalidRequestError(err).
			WithOperation("parse_character_id").
			WithResource("character")
	}

	// Get authenticated user ID from context
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return errors.NewUnauthorizedError(errors.ErrUnauthorized).
			WithOperation("get_user_id").
			WithResource("user")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.NewUnauthorizedError(err).
			WithOperation("parse_user_id").
			WithResource("user")
	}

	ctx := c.Request().Context()

	// Verify character belongs to user
	character, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		return errors.NewDatabaseError(err).
			WithOperation("get_character").
			WithResource("character")
	}
	if character.UserID != userID {
		return errors.NewForbiddenError(errors.ErrForbidden).
			WithOperation("verify_character_ownership").
			WithResource("character")
	}

	// Get unlocked soulcores
	soulcores, err := h.store.GetCharacterSoulcores(ctx, characterID)
	if err != nil {
		return errors.NewDatabaseError(err).
			WithOperation("get_character_soulcores").
			WithResource("soulcore")
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

// AddCharacterSoulcore adds a new soul core to a character
func (h *UsersHandler) AddCharacterSoulcore(c echo.Context) error {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid character ID")
	}

	var req struct {
		CreatureID uuid.UUID `json:"creature_id"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
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

	// Add the soulcore to the character
	err = h.store.AddCharacterSoulcore(ctx, db.AddCharacterSoulcoreParams{
		CharacterID: characterID,
		CreatureID:  req.CreatureID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add soul core")
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
		return errors.NewInvalidRequestError(err).
			WithOperation("parse_user_id").
			WithResource("user")
	}

	token, err := uuid.Parse(c.QueryParam("token"))
	if err != nil {
		return errors.NewInvalidRequestError(err).
			WithOperation("parse_verification_token").
			WithResource("token")
	}

	ctx := c.Request().Context()

	err = h.store.VerifyEmail(ctx, db.VerifyEmailParams{
		ID:                     userID,
		EmailVerificationToken: token,
	})
	if err != nil {
		return errors.NewDatabaseError(err).
			WithOperation("verify_email").
			WithResource("user")
	}

	return c.NoContent(http.StatusOK)
}

// GetUser returns details about a specific user
func (h *UsersHandler) GetUser(c echo.Context) error {
	requestedUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return errors.NewInvalidRequestError(err).
			WithOperation("parse_user_id").
			WithResource("user")
	}

	// Get authenticated user ID from context
	authedUserIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return errors.NewUnauthorizedError(errors.ErrUnauthorized).
			WithOperation("get_user_id").
			WithResource("user")
	}

	authedUserID, err := uuid.Parse(authedUserIDStr)
	if err != nil {
		return errors.NewUnauthorizedError(err).
			WithOperation("parse_user_id").
			WithResource("user")
	}

	// Only allow users to view their own details
	if requestedUserID != authedUserID {
		return errors.NewForbiddenError(errors.ErrForbidden).
			WithOperation("verify_user_ownership").
			WithResource("user")
	}

	ctx := c.Request().Context()

	// Get user details using the queries object
	user, err := h.store.GetUserByID(ctx, requestedUserID)
	if err != nil {
		return errors.NewDatabaseError(err).
			WithOperation("get_user").
			WithResource("user")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"email":          user.Email.String,
		"email_verified": user.EmailVerified,
	})
}
