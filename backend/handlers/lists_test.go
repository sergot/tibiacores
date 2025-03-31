package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

func TestGetList(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, list db.List, userID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response *handlers.ListDetailResponse, list db.List)
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, list db.List, userID uuid.UUID) {
				// Get list details
				store.EXPECT().
					GetList(gomock.Any(), list.ID).
					Return(list, nil)

				// Get list members - includes current user
				store.EXPECT().
					GetListMembers(gomock.Any(), list.ID).
					Return([]db.GetListMembersRow{
						{UserID: userID, CharacterName: "TestCharacter"},
						{UserID: list.AuthorID, CharacterName: "AuthorCharacter"},
					}, nil)

				// Get list soulcores
				store.EXPECT().
					GetListSoulcores(gomock.Any(), list.ID).
					Return([]db.GetListSoulcoresRow{
						{
							CreatureID: uuid.New(),
							ListID:     list.ID,
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response *handlers.ListDetailResponse, list db.List) {
				require.Equal(t, list.ID, response.ID)
				require.Equal(t, list.Name, response.Name)
				require.Equal(t, list.World, response.World)
				require.Equal(t, 2, len(response.Members))
				require.Equal(t, 1, len(response.SoulCores))
			},
		},
		{
			name: "Invalid List ID",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, list db.List, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid list ID",
		},
		{
			name: "User Not a Member",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, list db.List, userID uuid.UUID) {
				// Get list details
				store.EXPECT().
					GetList(gomock.Any(), list.ID).
					Return(list, nil)

				// Get list members - doesn't include current user
				store.EXPECT().
					GetListMembers(gomock.Any(), list.ID).
					Return([]db.GetListMembersRow{
						{UserID: uuid.New(), CharacterName: "OtherCharacter"},
						{UserID: list.AuthorID, CharacterName: "AuthorCharacter"},
					}, nil)
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "user is not a member of this list",
		},
		{
			name: "List Not Found",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, list db.List, userID uuid.UUID) {
				// List not found
				store.EXPECT().
					GetList(gomock.Any(), list.ID).
					Return(db.List{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "list not found",
		},
		{
			name: "Error Getting List Members",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, list db.List, userID uuid.UUID) {
				// Get list details
				store.EXPECT().
					GetList(gomock.Any(), list.ID).
					Return(list, nil)

				// Error getting list members
				store.EXPECT().
					GetListMembers(gomock.Any(), list.ID).
					Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get list members",
		},
		{
			name: "Error Getting Soul Cores",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, list db.List, userID uuid.UUID) {
				// Get list details
				store.EXPECT().
					GetList(gomock.Any(), list.ID).
					Return(list, nil)

				// Get list members - includes current user
				store.EXPECT().
					GetListMembers(gomock.Any(), list.ID).
					Return([]db.GetListMembersRow{
						{UserID: userID, CharacterName: "TestCharacter"},
					}, nil)

				// Error getting soulcores
				store.EXPECT().
					GetListSoulcores(gomock.Any(), list.ID).
					Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get soul cores",
		},
		{
			name: "Empty List (No Members or Soulcores)",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, list db.List, userID uuid.UUID) {
				// Get list details
				store.EXPECT().
					GetList(gomock.Any(), list.ID).
					Return(list, nil)

				// Only the current user is a member
				store.EXPECT().
					GetListMembers(gomock.Any(), list.ID).
					Return([]db.GetListMembersRow{
						{UserID: userID, CharacterName: "TestCharacter"},
					}, nil)

				// Empty soulcores list
				store.EXPECT().
					GetListSoulcores(gomock.Any(), list.ID).
					Return([]db.GetListSoulcoresRow{}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response *handlers.ListDetailResponse, list db.List) {
				require.Equal(t, list.ID, response.ID)
				require.Equal(t, 1, len(response.Members))
				require.Equal(t, 0, len(response.SoulCores))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			userID := uuid.New()
			list := db.List{
				ID:        uuid.New(),
				AuthorID:  uuid.New(),
				Name:      "Test List",
				World:     "Antica",
				ShareCode: uuid.New(),
			}

			// Create HTTP request
			url := fmt.Sprintf("/api/lists/%s", list.ID.String())
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/lists/:id")
			c.Set("user_id", userID.String())
			c.SetParamNames("id")
			c.SetParamValues(list.ID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, list, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.GetList(c)

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
				var response handlers.ListDetailResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, &response, list)
			}
		})
	}
}

func TestGetCharacterSuggestions(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response []db.GetCharacterSuggestionsRow)
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// Verify character belongs to user - expect to be called exactly once
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil).
					Times(1)

				// Get character suggestions - expect to be called exactly once
				store.EXPECT().
					GetCharacterSuggestions(gomock.Any(), characterID).
					Return([]db.GetCharacterSuggestionsRow{
						{
							CharacterID:  characterID,
							CreatureID:   uuid.New(),
							ListID:       uuid.New(),
							CreatureName: "Dragon",
						},
						{
							CharacterID:  characterID,
							CreatureID:   uuid.New(),
							ListID:       uuid.New(),
							CreatureName: "Demon",
						},
					}, nil).
					Times(1)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []db.GetCharacterSuggestionsRow) {
				require.Equal(t, 2, len(response))
				require.Equal(t, "Dragon", response[0].CreatureName)
				require.Equal(t, "Demon", response[1].CreatureName)
			},
		},
		{
			name: "Invalid Character ID",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid character ID",
		},
		{
			name: "Character Not Found",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// Character not found
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get character",
		},
		{
			name: "Character Belongs To Different User",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// Return a character with different userID
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: uuid.New(), // Different user
						Name:   "OtherCharacter",
						World:  "Antica",
					}, nil)
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "character does not belong to user",
		},
		{
			name: "Database Error Getting Suggestions",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// Verify character belongs to user
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Database error getting suggestions
				store.EXPECT().
					GetCharacterSuggestions(gomock.Any(), characterID).
					Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get suggestions",
		},
		{
			name: "No Suggestions",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// Verify character belongs to user
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Empty suggestions list
				store.EXPECT().
					GetCharacterSuggestions(gomock.Any(), characterID).
					Return([]db.GetCharacterSuggestionsRow{}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []db.GetCharacterSuggestionsRow) {
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
			characterID := uuid.New()
			userID := uuid.New()

			// Create HTTP request
			url := fmt.Sprintf("/api/characters/%s/suggestions", characterID.String())
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/characters/:id/suggestions")
			c.Set("user_id", userID.String())
			c.SetParamNames("id")
			c.SetParamValues(characterID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, characterID, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.GetCharacterSuggestions(c)

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
				var response []db.GetCharacterSuggestionsRow
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response)
			}
		})
	}
}

func TestAcceptSoulcoreSuggestion(t *testing.T) {
	// Test cases
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
				// Verify character belongs to user - expect to be called exactly once
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil).
					Times(1)

				// Add soulcore to character - expect to be called exactly once
				store.EXPECT().
					AddCharacterSoulcore(gomock.Any(), db.AddCharacterSoulcoreParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(nil).
					Times(1)

				// Delete the suggestion - expect to be called exactly once
				store.EXPECT().
					DeleteSoulcoreSuggestion(gomock.Any(), db.DeleteSoulcoreSuggestionParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(nil).
					Times(1)
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
			expectedError: "invalid character ID",
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
			expectedError: "invalid request body",
		},
		{
			name: "Character Not Found",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Character not found
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get character",
		},
		{
			name: "Character Belongs To Different User",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Return a character with different userID
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: uuid.New(), // Different user
						Name:   "OtherCharacter",
						World:  "Antica",
					}, nil)
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "character does not belong to user",
		},
		{
			name: "Error Adding Soulcore",
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

				// Error adding soulcore
				store.EXPECT().
					AddCharacterSoulcore(gomock.Any(), db.AddCharacterSoulcoreParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to add soulcore to character",
		},
		{
			name: "Error Deleting Suggestion",
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

				// Add soulcore successfully
				store.EXPECT().
					AddCharacterSoulcore(gomock.Any(), db.AddCharacterSoulcoreParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(nil)

				// Error deleting suggestion
				store.EXPECT().
					DeleteSoulcoreSuggestion(gomock.Any(), db.DeleteSoulcoreSuggestionParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to delete suggestion",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			characterID := uuid.New()
			creatureID := uuid.New()
			userID := uuid.New()

			// Create request body
			reqBody := &bytes.Buffer{}
			json.NewEncoder(reqBody).Encode(map[string]interface{}{
				"creature_id": creatureID,
			})

			// Create HTTP request
			url := fmt.Sprintf("/api/characters/%s/suggestions/accept", characterID.String())
			req := httptest.NewRequest(http.MethodPost, url, reqBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/characters/:id/suggestions/accept")
			c.Set("user_id", userID.String())
			c.SetParamNames("id")
			c.SetParamValues(characterID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c, reqBody)
			}

			// Setup mock expectations
			tc.setupMocks(store, characterID, creatureID, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.AcceptSoulcoreSuggestion(c)

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
		})
	}
}

func TestDismissSoulcoreSuggestion(t *testing.T) {
	// Test cases
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
				// Verify character belongs to user - expect to be called exactly once
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil).
					Times(1)

				// Delete the suggestion - expect to be called exactly once
				store.EXPECT().
					DeleteSoulcoreSuggestion(gomock.Any(), db.DeleteSoulcoreSuggestionParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(nil).
					Times(1)
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
			expectedError: "invalid character ID",
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
			expectedError: "invalid request body",
		},
		{
			name: "Character Not Found",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Character not found
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get character",
		},
		{
			name: "Character Belongs To Different User",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Return a character with different userID
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: uuid.New(), // Different user
						Name:   "OtherCharacter",
						World:  "Antica",
					}, nil)
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "character does not belong to user",
		},
		{
			name: "Error Deleting Suggestion",
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

				// Error deleting suggestion
				store.EXPECT().
					DeleteSoulcoreSuggestion(gomock.Any(), db.DeleteSoulcoreSuggestionParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to delete suggestion",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			characterID := uuid.New()
			creatureID := uuid.New()
			userID := uuid.New()

			// Create request body
			reqBody := &bytes.Buffer{}
			json.NewEncoder(reqBody).Encode(map[string]interface{}{
				"creature_id": creatureID,
			})

			// Create HTTP request
			url := fmt.Sprintf("/api/characters/%s/suggestions/dismiss", characterID.String())
			req := httptest.NewRequest(http.MethodPost, url, reqBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/characters/:id/suggestions/dismiss")
			c.Set("user_id", userID.String())
			c.SetParamNames("id")
			c.SetParamValues(characterID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c, reqBody)
			}

			// Setup mock expectations
			tc.setupMocks(store, characterID, creatureID, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.DismissSoulcoreSuggestion(c)

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
		})
	}
}
