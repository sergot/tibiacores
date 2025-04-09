package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
)

// UpdateSoulcoreStatus updates the status of a soul core in a list
func (h *ListsHandler) UpdateSoulcoreStatus(c echo.Context) error {
	listID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.ValidationError("Invalid list ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "list_id",
				Value:  c.Param("id"),
				Reason: "Invalid UUID format",
			})
	}

	var req struct {
		CreatureID uuid.UUID         `json:"creature_id"`
		Status     db.SoulcoreStatus `json:"status"`
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

	// Check if user is a member of the list
	isMember, err := h.store.IsUserListMember(ctx, db.IsUserListMemberParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to check list membership", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "IsUserListMember",
				Table:     "list_members",
			})
	}

	if !isMember {
		return apperror.AuthorizationError("User is not a member of this list", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "list_id",
				Value:  listID.String(),
				Reason: "User is not a member",
			})
	}

	// Get the soulcore to check ownership
	soulcore, err := h.store.GetListSoulcore(ctx, db.GetListSoulcoreParams{
		ListID:     listID,
		CreatureID: req.CreatureID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("Soulcore not found", err).
				WithDetails(&apperror.DatabaseErrorDetails{
					Operation: "GetListSoulcore",
					Table:     "list_soulcores",
				})
		}
		return apperror.DatabaseError("Failed to check soulcore ownership", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetListSoulcore",
				Table:     "list_soulcores",
			})
	}

	// Get list details to check if user is the owner
	list, err := h.store.GetList(ctx, listID)
	if err != nil {
		return apperror.DatabaseError("Failed to get list details", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetList",
				Table:     "lists",
			})
	}

	// Allow both the soulcore adder and list owner to modify it
	if soulcore.AddedByUserID != userID && list.AuthorID != userID {
		return apperror.AuthorizationError("Only the list owner or the user who added the soulcore can modify it", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  userID.String(),
				Reason: "Not authorized to modify soulcore",
			})
	}

	// Update soul core status
	err = h.store.UpdateSoulcoreStatus(ctx, db.UpdateSoulcoreStatusParams{
		ListID:     listID,
		CreatureID: req.CreatureID,
		Status:     req.Status,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to update soul core status", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "UpdateSoulcoreStatus",
				Table:     "list_soulcores",
			})
	}

	// If the status is being set to unlocked, create suggestions for all members
	if req.Status == "unlocked" {
		err = h.store.CreateSoulcoreSuggestions(ctx, db.CreateSoulcoreSuggestionsParams{
			ID:         listID,
			CreatureID: req.CreatureID,
		})
		if err != nil {
			// Don't fail the request if suggestions creation fails
			log.Printf("Failed to create soulcore suggestions: %v", err)
		}
	}

	return c.NoContent(http.StatusOK)
}

// AddSoulcore adds a new soul core to a list
func (h *ListsHandler) AddSoulcore(c echo.Context) error {
	listID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.ValidationError("Invalid list ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "list_id",
				Value:  c.Param("id"),
				Reason: "Invalid UUID format",
			})
	}

	var req struct {
		CreatureID uuid.UUID         `json:"creature_id"`
		Status     db.SoulcoreStatus `json:"status"`
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

	// Check if user is a member of the list
	isMember, err := h.store.IsUserListMember(ctx, db.IsUserListMemberParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to check list membership", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "IsUserListMember",
				Table:     "list_members",
			})
	}

	if !isMember {
		return apperror.AuthorizationError("User is not a member of this list", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "list_id",
				Value:  listID.String(),
				Reason: "User is not a member",
			})
	}

	// Add soul core with the user ID who added it
	err = h.store.AddSoulcoreToList(ctx, db.AddSoulcoreToListParams{
		ListID:        listID,
		CreatureID:    req.CreatureID,
		Status:        req.Status,
		AddedByUserID: userID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to add soul core", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "AddSoulcoreToList",
				Table:     "list_soulcores",
			})
	}

	return c.NoContent(http.StatusOK)
}

// RemoveSoulcore removes a soul core from a list
func (h *ListsHandler) RemoveSoulcore(c echo.Context) error {
	listID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.ValidationError("Invalid list ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "list_id",
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

	// Check if user is a member of the list
	isMember, err := h.store.IsUserListMember(ctx, db.IsUserListMemberParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to check list membership", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "IsUserListMember",
				Table:     "list_members",
			})
	}

	if !isMember {
		return apperror.AuthorizationError("User is not a member of this list", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "list_id",
				Value:  listID.String(),
				Reason: "User is not a member",
			})
	}

	// Get the soulcore to check ownership
	soulcore, err := h.store.GetListSoulcore(ctx, db.GetListSoulcoreParams{
		ListID:     listID,
		CreatureID: creatureID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("Soulcore not found", err).
				WithDetails(&apperror.DatabaseErrorDetails{
					Operation: "GetListSoulcore",
					Table:     "list_soulcores",
				})
		}
		return apperror.DatabaseError("Failed to check soulcore ownership", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetListSoulcore",
				Table:     "list_soulcores",
			})
	}

	// Get list details to check if user is the owner
	list, err := h.store.GetList(ctx, listID)
	if err != nil {
		return apperror.DatabaseError("Failed to get list details", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetList",
				Table:     "lists",
			})
	}

	// Allow both the soulcore adder and list owner to remove it
	if soulcore.AddedByUserID != userID && list.AuthorID != userID {
		return apperror.AuthorizationError("Only the list owner or the user who added the soulcore can remove it", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  userID.String(),
				Reason: "Not authorized to remove soulcore",
			})
	}

	// Delete the soulcore from the list
	err = h.store.RemoveListSoulcore(ctx, db.RemoveListSoulcoreParams{
		ListID:     listID,
		CreatureID: creatureID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to remove soul core", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "RemoveListSoulcore",
				Table:     "list_soulcores",
			})
	}

	return c.NoContent(http.StatusOK)
}
