package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/sergot/fiendlist/backend/db"
)

type CreaturesHandler struct {
	connPool *pgxpool.Pool
}

func NewCreaturesHandler(connPool *pgxpool.Pool) *CreaturesHandler {
	return &CreaturesHandler{connPool}
}

func (h *CreaturesHandler) GetCreatures(c echo.Context) error {
	queries := db.New(h.connPool)
	creatures, err := queries.GetCreatures(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(200, creatures)
}
