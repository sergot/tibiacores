package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/auth"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
)

type ListsHandler struct {
	store db.Store
}

func NewListsHandler(store db.Store) *ListsHandler {
	return &ListsHandler{store}
}

type CreateListRequest struct {
	CharacterID   *uuid.UUID `json:"character_id,omitempty"`
	CharacterName string     `json:"character_name,omitempty"`
	Name          string     `json:"name"`
	World         string     `json:"world,omitempty"`
}

type CreateListResponse struct {
	ID          uuid.UUID `json:"id"`
	AuthorID    uuid.UUID `json:"author_id"`
	Name        string    `json:"name"`
	ShareCode   uuid.UUID `json:"share_code"`
	World       string    `json:"world"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsAnonymous bool      `json:"is_anonymous"`
	HasEmail    bool      `json:"has_email"`
}

func (h *ListsHandler) CreateList(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateListRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return apperror.ValidationError("Invalid request body", err).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "request_body",
			Reason: "Failed to decode JSON request",
		})
	}

	if req.Name == "" {
		return apperror.ValidationError("Name is required", nil).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "name",
			Reason: "Name field cannot be empty",
		})
	}

	// Check if user is authenticated
	var userID uuid.UUID
	var err error

	// Get authenticated user ID from context
	if userIDStr, ok := c.Get("user_id").(string); ok && userIDStr != "" {
		// User is authenticated, parse their ID
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			return apperror.AuthorizationError("invalid user ID format", err)
		}
	} else {
		// Create new anonymous user account
		newUser, err := h.store.CreateAnonymousUser(ctx, uuid.New())
		if err != nil {
			return apperror.DatabaseError("failed to create user", err)
		}
		userID = newUser.ID

		// Generate token pair for cookies
		tokenPair, err := auth.GenerateTokenPair(userID.String(), false)
		if err != nil {
			return apperror.InternalError("failed to generate token pair", err).WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GenerateTokenPair",
				Table:     "auth",
			})
		}

		// Set cookies for authentication
		auth.SetTokenCookies(c, tokenPair.AccessToken, tokenPair.RefreshToken, tokenPair.ExpiresIn)

		// For new users, we need character info
		if req.CharacterName == "" || req.World == "" {
			return apperror.ValidationError("Character name and world are required for first list", nil).WithDetails(&apperror.ValidationErrorDetails{
				Field:  "character_name,world",
				Reason: "Both character name and world are required for new users",
			})
		}
	}

	// Handle existing character case
	if req.CharacterID != nil {
		// Verify character exists and belongs to user
		char, err := h.store.GetCharacter(ctx, *req.CharacterID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperror.NotFoundError("Character not found", err).WithDetails(&apperror.ValidationErrorDetails{
					Field:  "character_id",
					Value:  req.CharacterID.String(),
					Reason: "Character does not exist",
				})
			}
			return apperror.DatabaseError("Failed to retrieve character", err).WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetCharacter",
				Table:     "characters",
			})
		}
		if char.UserID != userID {
			return apperror.AuthorizationError("Character does not belong to user", nil).WithDetails(&apperror.ValidationErrorDetails{
				Field:  "character_id",
				Value:  req.CharacterID.String(),
				Reason: "Character belongs to a different user",
			})
		}

		// Create list using character's world
		list, err := h.store.CreateList(ctx, db.CreateListParams{
			AuthorID: userID,
			Name:     req.Name,
			World:    char.World,
		})
		if err != nil {
			return apperror.DatabaseError("Failed to create list", err).WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "CreateList",
				Table:     "lists",
			})
		}

		// Add character to list
		err = h.store.AddListCharacter(ctx, db.AddListCharacterParams{
			ListID:      list.ID,
			UserID:      userID,
			CharacterID: *req.CharacterID,
		})
		if err != nil {
			return apperror.DatabaseError("Failed to add character to list", err).WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "AddListCharacter",
				Table:     "list_characters",
			})
		}

		// Safely get has_email value
		hasEmail := false
		if hasEmailVal := c.Get("has_email"); hasEmailVal != nil {
			if val, ok := hasEmailVal.(bool); ok {
				hasEmail = val
			}
		}

		return c.JSON(http.StatusCreated, CreateListResponse{
			ID:        list.ID,
			AuthorID:  list.AuthorID,
			Name:      list.Name,
			ShareCode: list.ShareCode,
			World:     list.World,
			CreatedAt: list.CreatedAt.Time,
			UpdatedAt: list.UpdatedAt.Time,
			HasEmail:  hasEmail,
		})
	}

	// Handle new character case
	if req.CharacterName == "" || req.World == "" {
		return apperror.ValidationError("Character name and world are required for new character", nil).WithDetails(&apperror.ValidationErrorDetails{
			Field:  "character_name,world",
			Reason: "Both character name and world are required for new characters",
		})
	}

	// Check if the character name is already taken
	// This single check replaces the two separate checks in the original code
	if req.CharacterName != "" {
		existingChar, err := h.store.GetCharacterByName(ctx, req.CharacterName)
		if err == nil {
			// Character exists
			if existingChar.UserID != userID {
				// Character belongs to another user, return conflict error
				return apperror.ValidationError("Character name is already registered", nil).WithDetails(&apperror.ValidationErrorDetails{
					Field:  "character_name",
					Value:  req.CharacterName,
					Reason: "Character name is already taken by another user",
				})
			}
			// Character belongs to current user, we could use it but for simplicity
			// we'll create a new one as in the original flow
		}
	}

	// Create character
	character, err := h.store.CreateCharacter(ctx, db.CreateCharacterParams{
		UserID: userID,
		Name:   strings.TrimSpace(req.CharacterName),
		World:  strings.TrimSpace(req.World),
	})
	if err != nil {
		return apperror.DatabaseError("Failed to create character", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "CreateCharacter",
			Table:     "characters",
		})
	}

	// Create list
	list, err := h.store.CreateList(ctx, db.CreateListParams{
		AuthorID: userID,
		Name:     strings.TrimSpace(req.Name),
		World:    strings.TrimSpace(req.World),
	})
	if err != nil {
		return apperror.DatabaseError("Failed to create list", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "CreateList",
			Table:     "lists",
		})
	}

	// Add character to list
	err = h.store.AddListCharacter(ctx, db.AddListCharacterParams{
		ListID:      list.ID,
		UserID:      userID,
		CharacterID: character.ID,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to add character to list", err).WithDetails(&apperror.DatabaseErrorDetails{
			Operation: "AddListCharacter",
			Table:     "list_characters",
		})
	}

	// Safely get has_email value
	hasEmail := false
	if hasEmailVal := c.Get("has_email"); hasEmailVal != nil {
		if val, ok := hasEmailVal.(bool); ok {
			hasEmail = val
		}
	}

	return c.JSON(http.StatusCreated, CreateListResponse{
		ID:        list.ID,
		AuthorID:  list.AuthorID,
		Name:      list.Name,
		ShareCode: list.ShareCode,
		World:     list.World,
		CreatedAt: list.CreatedAt.Time,
		UpdatedAt: list.UpdatedAt.Time,
		HasEmail:  hasEmail,
	})
}

type ListDetailResponse struct {
	ID        uuid.UUID                `json:"id"`
	AuthorID  uuid.UUID                `json:"author_id"`
	Name      string                   `json:"name"`
	ShareCode uuid.UUID                `json:"share_code"`
	World     string                   `json:"world"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
	Members   []MemberStats            `json:"members"`
	SoulCores []db.GetListSoulcoresRow `json:"soul_cores"`
}

type MemberStats struct {
	UserID        uuid.UUID `json:"user_id"`
	CharacterName string    `json:"character_name"`
	ObtainedCount int64     `json:"obtained_count"`
	UnlockedCount int64     `json:"unlocked_count"`
	IsActive      bool      `json:"is_active"`
}

// JoinListRequest represents the request body for joining a list
type JoinListRequest struct {
	CharacterName string `json:"character_name,omitempty"`
	World         string `json:"world,omitempty"`
	CharacterID   string `json:"character_id,omitempty"`
}

// JoinList allows a user to join a list using its share code
func (h *ListsHandler) JoinList(c echo.Context) error {
	shareCode, err := uuid.Parse(c.Param("share_code"))
	if err != nil {
		return apperror.ValidationError("invalid share code", err)
	}

	var req JoinListRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return apperror.ValidationError("invalid request body", err)
	}

	ctx := c.Request().Context()

	// Get the list by share code
	list, err := h.store.GetListByShareCode(ctx, shareCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("list not found", err)
		}
		return apperror.DatabaseError("failed to retrieve list", err)
	}

	// Check if user is authenticated
	var userID uuid.UUID

	// Get authenticated user ID from context
	if userIDStr, ok := c.Get("user_id").(string); ok && userIDStr != "" {
		// User is authenticated, parse their ID
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			return apperror.AuthorizationError("invalid user ID format", err)
		}
	} else {
		// Create new anonymous user account
		newUser, err := h.store.CreateAnonymousUser(ctx, uuid.New())
		if err != nil {
			return apperror.DatabaseError("failed to create user", err)
		}
		userID = newUser.ID

		// Generate token pair for cookies
		tokenPair, err := auth.GenerateTokenPair(userID.String(), false)
		if err != nil {
			return apperror.InternalError("failed to generate token pair", err).WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GenerateTokenPair",
				Table:     "auth",
			})
		}

		// Set cookies for authentication
		auth.SetTokenCookies(c, tokenPair.AccessToken, tokenPair.RefreshToken, tokenPair.ExpiresIn)

		// For new users joining with a new character, require character info
		if req.CharacterID == "" && (req.CharacterName == "" || req.World == "") {
			return apperror.ValidationError("character_name and world are required for first join", nil)
		}
	}

	// Check if user is already a member
	isMember, err := h.store.IsUserListMember(ctx, db.IsUserListMemberParams{
		ListID: list.ID,
		UserID: userID,
	})
	if err != nil {
		return apperror.DatabaseError("failed to check list membership", err)
	}
	if isMember {
		return apperror.ValidationError("user is already a member of this list", nil)
	}

	var character db.Character
	if req.CharacterID != "" {
		// Parse character ID from string to UUID
		characterID, err := uuid.Parse(req.CharacterID)
		if err != nil {
			return apperror.ValidationError("invalid character ID format", err)
		}

		// Verify character exists and belongs to user
		character, err = h.store.GetCharacter(ctx, characterID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperror.NotFoundError("character not found", err)
			}
			return apperror.DatabaseError("failed to retrieve character", err)
		}

		if character.UserID != userID {
			return apperror.AuthorizationError("character does not belong to user", nil)
		}

		if character.World != list.World {
			return apperror.ValidationError("character world does not match list world", nil)
		}
	} else {
		// Check if the character name is already taken
		if req.CharacterName != "" {
			existingChar, err := h.store.GetCharacterByName(ctx, req.CharacterName)
			if err == nil && existingChar.UserID != userID {
				return apperror.ValidationError("character name is already registered", nil)
			}
		}

		// Create new character
		character, err = h.store.CreateCharacter(ctx, db.CreateCharacterParams{
			UserID: userID,
			Name:   strings.TrimSpace(req.CharacterName),
			World:  strings.TrimSpace(req.World),
		})
		if err != nil {
			return apperror.DatabaseError("failed to create character", err)
		}
	}

	// Add character to list
	err = h.store.AddListCharacter(ctx, db.AddListCharacterParams{
		ListID:      list.ID,
		UserID:      userID,
		CharacterID: character.ID,
	})
	if err != nil {
		return apperror.DatabaseError("failed to add character to list", err)
	}

	// Get member stats for the newly added member
	members, err := h.store.GetListMembers(ctx, list.ID)
	if err != nil {
		return apperror.DatabaseError("failed to get list members", err)
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

	// Return list details
	return c.JSON(http.StatusOK, ListDetailResponse{
		ID:        list.ID,
		AuthorID:  list.AuthorID,
		Name:      list.Name,
		ShareCode: list.ShareCode,
		World:     list.World,
		CreatedAt: list.CreatedAt.Time,
		UpdatedAt: list.UpdatedAt.Time,
		Members:   memberStats,
		SoulCores: []db.GetListSoulcoresRow{},
	})
}
