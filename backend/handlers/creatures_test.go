package handlers_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	mockdb "github.com/sergot/tibiacores/backend/db/mock"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/handlers"
	"github.com/sergot/tibiacores/backend/middleware"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetCreatures(t *testing.T) {
	testCases := []struct {
		name          string
		setupMocks    func(store *mockdb.MockStore)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, creatures []db.Creature, rec *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Single Creature",
			setupMocks: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCreatures(gomock.Any()).
					Return([]db.Creature{
						{
							ID:   uuid.New(),
							Name: "Dragon",
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, creatures []db.Creature, rec *httptest.ResponseRecorder) {
				require.Len(t, creatures, 1)
				require.Equal(t, "Dragon", creatures[0].Name)
			},
		},
		{
			name: "Success - Multiple Creatures",
			setupMocks: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCreatures(gomock.Any()).
					Return([]db.Creature{
						{
							ID:         uuid.New(),
							Name:       "Dragon",
							Difficulty: pgtype.Int4{Int32: 2},
						},
						{
							ID:         uuid.New(),
							Name:       "Demon",
							Difficulty: pgtype.Int4{Int32: 4},
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, creatures []db.Creature, rec *httptest.ResponseRecorder) {
				require.Len(t, creatures, 2)
				require.Equal(t, "Dragon", creatures[0].Name)
				require.Equal(t, 2, creatures[0].Difficulty.Int32)
				require.Equal(t, "Demon", creatures[1].Name)
				require.Equal(t, 4, creatures[1].Difficulty.Int32)
			},
		},
		{
			name: "Empty Creatures List",
			setupMocks: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCreatures(gomock.Any()).
					Return([]db.Creature{}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, creatures []db.Creature, rec *httptest.ResponseRecorder) {
				require.Len(t, creatures, 0)
				require.Equal(t, "[]\n", rec.Body.String())
			},
		},
		{
			name: "Database Error",
			setupMocks: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetCreatures(gomock.Any()).
					Return(nil, sql.ErrConnDone)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to retrieve creatures",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.setupMocks(store)

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/api/creatures", nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Create handler with mock store
			h := handlers.NewCreaturesHandler(store)

			// Execute handler
			err := h.GetCreatures(c)

			// Check for expected error response
			if tc.expectedError != "" {
				// Use the ErrorHandler to process the error
				middleware.ErrorHandler(err, c)

				// Check if we received an error with the correct status code and message
				require.Equal(t, tc.expectedCode, rec.Code)

				var errorResponse map[string]interface{}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errorResponse))
				require.Contains(t, errorResponse["message"].(string), tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)

			// Check response body
			if tc.checkResponse != nil {
				var creatures []db.Creature
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &creatures))
				tc.checkResponse(t, creatures, rec)
			}
		})
	}
}
