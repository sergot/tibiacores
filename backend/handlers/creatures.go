package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/errors"
)

type CreaturesHandler struct {
	store db.Store
}

func NewCreaturesHandler(store db.Store) *CreaturesHandler {
	return &CreaturesHandler{store: store}
}

func (h *CreaturesHandler) GetCreatures(c echo.Context) error {
	ctx := c.Request().Context()

	creatures, err := h.store.GetCreatures(ctx)
	if err != nil {
		return errors.NewCreatureUnavailableError(err).
			WithOperation("get_creatures").
			WithResource("creature")
	}

	if len(creatures) == 0 {
		return errors.NewCreatureNotFoundError(errors.ErrCreatureNotFound).
			WithOperation("get_creatures").
			WithResource("creature")
	}

	return c.JSON(http.StatusOK, creatures)
}
