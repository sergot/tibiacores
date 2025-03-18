package handlers

import (
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/sergot/fiendlist/backend/db"
)

type UsersHandler struct {
	connPool *pgxpool.Pool
}

func NewUsersHandler(connPool *pgxpool.Pool) *UsersHandler {
	return &UsersHandler{connPool}
}

func (h *UsersHandler) GetCharactersByUserId(c echo.Context) error {
	queries := db.New(h.connPool)

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(400, "invalid user ID")
	}

	characters, err := queries.GetCharactersByUserID(c.Request().Context(), userID)
	if err != nil {
		// log the error for debugging
		log.Printf("Error getting characters for user %s: %v", userID, err)
		return echo.NewHTTPError(500, "failed to get characters")
	}

	return c.JSON(200, characters)
}
