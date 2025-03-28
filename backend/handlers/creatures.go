package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
)

type CreaturesHandler struct {
	store db.Store
}

func NewCreaturesHandler(store db.Store) *CreaturesHandler {
	return &CreaturesHandler{store: store}
}

func (h *CreaturesHandler) GetCreatures(c echo.Context) error {
	creatures, err := h.store.GetCreatures(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, creatures)
}
