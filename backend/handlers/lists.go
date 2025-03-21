package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/sergot/fiendlist/backend/auth"
	"github.com/sergot/fiendlist/backend/db"
)

type ListsHandler struct {
	connPool *pgxpool.Pool
}

func NewListsHandler(connPool *pgxpool.Pool) *ListsHandler {
	return &ListsHandler{connPool}
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
	queries := db.New(h.connPool)
	ctx := c.Request().Context()

	var req CreateListRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	// Check if user is authenticated by looking for user_id in context
	userIDStr := c.Get("user_id")
	var userID uuid.UUID
	var token string

	if userIDStr == nil {
		// Create new user account
		newUser, err := queries.CreateAnonymousUser(ctx, uuid.New())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
		}
		userID = newUser.ID

		// Generate token
		token, err = auth.GenerateToken(userID.String(), false)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
		}
		c.Response().Header().Set("X-Auth-Token", token)

		// For new users, we need character info
		if req.CharacterName == "" || req.World == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "character_name and world are required for first list")
		}
	} else {
		// Use existing user
		var ok bool
		userIDStr, ok := userIDStr.(string)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid user authentication")
		}
		var err error
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID format")
		}
	}

	// Handle existing character case
	if req.CharacterID != nil {
		// Verify character exists and belongs to user
		char, err := queries.GetCharacter(ctx, *req.CharacterID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "character not found")
		}
		if char.UserID != userID {
			return echo.NewHTTPError(http.StatusForbidden, "character does not belong to user")
		}

		// Create list using character's world
		list, err := queries.CreateList(ctx, db.CreateListParams{
			AuthorID: userID,
			Name:     req.Name,
			World:    char.World,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create list")
		}

		// Add character to list
		err = queries.AddListCharacter(ctx, db.AddListCharacterParams{
			ListID:      list.ID,
			UserID:      userID,
			CharacterID: *req.CharacterID,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to add character to list")
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
		return echo.NewHTTPError(http.StatusBadRequest, "character_name and world are required for new character")
	}

	// Create character
	character, err := queries.CreateCharacter(ctx, db.CreateCharacterParams{
		UserID: userID,
		Name:   strings.TrimSpace(req.CharacterName),
		World:  strings.TrimSpace(req.World),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create character")
	}

	// Create list
	list, err := queries.CreateList(ctx, db.CreateListParams{
		AuthorID: userID,
		Name:     strings.TrimSpace(req.Name),
		World:    strings.TrimSpace(req.World),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create list")
	}

	// Add character to list
	err = queries.AddListCharacter(ctx, db.AddListCharacterParams{
		ListID:      list.ID,
		UserID:      userID,
		CharacterID: character.ID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add character to list")
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
	ID         uuid.UUID                `json:"id"`
	AuthorID   uuid.UUID                `json:"author_id"`
	Name       string                   `json:"name"`
	ShareCode  uuid.UUID                `json:"share_code"`
	World      string                   `json:"world"`
	CreatedAt  time.Time                `json:"created_at"`
	UpdatedAt  time.Time                `json:"updated_at"`
	Members    []MemberStats            `json:"members"`
	SoulCores  []db.GetListSoulcoresRow `json:"soul_cores"`
	TotalCores int                      `json:"total_cores"`
}

type MemberStats struct {
	UserID        uuid.UUID `json:"user_id"`
	CharacterName string    `json:"character_name"`
	ObtainedCount int64     `json:"obtained_count"`
	UnlockedCount int64     `json:"unlocked_count"`
}

// GetList returns detailed information about a specific list
func (h *ListsHandler) GetList(c echo.Context) error {
	listID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid list ID")
	}

	// Get authenticated user ID from context
	userIDStr := c.Get("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID format")
	}

	queries := db.New(h.connPool)
	ctx := c.Request().Context()

	// Check if user is a member of the list
	isMember, err := queries.IsUserListMember(ctx, db.IsUserListMemberParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check list membership")
	}

	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "user is not a member of this list")
	}

	// Get list details
	list, err := queries.GetList(ctx, listID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "list not found")
	}

	// Get member stats
	members, err := queries.GetListMembers(ctx, listID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get list members")
	}

	memberStats := make([]MemberStats, len(members))
	for i, m := range members {
		memberStats[i] = MemberStats{
			UserID:        m.UserID,
			CharacterName: m.CharacterName,
			ObtainedCount: m.ObtainedCount,
			UnlockedCount: m.UnlockedCount,
		}
	}

	// Get soul cores
	soulCores, err := queries.GetListSoulcores(ctx, listID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get soul cores")
	}

	// Get total number of creatures
	creatures, err := queries.GetCreatures(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get creatures")
	}

	return c.JSON(http.StatusOK, ListDetailResponse{
		ID:         list.ID,
		AuthorID:   list.AuthorID,
		Name:       list.Name,
		ShareCode:  list.ShareCode,
		World:      list.World,
		CreatedAt:  list.CreatedAt.Time,
		UpdatedAt:  list.UpdatedAt.Time,
		Members:    memberStats,
		SoulCores:  soulCores,
		TotalCores: len(creatures),
	})
}

// UpdateSoulcoreStatus updates the status of a soul core in a list
func (h *ListsHandler) UpdateSoulcoreStatus(c echo.Context) error {
	listID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid list ID")
	}

	var req struct {
		CreatureID uuid.UUID         `json:"creature_id"`
		Status     db.SoulcoreStatus `json:"status"`
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

	queries := db.New(h.connPool)
	ctx := c.Request().Context()

	// Check if user is a member of the list
	isMember, err := queries.IsUserListMember(ctx, db.IsUserListMemberParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check list membership")
	}

	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "user is not a member of this list")
	}

	// Update soul core status - the added_by_user_id remains the same as the original creator
	// while the status can be modified by any member
	err = queries.UpdateSoulcoreStatus(ctx, db.UpdateSoulcoreStatusParams{
		ListID:     listID,
		CreatureID: req.CreatureID,
		Status:     req.Status,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update soul core status")
	}

	return c.NoContent(http.StatusOK)
}

// AddSoulcore adds a new soul core to a list
func (h *ListsHandler) AddSoulcore(c echo.Context) error {
	listID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid list ID")
	}

	var req struct {
		CreatureID uuid.UUID         `json:"creature_id"`
		Status     db.SoulcoreStatus `json:"status"`
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

	queries := db.New(h.connPool)
	ctx := c.Request().Context()

	// Check if user is a member of the list
	isMember, err := queries.IsUserListMember(ctx, db.IsUserListMemberParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check list membership")
	}

	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "user is not a member of this list")
	}

	// Add soul core with the user ID who added it
	err = queries.AddSoulcoreToList(ctx, db.AddSoulcoreToListParams{
		ListID:        listID,
		CreatureID:    req.CreatureID,
		Status:        req.Status,
		AddedByUserID: userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add soul core")
	}

	return c.NoContent(http.StatusOK)
}

// JoinListRequest represents the request body for joining a list
type JoinListRequest struct {
	CharacterName string     `json:"character_name,omitempty"`
	World         string     `json:"world,omitempty"`
	CharacterID   *uuid.UUID `json:"character_id,omitempty"`
}

// JoinList allows a user to join a list using its share code
func (h *ListsHandler) JoinList(c echo.Context) error {
	shareCode, err := uuid.Parse(c.Param("share_code"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid share code")
	}

	var req JoinListRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	queries := db.New(h.connPool)
	ctx := c.Request().Context()

	// Get the list by share code
	list, err := queries.GetListByShareCode(ctx, shareCode)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "list not found")
	}

	// Check if user is authenticated
	userIDStr := c.Get("user_id")
	var userID uuid.UUID
	var token string

	if userIDStr == nil {
		// Create new anonymous user account
		newUser, err := queries.CreateAnonymousUser(ctx, uuid.New())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
		}
		userID = newUser.ID

		// Generate token
		token, err = auth.GenerateToken(userID.String(), false)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
		}
		c.Response().Header().Set("X-Auth-Token", token)

		// For new users, we need character info
		if req.CharacterName == "" || req.World == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "character_name and world are required for first join")
		}
	} else {
		// Use existing user
		var ok bool
		userIDStr, ok := userIDStr.(string)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid user authentication")
		}
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID format")
		}
	}

	// Check if user is already a member
	isMember, err := queries.IsUserListMember(ctx, db.IsUserListMemberParams{
		ListID: list.ID,
		UserID: userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check list membership")
	}
	if isMember {
		return echo.NewHTTPError(http.StatusBadRequest, "user is already a member of this list")
	}

	var character db.Character
	if req.CharacterID != nil {
		// Verify character exists and belongs to user
		character, err = queries.GetCharacter(ctx, *req.CharacterID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "character not found")
		}
		if character.UserID != userID {
			return echo.NewHTTPError(http.StatusForbidden, "character does not belong to user")
		}
		if character.World != list.World {
			return echo.NewHTTPError(http.StatusBadRequest, "character world does not match list world")
		}
	} else {
		// Create new character
		character, err = queries.CreateCharacter(ctx, db.CreateCharacterParams{
			UserID: userID,
			Name:   strings.TrimSpace(req.CharacterName),
			World:  strings.TrimSpace(req.World),
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create character")
		}
	}

	// Add character to list
	err = queries.AddListCharacter(ctx, db.AddListCharacterParams{
		ListID:      list.ID,
		UserID:      userID,
		CharacterID: character.ID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add character to list")
	}

	// Return list details
	return c.JSON(http.StatusOK, ListDetailResponse{
		ID:         list.ID,
		AuthorID:   list.AuthorID,
		Name:       list.Name,
		ShareCode:  list.ShareCode,
		World:      list.World,
		CreatedAt:  list.CreatedAt.Time,
		UpdatedAt:  list.UpdatedAt.Time,
		Members:    []MemberStats{},
		SoulCores:  []db.GetListSoulcoresRow{},
		TotalCores: 0,
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
		return echo.NewHTTPError(http.StatusBadRequest, "invalid share code")
	}

	queries := db.New(h.connPool)
	ctx := c.Request().Context()

	// Get the list by share code
	list, err := queries.GetListByShareCode(ctx, shareCode)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "list not found")
	}

	// Get member count
	members, err := queries.GetMembers(ctx, list.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get list members")
	}

	return c.JSON(http.StatusOK, ListPreviewResponse{
		ID:          list.ID,
		Name:        list.Name,
		World:       list.World,
		MemberCount: len(members),
	})
}
