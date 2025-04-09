package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
)

// GetListMembersWithUnlocks returns all members of a list with their unlocked soulcores
func (h *ListsHandler) GetListMembersWithUnlocks(c echo.Context) error {
	listID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.ValidationError("Invalid list ID", err).
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
				Reason: "User ID not found in context",
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
				Reason: "User is not a member of this list",
			})
	}

	members, err := h.store.GetListMembersWithUnlocks(ctx, listID)
	if err != nil {
		return apperror.DatabaseError("Failed to get list members", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetListMembersWithUnlocks",
				Table:     "list_members",
			})
	}

	return c.JSON(http.StatusOK, members)
}
