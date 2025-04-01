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
	"github.com/sergot/tibiacores/backend/services/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newMockEmailService(ctrl *gomock.Controller) *mock.MockEmailServiceInterface {
	return mock.NewMockEmailServiceInterface(ctrl)
}

func TestGetPendingSuggestions(t *testing.T) {
	testCases := []struct {
		name          string
		setupAuth     func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, userID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response []db.GetPendingSuggestionsForUserRow)
	}{
		{
			name: "Success",
			setupAuth: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				store.EXPECT().
					GetPendingSuggestionsForUser(gomock.Any(), userID).
					Return([]db.GetPendingSuggestionsForUserRow{
						{
							CharacterID:     uuid.New(),
							CharacterName:   "Character1",
							SuggestionCount: 2,
						},
						{
							CharacterID:     uuid.New(),
							CharacterName:   "Character2",
							SuggestionCount: 1,
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []db.GetPendingSuggestionsForUserRow) {
				require.Equal(t, 2, len(response))
				require.Equal(t, "Character1", response[0].CharacterName)
				require.Equal(t, int64(2), response[0].SuggestionCount)
				require.Equal(t, "Character2", response[1].CharacterName)
				require.Equal(t, int64(1), response[1].SuggestionCount)
			},
		},
		{
			name: "Invalid User ID",
			setupAuth: func(c echo.Context) {
				c.Set("user_id", "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "invalid user ID format",
		},
		{
			name: "Database Error",
			setupAuth: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				store.EXPECT().
					GetPendingSuggestionsForUser(gomock.Any(), userID).
					Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get pending suggestions",
		},
		{
			name: "No Pending Suggestions",
			setupAuth: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				store.EXPECT().
					GetPendingSuggestionsForUser(gomock.Any(), userID).
					Return([]db.GetPendingSuggestionsForUserRow{}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []db.GetPendingSuggestionsForUserRow) {
				require.Empty(t, response)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			emailService := newMockEmailService(ctrl)
			userID := uuid.New()

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/api/pending-suggestions", nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.Set("user_id", userID.String())

			// Custom auth setup if needed
			if tc.setupAuth != nil {
				tc.setupAuth(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, userID)

			// Execute handler
			h := handlers.NewUsersHandler(store, emailService)
			err := h.GetPendingSuggestions(c)

			// Check for expected error response
			if tc.expectedError != "" {
				httpError, ok := err.(*echo.HTTPError)
				require.True(t, ok)
				require.Equal(t, tc.expectedCode, httpError.Code)
				require.Contains(t, httpError.Message, tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)

			// Check response body
			if tc.checkResponse != nil {
				var response []db.GetPendingSuggestionsForUserRow
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response)
			}
		})
	}
}
