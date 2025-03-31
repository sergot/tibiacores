package handlers_test

import (
	"encoding/json"
	"errors"
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

func TestGetCreatures(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		setupMocks     func(store *mockdb.MockStore)
		expectedStatus int
		validateResp   func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Success",
			setupMocks: func(store *mockdb.MockStore) {
				creatures := []db.Creature{
					{
						ID:   uuid.New(),
						Name: "Demon",
					},
					{
						ID:   uuid.New(),
						Name: "Dragon",
					},
				}
				store.EXPECT().
					GetCreatures(gomock.Any()).
					Return(creatures, nil)
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				// Parse the response
				var creatures []db.Creature
				err := json.Unmarshal(recorder.Body.Bytes(), &creatures)
				require.NoError(t, err)

				// Validate response content
				require.Len(t, creatures, 2)
				require.Equal(t, "Demon", creatures[0].Name)
				require.Equal(t, "Dragon", creatures[1].Name)
				require.NotEmpty(t, creatures[0].ID)
				require.NotEmpty(t, creatures[1].ID)
			},
		},
		{
			name: "Database Error",
			setupMocks: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCreatures(gomock.Any()).
					Return(nil, errors.New("database connection error"))
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)

				// Validate error response
				var errorResp map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
				require.NoError(t, err)
				require.Contains(t, errorResp["error"], "database connection error")
			},
		},
		{
			name: "Empty Creatures List",
			setupMocks: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCreatures(gomock.Any()).
					Return([]db.Creature{}, nil)
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				// Parse the response
				var creatures []db.Creature
				err := json.Unmarshal(recorder.Body.Bytes(), &creatures)
				require.NoError(t, err)

				// Validate empty array
				require.Len(t, creatures, 0)
				require.Equal(t, "[]\n", recorder.Body.String())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.setupMocks(store)

			// Create request and recorder
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/creatures", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/api/creatures")

			// Execute the handler
			h := handlers.NewCreaturesHandler(store)
			err := h.GetCreatures(c)

			// Validate results
			require.NoError(t, err)
			tc.validateResp(t, rec)
		})
	}
}
