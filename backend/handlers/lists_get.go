package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
)

func (h *ListsHandler) GetList(c echo.Context) error {
	listID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperror.ValidationError("Invalid list ID", err).WithDetails(&apperror.ValidationErrorDetails{
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

	// Get list details
	list, err := h.store.GetList(ctx, listID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("List not found", err).WithDetails(&apperror.ValidationErrorDetails{
				Field:  "list_id",
				Value:  listID.String(),
				Reason: "List does not exist",
			})
		}
		return apperror.DatabaseError("Failed to get list", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "GetList",
			Table:     "lists",
		})
	}

	// Get member stats
	members, err := h.store.GetListMembers(ctx, listID)
	if err != nil {
		return apperror.DatabaseError("Failed to get list members", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "GetListMembers",
			Table:     "lists_users",
		})
	}

	isMember := false
	for _, member := range members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return apperror.AuthorizationError("User is not a member of this list", nil).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "user_id",
			Value:  userID.String(),
			Reason: "User is not a member of the list",
		})
	}

	memberStats := make([]MemberStats, len(members))
	for i, m := range members {
		memberStats[i] = MemberStats{
			UserID:        m.UserID,
			CharacterName: m.CharacterName,
			ObtainedCount: m.ObtainedCount,
			UnlockedCount: m.UnlockedCount,
			IsActive:      m.IsActive,
		}
	}

	// Get soul cores
	soulCores, err := h.store.GetListSoulcores(ctx, listID)
	if err != nil {
		return apperror.DatabaseError("Failed to get soul cores", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "GetListSoulcores",
			Table:     "list_soulcores",
		})
	}

	return c.JSON(http.StatusOK, ListDetailResponse{
		ID:        list.ID,
		AuthorID:  list.AuthorID,
		Name:      list.Name,
		ShareCode: list.ShareCode,
		World:     list.World,
		CreatedAt: list.CreatedAt.Time,
		UpdatedAt: list.UpdatedAt.Time,
		Members:   memberStats,
		SoulCores: soulCores,
	})
}

// ListPreviewResponse represents the public preview of a list
type ListPreviewResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	World       string    `json:"world"`
	MemberCount int       `json:"member_count"`
}

// GetListPreview returns basic information about a list by its share code
func (h *ListsHandler) GetListPreview(c echo.Context) error {
	shareCode, err := uuid.Parse(c.Param("share_code"))
	if err != nil {
		return apperror.ValidationError("Invalid share code", err).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "share_code",
			Value:  c.Param("share_code"),
			Reason: "Invalid UUID format",
		})
	}

	ctx := c.Request().Context()

	// Get the list by share code
	list, err := h.store.GetListByShareCode(ctx, shareCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("List not found", err).WithDetails(&apperror.ValidationErrorDetails{
				Field:  "share_code",
				Value:  shareCode.String(),
				Reason: "List does not exist",
			})
		}
		return apperror.DatabaseError("Failed to retrieve list", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "GetListByShareCode",
			Table:     "lists",
		})
	}

	// Get member count
	members, err := h.store.GetMembers(ctx, list.ID)
	if err != nil {
		return apperror.DatabaseError("Failed to get list members", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "GetMembers",
			Table:     "lists_users",
		})
	}

	return c.JSON(http.StatusOK, ListPreviewResponse{
		ID:          list.ID,
		Name:        list.Name,
		World:       list.World,
		MemberCount: len(members),
	})
}
