package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sergot/fiendlist/backend/db"
)

type CreaturesHandler struct {
	queries *db.Queries
}

func NewCreaturesHandler(queries *db.Queries) *CreaturesHandler {
	return &CreaturesHandler{queries}
}

func (h *CreaturesHandler) GetCreatures(c echo.Context) error {
	creatures, err := h.queries.GetCreatures(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(200, creatures)
}
