package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	mockdb "github.com/sergot/tibiacores/backend/db/mock"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/handlers"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreaturesGet(t *testing.T) {
	e := echo.New()

	creatures := []db.Creature{
		{
			ID:   uuid.New(),
			Name: "demon",
		},
		{
			ID:   uuid.New(),
			Name: "dragon",
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	store.
		EXPECT().
		GetCreatures(gomock.Any()).Return(creatures, nil).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/api/creatures", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/creatures")

	h := handlers.NewCreaturesHandler(store)

	require.NoError(t, h.GetCreatures(c))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "demon")
}
