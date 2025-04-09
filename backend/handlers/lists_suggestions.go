package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
)

// GetCharacterSuggestions returns all soulcore suggestions for a character
func (h *ListsHandler) GetCharacterSuggestions(c echo.Context) error {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.ValidationError("Invalid character ID", err).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "id",
			Value:  c.Param("id"),
			Reason: "Invalid UUID format",
		})
	}

	// Get authenticated user ID from context
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return apperror.AuthorizationError("Invalid user authentication", nil).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "user_id",
			Reason: "Missing or invalid user ID in context",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperror.AuthorizationError("Invalid user ID format", err).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "user_id",
			Value:  userIDStr,
			Reason: "Invalid UUID format",
		})
	}

	ctx := c.Request().Context()

	// Verify that the character belongs to the user
	char, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("Character not found", err).WithDetails(&apperror.ValidationErrorDetails{
				Field:  "character_id",
				Value:  characterID.String(),
				Reason: "Character does not exist",
			})
		}
		return apperror.DatabaseError("Failed to get character", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "GetCharacter",
			Table:     "characters",
		})
	}

	if char.UserID != userID {
		return apperror.AuthorizationError("Character does not belong to user", nil).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "character_id",
			Value:  characterID.String(),
			Reason: "Character belongs to a different user",
		})
	}

	suggestions, err := h.store.GetCharacterSuggestions(ctx, characterID)
	if err != nil {
		return apperror.DatabaseError("Failed to get suggestions", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "GetCharacterSuggestions",
			Table:     "soulcore_suggestions",
		})
	}

	return c.JSON(http.StatusOK, suggestions)
}

// AcceptSoulcoreSuggestion accepts a soulcore suggestion for a character
func (h *ListsHandler) AcceptSoulcoreSuggestion(c echo.Context) error {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.ValidationError("Invalid character ID", err).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "id",
			Value:  c.Param("id"),
			Reason: "Invalid UUID format",
		})
	}

	var req struct {
		CreatureID uuid.UUID `json:"creature_id"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return apperror.ValidationError("Invalid request body", err).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "request_body",
			Reason: "Failed to decode JSON request",
		})
	}

	// Get authenticated user ID from context
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return apperror.AuthorizationError("Invalid user authentication", nil).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "user_id",
			Reason: "Missing or invalid user ID in context",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperror.AuthorizationError("Invalid user ID format", err).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "user_id",
			Value:  userIDStr,
			Reason: "Invalid UUID format",
		})
	}

	ctx := c.Request().Context()

	// Verify that the character belongs to the user
	char, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("Character not found", err).WithDetails(&apperror.ValidationErrorDetails{
				Field:  "character_id",
				Value:  characterID.String(),
				Reason: "Character does not exist",
			})
		}
		return apperror.DatabaseError("Failed to get character", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "GetCharacter",
			Table:     "characters",
		})
	}

	if char.UserID != userID {
		return apperror.AuthorizationError("Character does not belong to user", nil).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "character_id",
			Value:  characterID.String(),
			Reason: "Character belongs to a different user",
		})
	}

	// Add the soulcore to the character
	err = h.store.AddCharacterSoulcore(ctx, db.AddCharacterSoulcoreParams{
		CharacterID: characterID,
		CreatureID:  req.CreatureID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to add soulcore to character", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "AddCharacterSoulcore",
			Table:     "character_soulcores",
		})
	}

	// Remove the suggestion
	err = h.store.DeleteSoulcoreSuggestion(ctx, db.DeleteSoulcoreSuggestionParams{
		CharacterID: characterID,
		CreatureID:  req.CreatureID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to delete suggestion", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "DeleteSoulcoreSuggestion",
			Table:     "soulcore_suggestions",
		})
	}

	return c.NoContent(http.StatusOK)
}

// DismissSoulcoreSuggestion dismisses a soulcore suggestion without adding it to the character
func (h *ListsHandler) DismissSoulcoreSuggestion(c echo.Context) error {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.ValidationError("Invalid character ID", err).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "id",
			Value:  c.Param("id"),
			Reason: "Invalid UUID format",
		})
	}

	var req struct {
		CreatureID uuid.UUID `json:"creature_id"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return apperror.ValidationError("Invalid request body", err).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "request_body",
			Reason: "Failed to decode JSON request",
		})
	}

	// Get authenticated user ID from context
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return apperror.AuthorizationError("Invalid user authentication", nil).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "user_id",
			Reason: "Missing or invalid user ID in context",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperror.AuthorizationError("Invalid user ID format", err).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "user_id",
			Value:  userIDStr,
			Reason: "Invalid UUID format",
		})
	}

	ctx := c.Request().Context()

	// Verify that the character belongs to the user
	char, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("Character not found", err).WithDetails(&apperror.ValidationErrorDetails{
				Field:  "character_id",
				Value:  characterID.String(),
				Reason: "Character does not exist",
			})
		}
		return apperror.DatabaseError("Failed to get character", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "GetCharacter",
			Table:     "characters",
		})
	}

	if char.UserID != userID {
		return apperror.AuthorizationError("Character does not belong to user", nil).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "character_id",
			Value:  characterID.String(),
			Reason: "Character belongs to a different user",
		})
	}

	// Remove the suggestion
	err = h.store.DeleteSoulcoreSuggestion(ctx, db.DeleteSoulcoreSuggestionParams{
		CharacterID: characterID,
		CreatureID:  req.CreatureID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to delete suggestion", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "DeleteSoulcoreSuggestion",
			Table:     "soulcore_suggestions",
		})
	}

	return c.NoContent(http.StatusOK)
}
