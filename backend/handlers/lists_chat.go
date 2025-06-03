package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
)

// ChatMessage represents a chat message response
type ChatMessage struct {
	ID            string    `json:"id"`
	ListID        string    `json:"list_id"`
	UserID        string    `json:"user_id"`
	CharacterName string    `json:"character_name"`
	Message       string    `json:"message"`
	CreatedAt     time.Time `json:"created_at"`
}

// ChatNotification represents a chat notification response
type ChatNotification struct {
	ListID            string    `json:"list_id"`
	ListName          string    `json:"list_name"`
	LastMessageTime   time.Time `json:"last_message_time"`
	UnreadCount       int32     `json:"unread_count"`
	LastCharacterName string    `json:"last_character_name"`
}

// CreateChatMessage adds a new chat message to a list
func (h *ListsHandler) CreateChatMessage(c echo.Context) error {
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
	member, err := h.store.IsUserListMember(ctx, db.IsUserListMemberParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to check list membership", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "IsUserListMember",
				Table:     "lists_users",
			})
	}
	if !member {
		return apperror.AuthorizationError("User is not a member of this list", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "list_id",
				Value:  listID.String(),
				Reason: "User is not a member of this list",
			})
	}

	// Parse message from request body
	var messageReq struct {
		Message     string `json:"message" validate:"required"`
		CharacterID string `json:"character_id" validate:"required,uuid"`
	}

	if err := c.Bind(&messageReq); err != nil {
		return apperror.ValidationError("Invalid request body", err)
	}

	if err := c.Validate(&messageReq); err != nil {
		return apperror.ValidationError("Validation failed", err)
	}

	characterID, err := uuid.Parse(messageReq.CharacterID)
	if err != nil {
		return apperror.ValidationError("Invalid character ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "character_id",
				Value:  messageReq.CharacterID,
				Reason: "Invalid UUID format",
			})
	}

	// Verify the character belongs to the user
	character, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		return apperror.DatabaseError("Failed to get character", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetCharacter",
				Table:     "characters",
			})
	}

	if character.UserID != userID {
		return apperror.AuthorizationError("Character does not belong to user", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "character_id",
				Value:  characterID.String(),
				Reason: "Character does not belong to user",
			})
	}

	// Create chat message
	message, err := h.store.CreateChatMessage(ctx, db.CreateChatMessageParams{
		ListID:      listID,
		UserID:      userID,
		CharacterID: characterID,
		Message:     messageReq.Message,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to create chat message", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "CreateChatMessage",
				Table:     "list_chat_messages",
			})
	}

	// Format response
	response := ChatMessage{
		ID:            message.ID.String(),
		ListID:        message.ListID.String(),
		UserID:        message.UserID.String(),
		CharacterName: character.Name,
		Message:       message.Message,
		CreatedAt:     message.CreatedAt.Time,
	}

	return c.JSON(http.StatusCreated, response)
}

// GetChatMessages returns chat messages for a list
func (h *ListsHandler) GetChatMessages(c echo.Context) error {
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
	member, err := h.store.IsUserListMember(ctx, db.IsUserListMemberParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to check list membership", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "IsUserListMember",
				Table:     "lists_users",
			})
	}
	if !member {
		return apperror.AuthorizationError("User is not a member of this list", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "list_id",
				Value:  listID.String(),
				Reason: "User is not a member of this list",
			})
	}

	// Get pagination parameters
	limit := int32(50) // default limit
	offset := int32(0) // default offset

	limitParam := c.QueryParam("limit")
	if limitParam != "" {
		parsedLimit, err := strconv.ParseInt(limitParam, 10, 32)
		if err != nil {
			return apperror.ValidationError("Invalid limit parameter", err).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "limit",
					Value:  limitParam,
					Reason: "Must be a valid 32-bit integer",
				})
		}
		if parsedLimit > 0 {
			limit = int32(parsedLimit)
		}
	}

	offsetParam := c.QueryParam("offset")
	if offsetParam != "" {
		parsedOffset, err := strconv.ParseInt(offsetParam, 10, 32)
		if err != nil {
			return apperror.ValidationError("Invalid offset parameter", err).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "offset",
					Value:  offsetParam,
					Reason: "Must be a valid 32-bit integer",
				})
		}
		if parsedOffset >= 0 {
			offset = int32(parsedOffset)
		}
	}

	// Check for since parameter for real-time updates
	sinceParam := c.QueryParam("since")

	if sinceParam != "" {
		sinceTime, err := time.Parse(time.RFC3339, sinceParam)
		if err != nil {
			return apperror.ValidationError("Invalid since timestamp", err).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "since",
					Value:  sinceParam,
					Reason: "Invalid timestamp format, expected RFC3339",
				})
		}

		// Convert time.Time to pgtype.Timestamptz
		var pgTime pgtype.Timestamptz
		pgTime.Time = sinceTime
		pgTime.Valid = true

		byTimestampMessages, err := h.store.GetChatMessagesByTimestamp(ctx, db.GetChatMessagesByTimestampParams{
			ListID:    listID,
			CreatedAt: pgTime,
		})
		if err != nil {
			return apperror.DatabaseError("Failed to get chat messages", err).
				WithDetails(&apperror.DatabaseErrorDetails{
					Operation: "GetChatMessagesByTimestamp",
					Table:     "list_chat_messages",
				})
		}

		// Format response
		response := make([]ChatMessage, len(byTimestampMessages))
		for i, msg := range byTimestampMessages {
			response[i] = ChatMessage{
				ID:            msg.ID.String(),
				ListID:        msg.ListID.String(),
				UserID:        msg.UserID.String(),
				CharacterName: msg.CharacterName,
				Message:       msg.Message,
				CreatedAt:     msg.CreatedAt.Time,
			}
		}

		return c.JSON(http.StatusOK, response)
	}

	// If no since parameter, get paginated messages
	messages, err := h.store.GetChatMessages(ctx, db.GetChatMessagesParams{
		ListID: listID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to get chat messages", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetChatMessages",
				Table:     "list_chat_messages",
			})
	}

	// Format response
	response := make([]ChatMessage, len(messages))
	for i, msg := range messages {
		response[i] = ChatMessage{
			ID:            msg.ID.String(),
			ListID:        msg.ListID.String(),
			UserID:        msg.UserID.String(),
			CharacterName: msg.CharacterName,
			Message:       msg.Message,
			CreatedAt:     msg.CreatedAt.Time,
		}
	}

	return c.JSON(http.StatusOK, response)
}

// DeleteChatMessage deletes a chat message
func (h *ListsHandler) DeleteChatMessage(c echo.Context) error {
	messageID, err := uuid.Parse(c.Param("messageId"))
	if err != nil {
		return apperror.ValidationError("Invalid message ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "messageId",
				Value:  c.Param("messageId"),
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

	// Delete the message only if the user is the owner
	err = h.store.DeleteChatMessage(ctx, db.DeleteChatMessageParams{
		ID:     messageID,
		UserID: userID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to delete chat message", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "DeleteChatMessage",
				Table:     "list_chat_messages",
			})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetChatNotifications returns chat notifications for the current user
func (h *ListsHandler) GetChatNotifications(c echo.Context) error {
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

	// Get chat notifications for the user
	notifications, err := h.store.GetChatNotificationsForUser(ctx, userID)
	if err != nil {
		return apperror.DatabaseError("Failed to get chat notifications", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetChatNotificationsForUser",
				Table:     "list_chat_messages",
			})
	}

	// Format response
	response := make([]ChatNotification, len(notifications))
	for i, notification := range notifications {
		response[i] = ChatNotification{
			ListID:            notification.ListID.String(),
			ListName:          notification.ListName,
			LastMessageTime:   notification.LastMessageTime.Time,
			UnreadCount:       int32(notification.UnreadCount),
			LastCharacterName: notification.LastCharacterName,
		}
	}

	return c.JSON(http.StatusOK, response)
}

// MarkChatMessagesAsRead marks all chat messages in a list as read for the current user
func (h *ListsHandler) MarkChatMessagesAsRead(c echo.Context) error {
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
	member, err := h.store.IsUserListMember(ctx, db.IsUserListMemberParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to check list membership", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "IsUserListMember",
				Table:     "lists_users",
			})
	}
	if !member {
		return apperror.AuthorizationError("User is not a member of this list", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "list_id",
				Value:  listID.String(),
				Reason: "User is not a member of this list",
			})
	}

	// Mark messages as read
	err = h.store.MarkListMessagesAsRead(ctx, db.MarkListMessagesAsReadParams{
		UserID: userID,
		ListID: listID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to mark messages as read", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "MarkListMessagesAsRead",
				Table:     "list_user_read_status",
			})
	}

	return c.NoContent(http.StatusNoContent)
}
