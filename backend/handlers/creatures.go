package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
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
		return apperror.DatabaseError("Failed to retrieve creatures", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetCreatures",
				Table:     "creatures",
			}).
			Wrap(err)
	}

	return c.JSON(http.StatusOK, creatures)
}
