package handlers_test

import (
	"bytes"
	"database/sql"
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
	"github.com/sergot/tibiacores/backend/middleware"
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
			expectedError: "Invalid user ID format",
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
			expectedError: "Failed to get pending suggestions",
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
				var response []db.GetPendingSuggestionsForUserRow
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response)
			}
		})
	}
}

func TestAddCharacterSoulcore(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context, reqBody *bytes.Buffer)
		setupMocks    func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Verify character belongs to user
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Add soulcore to character
				store.EXPECT().
					AddCharacterSoulcore(gomock.Any(), db.AddCharacterSoulcoreParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid Character ID",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid character ID",
		},
		{
			name: "Invalid Request Body",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString("{invalid json")
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
		},
		{
			name: "Character Not Found",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "Character not found",
		},
		{
			name: "Character Belongs To Different User",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: uuid.New(), // Different user
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Character does not belong to user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			emailService := newMockEmailService(ctrl)
			characterID := uuid.New()
			creatureID := uuid.MustParse("808c0ee1-7b92-4795-b56f-20537aa46e0a") // Use consistent creature ID
			userID := uuid.New()

			// Create HTTP request
			reqBody := bytes.NewBuffer([]byte(`{"creature_id": "808c0ee1-7b92-4795-b56f-20537aa46e0a"}`))
			req := httptest.NewRequest(http.MethodPost, "/api/characters/"+characterID.String()+"/soulcores", reqBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/characters/:id/soulcores")
			c.SetParamNames("id")
			c.SetParamValues(characterID.String())
			c.Set("user_id", userID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c, reqBody)
			}

			// Setup mock expectations
			tc.setupMocks(store, characterID, creatureID, userID)

			// Execute handler
			h := handlers.NewUsersHandler(store, emailService)
			err := h.AddCharacterSoulcore(c)

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
		})
	}
}
