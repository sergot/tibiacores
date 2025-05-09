package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/auth"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
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

// CharacterPreview represents a character with their unlocked soulcores
type CharacterPreview struct {
	Character     db.Character                  `json:"character"`
	UnlockedCores []db.GetCharacterSoulcoresRow `json:"unlocked_cores"`
}

func NewUsersHandler(store db.Store, emailService services.EmailServiceInterface) *UsersHandler {
	return &UsersHandler{store: store, emailService: emailService}
}

// Login authenticates a user with email and password
func (h *UsersHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return apperror.ValidationError("Invalid request body", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "body",
				Reason: "Invalid JSON format",
			})
	}

	if req.Email == "" || req.Password == "" {
		return apperror.ValidationError("Email and password are required", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "credentials",
				Reason: "Missing required fields",
			})
	}

	ctx := c.Request().Context()

	var email pgtype.Text
	email.String = req.Email
	email.Valid = true

	user, err := h.store.GetUserByEmail(ctx, email)
	if err != nil {
		log.Printf("Failed to get user by email: %v", err)
		return apperror.AuthorizationError("Invalid email or password", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetUserByEmail",
				Table:     "users",
			}).
			Wrap(err)
	}

	// Check password
	if !user.Password.Valid || !auth.CheckPasswordHash(req.Password, user.Password.String) {
		return apperror.AuthorizationError("Invalid email or password", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "password",
				Reason: "Incorrect password",
			})
	}

	// Generate token pair
	tokenPair, err := auth.GenerateTokenPair(user.ID.String(), true)
	if err != nil {
		return apperror.InternalError("Failed to generate tokens", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "token",
				Reason: "Token generation failed",
			}).
			Wrap(err)
	}

	// Set cookies for authentication
	auth.SetTokenCookies(c, tokenPair.AccessToken, tokenPair.RefreshToken, tokenPair.ExpiresIn)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":        user.ID,
		"has_email": true,
	})
}

// Signup adds email/password to an account (new or existing)
func (h *UsersHandler) Signup(c echo.Context) error {
	var req SignupRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return apperror.ValidationError("Invalid request body", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "body",
				Reason: "Invalid JSON format",
			})
	}

	if req.Email == "" || req.Password == "" {
		return apperror.ValidationError("Email and password are required", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "credentials",
				Reason: "Missing required fields",
			})
	}

	ctx := c.Request().Context()

	// Check if user exists with this email
	var email pgtype.Text
	email.String = req.Email
	email.Valid = true

	existingUser, getUserErr := h.store.GetUserByEmail(ctx, email)

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return apperror.InternalError("Failed to process password", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "password",
				Reason: "Password hashing failed",
			}).
			Wrap(err)
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
		if claims, err := auth.ValidateAccessToken(strings.TrimPrefix(authHeader, "Bearer ")); err == nil {
			existingUserID = &claims.UserID
		}
	}

	if getUserErr == nil {
		// If it's not anonymous, prevent registration with this email
		if !existingUser.IsAnonymous {
			return apperror.ValidationError("Email already in use", nil).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "email",
					Reason: "Email already registered",
				})
		}

		// Migrate the anonymous user
		user, err = h.store.MigrateAnonymousUser(ctx, db.MigrateAnonymousUserParams{
			Email:                      email,
			Password:                   password,
			EmailVerificationToken:     verificationToken,
			EmailVerificationExpiresAt: expiresAt,
			ID:                         existingUser.ID,
		})

		if err != nil {
			return apperror.DatabaseError("Failed to migrate anonymous user", err).
				WithDetails(&apperror.DatabaseErrorDetails{
					Operation: "MigrateAnonymousUser",
					Table:     "users",
				}).
				Wrap(err)
		}
	} else if existingUserID != nil {
		// If user has auth token, use that ID to migrate their anonymous account
		existingID, err := uuid.Parse(*existingUserID)
		if err != nil {
			return apperror.ValidationError("Invalid user ID", err).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "user_id",
					Value:  *existingUserID,
					Reason: "Invalid UUID format",
				})
		}

		// Verify this is actually an anonymous account before migrating
		existingUser, err := h.store.GetUserByID(ctx, existingID)
		if err != nil {
			return apperror.DatabaseError("Failed to get user", err).
				WithDetails(&apperror.DatabaseErrorDetails{
					Operation: "GetUserByID",
					Table:     "users",
				}).
				Wrap(err)
		}

		// Only migrate anonymous accounts
		if !existingUser.IsAnonymous {
			return apperror.ValidationError("Cannot update existing account", nil).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "user_id",
					Value:  existingID.String(),
					Reason: "Account already has credentials",
				})
		}

		// Make sure the anonymous account doesn't already have an email
		if existingUser.Email.Valid && existingUser.Email.String != "" {
			return apperror.ValidationError("Account already has an email", nil).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "email",
					Reason: "Email already exists on this account",
				})
		}

		// Migrate anonymous user
		user, err = h.store.MigrateAnonymousUser(ctx, db.MigrateAnonymousUserParams{
			Email:                      email,
			Password:                   password,
			EmailVerificationToken:     verificationToken,
			EmailVerificationExpiresAt: expiresAt,
			ID:                         existingID,
		})

		if err != nil {
			return apperror.DatabaseError("Failed to migrate anonymous user", err).
				WithDetails(&apperror.DatabaseErrorDetails{
					Operation: "MigrateAnonymousUser",
					Table:     "users",
				}).
				Wrap(err)
		}
	} else {
		// Create a new user
		user, err = h.store.CreateUser(ctx, db.CreateUserParams{
			Email:                      email,
			Password:                   password,
			EmailVerificationToken:     verificationToken,
			EmailVerificationExpiresAt: expiresAt,
			EmailVerified:              false, // Require email verification
		})

		if err != nil {
			return apperror.DatabaseError("Failed to create user", err).
				WithDetails(&apperror.DatabaseErrorDetails{
					Operation: "CreateUser",
					Table:     "users",
				}).
				Wrap(err)
		}
	}

	// Send verification email
	if err := h.emailService.SendVerificationEmail(ctx, email.String, verificationToken.String(), user.ID.String()); err != nil {
		log.Printf("Error sending verification email: %v", err)
		// Continue despite email sending error - user can request another email later
	}

	// Generate token pair
	tokenPair, err := auth.GenerateTokenPair(user.ID.String(), true)
	if err != nil {
		return apperror.InternalError("Failed to generate tokens", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "token",
				Reason: "Token generation failed",
			}).
			Wrap(err)
	}

	// Set cookies for authentication
	auth.SetTokenCookies(c, tokenPair.AccessToken, tokenPair.RefreshToken, tokenPair.ExpiresIn)

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id":         user.ID,
		"has_email":  true,
		"email":      user.Email.String,
		"verified":   user.EmailVerified,
		"expires_in": tokenPair.ExpiresIn,
	})
}

func (h *UsersHandler) GetCharactersByUserId(c echo.Context) error {
	ctx := c.Request().Context()

	requestedUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return apperror.ValidationError("Invalid user ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  c.Param("user_id"),
				Reason: "Invalid UUID format",
			})
	}

	// Get authenticated user ID from context
	authedUserIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return apperror.AuthorizationError("Invalid user authentication", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Reason: "Not found in context",
			})
	}

	authedUserID, err := uuid.Parse(authedUserIDStr)
	if err != nil {
		return apperror.AuthorizationError("Invalid user ID format", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  authedUserIDStr,
				Reason: "Invalid UUID format",
			})
	}

	// Only allow users to view their own characters
	if requestedUserID != authedUserID {
		return apperror.AuthorizationError("Cannot access other users' characters", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  requestedUserID.String(),
				Reason: "Access denied to other user's characters",
			})
	}

	characters, err := h.store.GetCharactersByUserID(ctx, requestedUserID)
	if err != nil {
		log.Printf("Error getting characters for user %s: %v", requestedUserID, err)
		return apperror.DatabaseError("Failed to get characters", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetCharactersByUserID",
				Table:     "characters",
			}).
			Wrap(err)
	}

	return c.JSON(http.StatusOK, characters)
}

// GetUserLists returns all lists where the user is either an author or a member
func (h *UsersHandler) GetUserLists(c echo.Context) error {
	requestedUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return apperror.ValidationError("Invalid user ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  c.Param("user_id"),
				Reason: "Invalid UUID format",
			})
	}

	// Get authenticated user ID from context
	authedUserIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return apperror.AuthorizationError("Invalid user authentication", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Reason: "Not found in context",
			})
	}

	authedUserID, err := uuid.Parse(authedUserIDStr)
	if err != nil {
		return apperror.AuthorizationError("Invalid user ID format", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  authedUserIDStr,
				Reason: "Invalid UUID format",
			})
	}

	// Only allow users to view their own lists
	if requestedUserID != authedUserID {
		return apperror.AuthorizationError("Cannot access other users' lists", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  requestedUserID.String(),
				Reason: "Access denied to other user's lists",
			})
	}

	ctx := c.Request().Context()
	lists, err := h.store.GetUserLists(ctx, requestedUserID)
	if err != nil {
		return apperror.DatabaseError("Failed to get lists", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetUserLists",
				Table:     "lists",
			}).
			Wrap(err)
	}

	return c.JSON(http.StatusOK, lists)
}

// GetCharacter returns details about a specific character
func (h *UsersHandler) GetCharacter(c echo.Context) error {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.ValidationError("Invalid character ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "id",
				Value:  c.Param("id"),
				Reason: "Invalid UUID format",
			})
	}

	// Get authenticated user ID from context
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return apperror.AuthorizationError("Invalid user authentication", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Reason: "Not found in context",
			})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperror.AuthorizationError("Invalid user ID format", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  userIDStr,
				Reason: "Invalid UUID format",
			})
	}

	ctx := c.Request().Context()

	// Get character details
	character, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("Character not found", err).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "id",
					Value:  characterID.String(),
					Reason: "Character does not exist",
				})
		}
		return apperror.DatabaseError("Failed to get character", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetCharacter",
				Table:     "characters",
			}).
			Wrap(err)
	}

	// Verify character belongs to user
	if character.UserID != userID {
		return apperror.AuthorizationError("Character does not belong to user", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  userID.String(),
				Reason: "Access denied to other user's character",
			})
	}

	return c.JSON(http.StatusOK, character)
}

// GetCharacterSoulcores returns all unlocked soulcores for a character
func (h *UsersHandler) GetCharacterSoulcores(c echo.Context) error {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.ValidationError("Invalid character ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "id",
				Value:  c.Param("id"),
				Reason: "Invalid UUID format",
			})
	}

	// Get authenticated user ID from context
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return apperror.AuthorizationError("Invalid user authentication", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Reason: "Not found in context",
			})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperror.AuthorizationError("Invalid user ID format", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  userIDStr,
				Reason: "Invalid UUID format",
			})
	}

	ctx := c.Request().Context()

	// Verify character belongs to user
	character, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("Character not found", err).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "id",
					Value:  characterID.String(),
					Reason: "Character does not exist",
				})
		}
		return apperror.DatabaseError("Failed to get character", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetCharacter",
				Table:     "characters",
			}).
			Wrap(err)
	}

	if character.UserID != userID {
		return apperror.AuthorizationError("Character does not belong to user", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  userID.String(),
				Reason: "Access denied to other user's character",
			})
	}

	// Get unlocked soulcores
	soulcores, err := h.store.GetCharacterSoulcores(ctx, characterID)
	if err != nil {
		return apperror.DatabaseError("Failed to get character soulcores", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetCharacterSoulcores",
				Table:     "character_soulcores",
			}).
			Wrap(err)
	}

	return c.JSON(http.StatusOK, soulcores)
}

// RemoveCharacterSoulcore removes a soulcore from a character
func (h *UsersHandler) RemoveCharacterSoulcore(c echo.Context) error {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.ValidationError("Invalid character ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "id",
				Value:  c.Param("id"),
				Reason: "Invalid UUID format",
			})
	}

	creatureID, err := uuid.Parse(c.Param("creature_id"))
	if err != nil {
		return apperror.ValidationError("Invalid creature ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "creature_id",
				Value:  c.Param("creature_id"),
				Reason: "Invalid UUID format",
			})
	}

	// Get authenticated user ID from context
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return apperror.AuthorizationError("Invalid user authentication", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Reason: "Not found in context",
			})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperror.AuthorizationError("Invalid user ID format", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  userIDStr,
				Reason: "Invalid UUID format",
			})
	}

	ctx := c.Request().Context()

	// Verify character belongs to user
	character, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("Character not found", err).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "id",
					Value:  characterID.String(),
					Reason: "Character does not exist",
				})
		}
		return apperror.DatabaseError("Failed to get character", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetCharacter",
				Table:     "characters",
			}).
			Wrap(err)
	}

	if character.UserID != userID {
		return apperror.AuthorizationError("Character does not belong to user", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  userID.String(),
				Reason: "Access denied to other user's character",
			})
	}

	// Remove the soulcore
	err = h.store.RemoveCharacterSoulcore(ctx, db.RemoveCharacterSoulcoreParams{
		CharacterID: characterID,
		CreatureID:  creatureID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to remove soul core", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "RemoveCharacterSoulcore",
				Table:     "character_soulcores",
			}).
			Wrap(err)
	}

	return c.NoContent(http.StatusOK)
}

// AddCharacterSoulcore adds a new soul core to a character
func (h *UsersHandler) AddCharacterSoulcore(c echo.Context) error {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.ValidationError("Invalid character ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "id",
				Value:  c.Param("id"),
				Reason: "Invalid UUID format",
			})
	}

	var req struct {
		CreatureID uuid.UUID `json:"creature_id"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return apperror.ValidationError("Invalid request body", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "body",
				Reason: "Invalid JSON format",
			})
	}

	// Get authenticated user ID from context
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return apperror.AuthorizationError("Invalid user authentication", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Reason: "Not found in context",
			})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperror.AuthorizationError("Invalid user ID format", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  userIDStr,
				Reason: "Invalid UUID format",
			})
	}

	ctx := c.Request().Context()

	// Verify character belongs to user
	character, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("Character not found", err).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "id",
					Value:  characterID.String(),
					Reason: "Character does not exist",
				})
		}
		return apperror.DatabaseError("Failed to get character", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetCharacter",
				Table:     "characters",
			}).
			Wrap(err)
	}

	if character.UserID != userID {
		return apperror.AuthorizationError("Character does not belong to user", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  userID.String(),
				Reason: "Access denied to other user's character",
			})
	}

	// Add the soulcore to the character
	err = h.store.AddCharacterSoulcore(ctx, db.AddCharacterSoulcoreParams{
		CharacterID: characterID,
		CreatureID:  req.CreatureID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to add soul core", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "AddCharacterSoulcore",
				Table:     "character_soulcores",
			}).
			Wrap(err)
	}

	return c.NoContent(http.StatusOK)
}

// GetPendingSuggestions returns all characters with pending soulcore suggestions
func (h *UsersHandler) GetPendingSuggestions(c echo.Context) error {
	// Get authenticated user ID from context
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return apperror.AuthorizationError("Invalid user authentication", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Reason: "Not found in context",
			})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperror.AuthorizationError("Invalid user ID format", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  userIDStr,
				Reason: "Invalid UUID format",
			})
	}

	ctx := c.Request().Context()

	suggestions, err := h.store.GetPendingSuggestionsForUser(ctx, userID)
	if err != nil {
		return apperror.DatabaseError("Failed to get pending suggestions", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetPendingSuggestionsForUser",
				Table:     "suggestions",
			}).
			Wrap(err)
	}

	return c.JSON(http.StatusOK, suggestions)
}

// VerifyEmail verifies a user's email address using the verification token
func (h *UsersHandler) VerifyEmail(c echo.Context) error {
	userID, err := uuid.Parse(c.QueryParam("user_id"))
	if err != nil {
		return apperror.ValidationError("Invalid user ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  c.QueryParam("user_id"),
				Reason: "Invalid UUID format",
			})
	}

	token, err := uuid.Parse(c.QueryParam("token"))
	if err != nil {
		return apperror.ValidationError("Invalid verification token", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "token",
				Value:  c.QueryParam("token"),
				Reason: "Invalid UUID format",
			})
	}

	ctx := c.Request().Context()

	err = h.store.VerifyEmail(ctx, db.VerifyEmailParams{
		ID:                     userID,
		EmailVerificationToken: token,
	})
	if err != nil {
		log.Printf("Failed to verify email: %v", err)
		return apperror.ValidationError("Invalid or expired verification token", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "token",
				Value:  token.String(),
				Reason: "Token verification failed",
			}).
			Wrap(err)
	}

	return c.NoContent(http.StatusOK)
}

// GetUser returns details about a specific user
func (h *UsersHandler) GetUser(c echo.Context) error {
	requestedUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return apperror.ValidationError("Invalid user ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  c.Param("user_id"),
				Reason: "Invalid UUID format",
			})
	}

	// Get authenticated user ID from context
	authedUserIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return apperror.AuthorizationError("Invalid user authentication", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Reason: "Not found in context",
			})
	}

	authedUserID, err := uuid.Parse(authedUserIDStr)
	if err != nil {
		return apperror.AuthorizationError("Invalid user ID format", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  authedUserIDStr,
				Reason: "Invalid UUID format",
			})
	}

	// Only allow users to view their own details
	if requestedUserID != authedUserID {
		return apperror.AuthorizationError("Cannot access other users' details", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  requestedUserID.String(),
				Reason: "Access denied to other user's details",
			})
	}

	ctx := c.Request().Context()

	// Get user details using the queries object
	user, err := h.store.GetUserByID(ctx, requestedUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("User not found", err).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "user_id",
					Value:  requestedUserID.String(),
					Reason: "User does not exist",
				})
		}
		return apperror.DatabaseError("Failed to get user details", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetUserByID",
				Table:     "users",
			}).
			Wrap(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"email":          user.Email.String,
		"email_verified": user.EmailVerified,
	})
}

// GetCharacterPublic returns character details including their unlocked soulcores
func (h *UsersHandler) GetCharacterPublic(c echo.Context) error {
	characterName := c.Param("name")
	if characterName == "" {
		return apperror.ValidationError("Character name is required", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "name",
				Reason: "Character name cannot be empty",
			})
	}

	ctx := c.Request().Context()

	// Get character details by name
	character, err := h.store.GetCharacterByName(ctx, characterName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("Character not found", err).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "name",
					Value:  characterName,
					Reason: "Character does not exist",
				})
		}
		return apperror.DatabaseError("Failed to get character", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetCharacterByName",
				Table:     "characters",
			}).
			Wrap(err)
	}

	// Get unlocked soulcores
	soulcores, err := h.store.GetCharacterSoulcores(ctx, character.ID)
	if err != nil {
		return apperror.DatabaseError("Failed to get character soulcores", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetCharacterSoulcores",
				Table:     "character_soulcores",
			}).
			Wrap(err)
	}

	// Return combined response
	return c.JSON(http.StatusOK, CharacterPreview{
		Character:     character,
		UnlockedCores: soulcores,
	})
}
