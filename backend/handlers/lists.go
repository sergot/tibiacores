package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/auth"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
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
		newUser, err := h.store.CreateAnonymousUser(ctx, uuid.New())
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
		char, err := h.store.GetCharacter(ctx, *req.CharacterID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "character not found")
		}
		if char.UserID != userID {
			return echo.NewHTTPError(http.StatusForbidden, "character does not belong to user")
		}

		// Create list using character's world
		list, err := h.store.CreateList(ctx, db.CreateListParams{
			AuthorID: userID,
			Name:     req.Name,
			World:    char.World,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create list")
		}

		// Add character to list
		err = h.store.AddListCharacter(ctx, db.AddListCharacterParams{
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

	// Check if the character name is already taken
	// This single check replaces the two separate checks in the original code
	if req.CharacterName != "" {
		existingChar, err := h.store.GetCharacterByName(ctx, req.CharacterName)
		if err == nil {
			// Character exists
			if existingChar.UserID != userID {
				// Character belongs to another user, return conflict error
				log.Printf("Character %s already exists and belongs to user %s", existingChar.Name, existingChar.UserID)
				return echo.NewHTTPError(http.StatusConflict, "character name is already registered")
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
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create character")
	}

	// Create list
	list, err := h.store.CreateList(ctx, db.CreateListParams{
		AuthorID: userID,
		Name:     strings.TrimSpace(req.Name),
		World:    strings.TrimSpace(req.World),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create list")
	}

	// Add character to list
	err = h.store.AddListCharacter(ctx, db.AddListCharacterParams{
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

	ctx := c.Request().Context()

	// Get list details
	list, err := h.store.GetList(ctx, listID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "list not found")
	}

	// Get member stats
	members, err := h.store.GetListMembers(ctx, listID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get list members")
	}

	isMember := false
	for _, member := range members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "user is not a member of this list")
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
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get soul cores")
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

	ctx := c.Request().Context()

	// Check if user is a member of the list
	isMember, err := h.store.IsUserListMember(ctx, db.IsUserListMemberParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check list membership")
	}

	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "user is not a member of this list")
	}

	// Get the soulcore to check ownership
	soulcore, err := h.store.GetListSoulcore(ctx, db.GetListSoulcoreParams{
		ListID:     listID,
		CreatureID: req.CreatureID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "soulcore not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check soulcore ownership")
	}

	// Check if the user is the one who added the soulcore
	if soulcore.AddedByUserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "only the user who added the soulcore can modify it")
	}

	// Update soul core status
	err = h.store.UpdateSoulcoreStatus(ctx, db.UpdateSoulcoreStatusParams{
		ListID:     listID,
		CreatureID: req.CreatureID,
		Status:     req.Status,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update soul core status")
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

// GetCharacterSuggestions returns all soulcore suggestions for a character
func (h *ListsHandler) GetCharacterSuggestions(c echo.Context) error {
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

	// Verify that the character belongs to the user
	char, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get character")
	}
	if char.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "character does not belong to user")
	}

	suggestions, err := h.store.GetCharacterSuggestions(ctx, characterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get suggestions")
	}

	return c.JSON(http.StatusOK, suggestions)
}

// AcceptSoulcoreSuggestion accepts a soulcore suggestion for a character
func (h *ListsHandler) AcceptSoulcoreSuggestion(c echo.Context) error {
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

	// Verify that the character belongs to the user
	char, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get character")
	}
	if char.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "character does not belong to user")
	}

	// Add the soulcore to the character
	err = h.store.AddCharacterSoulcore(ctx, db.AddCharacterSoulcoreParams{
		CharacterID: characterID,
		CreatureID:  req.CreatureID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add soulcore to character")
	}

	// Remove the suggestion
	err = h.store.DeleteSoulcoreSuggestion(ctx, db.DeleteSoulcoreSuggestionParams{
		CharacterID: characterID,
		CreatureID:  req.CreatureID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete suggestion")
	}

	return c.NoContent(http.StatusOK)
}

// DismissSoulcoreSuggestion dismisses a soulcore suggestion without adding it to the character
func (h *ListsHandler) DismissSoulcoreSuggestion(c echo.Context) error {
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

	// Verify that the character belongs to the user
	char, err := h.store.GetCharacter(ctx, characterID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get character")
	}
	if char.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "character does not belong to user")
	}

	// Remove the suggestion
	err = h.store.DeleteSoulcoreSuggestion(ctx, db.DeleteSoulcoreSuggestionParams{
		CharacterID: characterID,
		CreatureID:  req.CreatureID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete suggestion")
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

	ctx := c.Request().Context()

	// Check if user is a member of the list
	isMember, err := h.store.IsUserListMember(ctx, db.IsUserListMemberParams{
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
	err = h.store.AddSoulcoreToList(ctx, db.AddSoulcoreToListParams{
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

// RemoveSoulcore removes a soul core from a list
func (h *ListsHandler) RemoveSoulcore(c echo.Context) error {
	listID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid list ID")
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

	// Check if user is a member of the list
	isMember, err := h.store.IsUserListMember(ctx, db.IsUserListMemberParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check list membership")
	}

	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "user is not a member of this list")
	}

	// Get the soulcore to check ownership
	soulcore, err := h.store.GetListSoulcore(ctx, db.GetListSoulcoreParams{
		ListID:     listID,
		CreatureID: creatureID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "soulcore not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to check soulcore ownership")
	}

	// Check if the user is the one who added the soulcore
	if soulcore.AddedByUserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "only the user who added the soulcore can remove it")
	}

	// Delete the soulcore from the list
	err = h.store.RemoveListSoulcore(ctx, db.RemoveListSoulcoreParams{
		ListID:     listID,
		CreatureID: creatureID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove soul core")
	}

	return c.NoContent(http.StatusOK)
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
		return echo.NewHTTPError(http.StatusBadRequest, "invalid share code")
	}

	var req JoinListRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	ctx := c.Request().Context()

	// Get the list by share code
	list, err := h.store.GetListByShareCode(ctx, shareCode)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "list not found")
	}

	// Check if user is authenticated
	var userID uuid.UUID
	var token string

	// Get authenticated user ID from context
	if userIDStr, ok := c.Get("user_id").(string); ok && userIDStr != "" {
		// User is authenticated, parse their ID
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID format")
		}
		log.Printf("Using existing authenticated user: %s", userID)
	} else {
		// Create new anonymous user account
		log.Printf("No authenticated user found, creating anonymous user")
		newUser, err := h.store.CreateAnonymousUser(ctx, uuid.New())
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

		// For new users joining with a new character, require character info
		if req.CharacterID == "" && (req.CharacterName == "" || req.World == "") {
			return echo.NewHTTPError(http.StatusBadRequest, "character_name and world are required for first join")
		}
	}

	// Check if user is already a member
	isMember, err := h.store.IsUserListMember(ctx, db.IsUserListMemberParams{
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
	if req.CharacterID != "" {
		// Parse character ID from string to UUID
		characterID, err := uuid.Parse(req.CharacterID)
		if err != nil {
			log.Printf("Failed to parse character ID %q: %v", req.CharacterID, err)
			return echo.NewHTTPError(http.StatusBadRequest, "invalid character ID format")
		}
		log.Printf("Parsed character ID: %s", characterID)

		// Verify character exists and belongs to user
		character, err = h.store.GetCharacter(ctx, characterID)
		if err != nil {
			log.Printf("Failed to find character with ID %s: %v", characterID, err)
			return echo.NewHTTPError(http.StatusNotFound, "character not found")
		}
		log.Printf("Found character: ID=%s, UserID=%s, Name=%s, World=%s", character.ID, character.UserID, character.Name, character.World)
		log.Printf("Current user ID: %s", userID)

		if character.UserID != userID {
			log.Printf("Character %s belongs to user %s, but current user is %s", character.ID, character.UserID, userID)
			return echo.NewHTTPError(http.StatusForbidden, "character does not belong to user")
		}
		if character.World != list.World {
			log.Printf("Character world %s does not match list world %s", character.World, list.World)
			return echo.NewHTTPError(http.StatusBadRequest, "character world does not match list world")
		}
	} else {
		// Check if the character name is already taken
		if req.CharacterName != "" {
			existingChar, err := h.store.GetCharacterByName(ctx, req.CharacterName)
			if err == nil && existingChar.UserID != userID {
				return echo.NewHTTPError(http.StatusConflict, "character name is already registered")
			}
		}

		// Create new character
		character, err = h.store.CreateCharacter(ctx, db.CreateCharacterParams{
			UserID: userID,
			Name:   strings.TrimSpace(req.CharacterName),
			World:  strings.TrimSpace(req.World),
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create character")
		}
	}

	// Add character to list
	err = h.store.AddListCharacter(ctx, db.AddListCharacterParams{
		ListID:      list.ID,
		UserID:      userID,
		CharacterID: character.ID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add character to list")
	}

	// Get member stats for the newly added member
	members, err := h.store.GetListMembers(ctx, list.ID)
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

	ctx := c.Request().Context()

	// Get the list by share code
	list, err := h.store.GetListByShareCode(ctx, shareCode)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "list not found")
	}

	// Get member count
	members, err := h.store.GetMembers(ctx, list.ID)
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
