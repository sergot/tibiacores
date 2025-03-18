package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/sergot/fiendlist/backend/db"
)

type ListsHandler struct {
	connPool *pgxpool.Pool
}

func NewListsHandler(connPool *pgxpool.Pool) *ListsHandler {
	return &ListsHandler{connPool}
}

type CreateListRequest struct {
	SessionToken  *string    `json:"session_token,omitempty"`
	CharacterID   *uuid.UUID `json:"character_id,omitempty"`
	CharacterName string     `json:"character_name,omitempty"`
	Name          string     `json:"name"`
	World         string     `json:"world,omitempty"`
	UserID        *uuid.UUID `json:"user_id,omitempty"`
}

func (h *ListsHandler) CreateList(c echo.Context) error {
	queries := db.New(h.connPool)

	var req CreateListRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate required fields
	// TODO: assign default name
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	ctx := c.Request().Context()

	// Case 1: First-time user with session token
	if req.SessionToken != nil {
		if req.CharacterName == "" || req.World == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "character_name and world are required for new users")
		}

		sessionToken, err := uuid.Parse(*req.SessionToken)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid session token")
		}

		user, err := queries.CreateAnonymousUser(ctx, sessionToken)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
		}

		// Create list
		list, err := queries.CreateList(ctx, db.CreateListParams{
			AuthorID: user.ID,
			Name:     req.Name,
			World:    req.World,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create list")
		}

		// Create character
		character, err := queries.CreateCharacter(ctx, db.CreateCharacterParams{
			UserID: user.ID,
			Name:   req.CharacterName,
			World:  req.World,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create character")
		}

		// Add character to list
		err = queries.AddListCharacter(ctx, db.AddListCharacterParams{
			ListID:      list.ID,
			UserID:      user.ID,
			CharacterID: character.ID,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to add character to list")
		}

		return c.JSON(http.StatusCreated, list)
	}

	// Get user ID either from request or auth context
	var userID uuid.UUID
	if req.UserID != nil {
		userID = *req.UserID
	} else {
		id, ok := c.Get("user_id").(uuid.UUID)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid user authentication")
		}
		userID = id
	}

	// Case 2a: Existing character
	if req.CharacterID != nil {
		// Verify character belongs to user
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

		return c.JSON(http.StatusCreated, list)
	}

	// Case 2b: New character
	if req.CharacterName == "" || req.World == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "character_name and world are required for new character")
	}

	// Create character
	character, err := queries.CreateCharacter(ctx, db.CreateCharacterParams{
		UserID: userID,
		Name:   req.CharacterName,
		World:  req.World,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create character")
	}

	// Create list
	list, err := queries.CreateList(ctx, db.CreateListParams{
		AuthorID: userID,
		Name:     req.Name,
		World:    req.World,
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

	return c.JSON(http.StatusCreated, list)
}

// GetUserLists returns all lists where the user is either an author or a member
func (h *ListsHandler) GetUserLists(c echo.Context) error {
	queries := db.New(h.connPool)

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user authentication")
	}

	ctx := c.Request().Context()
	lists, err := queries.GetUserLists(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get lists")
	}

	return c.JSON(http.StatusOK, lists)
}
