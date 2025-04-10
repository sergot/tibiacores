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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/auth"
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

// MustHashPassword is a helper function for tests to hash passwords
// It panics if password hashing fails, which should not happen in tests
func MustHashPassword(password string) string {
	hashed, err := auth.HashPassword(password)
	if err != nil {
		panic("failed to hash password for test: " + err.Error())
	}
	return hashed
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

func TestGetUser(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// Both IDs should be the same in success case
				require.Equal(t, requestedUserID, authedUserID)

				// Setup mock user response
				store.EXPECT().
					GetUserByID(gomock.Any(), requestedUserID).
					Return(db.User{
						ID:            requestedUserID,
						Email:         pgtype.Text{String: "test@example.com", Valid: true},
						EmailVerified: true,
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				require.Equal(t, "test@example.com", response["email"])
				require.Equal(t, true, response["email_verified"])
			},
		},
		{
			name: "Invalid Requested User ID",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid user ID",
		},
		{
			name: "Invalid Auth User ID in Context",
			setupRequest: func(c echo.Context) {
				c.Set("user_id", "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user ID format",
		},
		{
			name: "No Auth User ID in Context",
			setupRequest: func(c echo.Context) {
				// Remove user_id from context
				c.Set("user_id", nil)
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user authentication",
		},
		{
			name: "Unauthorized Access - Different User",
			setupRequest: func(c echo.Context) {
				// Default request is fine
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// We're explicitly testing that these IDs are different
				require.NotEqual(t, requestedUserID, authedUserID)

				// No database calls should be made
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Cannot access other users' details",
		},
		{
			name: "User Not Found",
			setupRequest: func(c echo.Context) {
				// Default request is fine
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// Both IDs should be the same
				require.Equal(t, requestedUserID, authedUserID)

				store.EXPECT().
					GetUserByID(gomock.Any(), requestedUserID).
					Return(db.User{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "User not found",
		},
		{
			name: "Database Error",
			setupRequest: func(c echo.Context) {
				// Default request is fine
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// Both IDs should be the same
				require.Equal(t, requestedUserID, authedUserID)

				store.EXPECT().
					GetUserByID(gomock.Any(), requestedUserID).
					Return(db.User{}, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to get user details",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			emailService := newMockEmailService(ctrl)

			// Setup default user IDs
			requestedUserID := uuid.New()
			authedUserID := requestedUserID

			// For the "Unauthorized Access" test case, we need different IDs
			if tc.name == "Unauthorized Access - Different User" {
				authedUserID = uuid.New() // Different ID for auth user
			}

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/api/users/"+requestedUserID.String(), nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/users/:user_id")
			c.SetParamNames("user_id")
			c.SetParamValues(requestedUserID.String())
			c.Set("user_id", authedUserID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, requestedUserID, authedUserID)

			// Execute handler
			h := handlers.NewUsersHandler(store, emailService)
			err := h.GetUser(c)

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
				var response map[string]interface{}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response)
			}
		})
	}
}

func TestGetUserLists(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response []db.GetUserListsRow)
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// Both IDs should be the same in success case
				require.Equal(t, requestedUserID, authedUserID)

				// Setup mock lists response
				store.EXPECT().
					GetUserLists(gomock.Any(), requestedUserID).
					Return([]db.GetUserListsRow{
						{
							ID:       uuid.New(),
							Name:     "List 1",
							AuthorID: requestedUserID,
						},
						{
							ID:       uuid.New(),
							Name:     "List 2",
							AuthorID: uuid.New(), // Different author, user is a member
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []db.GetUserListsRow) {
				require.Equal(t, 2, len(response))
				require.Equal(t, "List 1", response[0].Name)
				require.Equal(t, "List 2", response[1].Name)
			},
		},
		{
			name: "Invalid Requested User ID",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid user ID",
		},
		{
			name: "Invalid Auth User ID in Context",
			setupRequest: func(c echo.Context) {
				c.Set("user_id", "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user ID format",
		},
		{
			name: "No Auth User ID in Context",
			setupRequest: func(c echo.Context) {
				// Remove user_id from context
				c.Set("user_id", nil)
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user authentication",
		},
		{
			name: "Unauthorized Access - Different User",
			setupRequest: func(c echo.Context) {
				// Default request is fine
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// We're explicitly testing that these IDs are different
				require.NotEqual(t, requestedUserID, authedUserID)

				// No database calls should be made
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Cannot access other users' lists",
		},
		{
			name: "Database Error",
			setupRequest: func(c echo.Context) {
				// Default request is fine
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// Both IDs should be the same
				require.Equal(t, requestedUserID, authedUserID)

				store.EXPECT().
					GetUserLists(gomock.Any(), requestedUserID).
					Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to get lists",
		},
		{
			name: "No Lists",
			setupRequest: func(c echo.Context) {
				// Default request is fine
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// Both IDs should be the same
				require.Equal(t, requestedUserID, authedUserID)

				store.EXPECT().
					GetUserLists(gomock.Any(), requestedUserID).
					Return([]db.GetUserListsRow{}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []db.GetUserListsRow) {
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

			// Setup default user IDs
			requestedUserID := uuid.New()
			authedUserID := requestedUserID

			// For the "Unauthorized Access" test case, we need different IDs
			if tc.name == "Unauthorized Access - Different User" {
				authedUserID = uuid.New() // Different ID for auth user
			}

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/api/users/"+requestedUserID.String()+"/lists", nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/users/:user_id/lists")
			c.SetParamNames("user_id")
			c.SetParamValues(requestedUserID.String())
			c.Set("user_id", authedUserID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, requestedUserID, authedUserID)

			// Execute handler
			h := handlers.NewUsersHandler(store, emailService)
			err := h.GetUserLists(c)

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
				var response []db.GetUserListsRow
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response)
			}
		})
	}
}

func TestGetCharactersByUserId(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response []db.Character)
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// Both IDs should be the same in success case
				require.Equal(t, requestedUserID, authedUserID)

				// Setup mock characters response
				store.EXPECT().
					GetCharactersByUserID(gomock.Any(), requestedUserID).
					Return([]db.Character{
						{
							ID:     uuid.New(),
							UserID: requestedUserID,
							Name:   "Character1",
							World:  "Antica",
						},
						{
							ID:     uuid.New(),
							UserID: requestedUserID,
							Name:   "Character2",
							World:  "Secura",
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []db.Character) {
				require.Equal(t, 2, len(response))
				require.Equal(t, "Character1", response[0].Name)
				require.Equal(t, "Antica", response[0].World)
				require.Equal(t, "Character2", response[1].Name)
				require.Equal(t, "Secura", response[1].World)
			},
		},
		{
			name: "Invalid Requested User ID",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid user ID",
		},
		{
			name: "Invalid Auth User ID in Context",
			setupRequest: func(c echo.Context) {
				c.Set("user_id", "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user ID format",
		},
		{
			name: "No Auth User ID in Context",
			setupRequest: func(c echo.Context) {
				// Remove user_id from context
				c.Set("user_id", nil)
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user authentication",
		},
		{
			name: "Unauthorized Access - Different User",
			setupRequest: func(c echo.Context) {
				// Default request is fine
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// We're explicitly testing that these IDs are different
				require.NotEqual(t, requestedUserID, authedUserID)

				// No database calls should be made
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Cannot access other users' characters",
		},
		{
			name: "Database Error",
			setupRequest: func(c echo.Context) {
				// Default request is fine
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// Both IDs should be the same
				require.Equal(t, requestedUserID, authedUserID)

				store.EXPECT().
					GetCharactersByUserID(gomock.Any(), requestedUserID).
					Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to get characters",
		},
		{
			name: "No Characters",
			setupRequest: func(c echo.Context) {
				// Default request is fine
			},
			setupMocks: func(store *mockdb.MockStore, requestedUserID uuid.UUID, authedUserID uuid.UUID) {
				// Both IDs should be the same
				require.Equal(t, requestedUserID, authedUserID)

				store.EXPECT().
					GetCharactersByUserID(gomock.Any(), requestedUserID).
					Return([]db.Character{}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []db.Character) {
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

			// Setup default user IDs
			requestedUserID := uuid.New()
			authedUserID := requestedUserID

			// For the "Unauthorized Access" test case, we need different IDs
			if tc.name == "Unauthorized Access - Different User" {
				authedUserID = uuid.New() // Different ID for auth user
			}

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/api/users/"+requestedUserID.String()+"/characters", nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/users/:user_id/characters")
			c.SetParamNames("user_id")
			c.SetParamValues(requestedUserID.String())
			c.Set("user_id", authedUserID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, requestedUserID, authedUserID)

			// Execute handler
			h := handlers.NewUsersHandler(store, emailService)
			err := h.GetCharactersByUserId(c)

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
				var response []db.Character
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response)
			}
		})
	}
}

func TestVerifyEmail(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, userID uuid.UUID, token uuid.UUID)
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID, token uuid.UUID) {
				store.EXPECT().
					VerifyEmail(gomock.Any(), db.VerifyEmailParams{
						ID:                     userID,
						EmailVerificationToken: token,
					}).
					Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid User ID",
			setupRequest: func(c echo.Context) {
				c.QueryParams().Set("user_id", "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID, token uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid user ID",
		},
		{
			name: "Invalid Token",
			setupRequest: func(c echo.Context) {
				c.QueryParams().Set("token", "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID, token uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid verification token",
		},
		{
			name: "Missing Token",
			setupRequest: func(c echo.Context) {
				c.QueryParams().Del("token")
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID, token uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid verification token",
		},
		{
			name: "Verification Failed",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID, token uuid.UUID) {
				store.EXPECT().
					VerifyEmail(gomock.Any(), db.VerifyEmailParams{
						ID:                     userID,
						EmailVerificationToken: token,
					}).
					Return(errors.New("token invalid or expired"))
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid or expired verification token",
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
			token := uuid.New()

			// Create HTTP request with query parameters
			req := httptest.NewRequest(http.MethodGet, "/api/verify-email?user_id="+userID.String()+"&token="+token.String(), nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, userID, token)

			// Execute handler
			h := handlers.NewUsersHandler(store, emailService)
			err := h.VerifyEmail(c)

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

func TestGetCharacter(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response db.Character)
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response db.Character) {
				require.Equal(t, "TestCharacter", response.Name)
				require.Equal(t, "Antica", response.World)
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
			expectedError: "Invalid character ID",
		},
		{
			name: "Invalid User ID in Context",
			setupRequest: func(c echo.Context) {
				c.Set("user_id", "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user ID format",
		},
		{
			name: "No User ID in Context",
			setupRequest: func(c echo.Context) {
				// Remove user_id from context
				c.Set("user_id", nil)
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user authentication",
		},
		{
			name: "Character Not Found",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "Character not found",
		},
		{
			name: "Database Error",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{}, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to get character",
		},
		{
			name: "Character Belongs to Different User",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: uuid.New(), // Different user ID
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
			userID := uuid.New()

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/api/characters/"+characterID.String(), nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/characters/:id")
			c.SetParamNames("id")
			c.SetParamValues(characterID.String())
			c.Set("user_id", userID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, characterID, userID)

			// Execute handler
			h := handlers.NewUsersHandler(store, emailService)
			err := h.GetCharacter(c)

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
				var response db.Character
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response)
			}
		})
	}
}

func TestGetCharacterSoulcores(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response []db.GetCharacterSoulcoresRow)
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// First verify that the character belongs to the user
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Then get the soulcores for the character
				store.EXPECT().
					GetCharacterSoulcores(gomock.Any(), characterID).
					Return([]db.GetCharacterSoulcoresRow{
						{
							CharacterID:  characterID,
							CreatureID:   uuid.New(),
							CreatureName: "Dragon",
						},
						{
							CharacterID:  characterID,
							CreatureID:   uuid.New(),
							CreatureName: "Giant Spider",
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []db.GetCharacterSoulcoresRow) {
				require.Equal(t, 2, len(response))
				require.Equal(t, "Dragon", response[0].CreatureName)
				require.Equal(t, "Giant Spider", response[1].CreatureName)
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
			expectedError: "Invalid character ID",
		},
		{
			name: "Invalid User ID in Context",
			setupRequest: func(c echo.Context) {
				c.Set("user_id", "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user ID format",
		},
		{
			name: "No User ID in Context",
			setupRequest: func(c echo.Context) {
				// Remove user_id from context
				c.Set("user_id", nil)
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user authentication",
		},
		{
			name: "Character Not Found",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "Character not found",
		},
		{
			name: "Character Belongs to Different User",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: uuid.New(), // Different user ID
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Character does not belong to user",
		},
		{
			name: "Database Error - GetCharacter",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{}, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to get character",
		},
		{
			name: "Database Error - GetCharacterSoulcores",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// First verify that the character belongs to the user
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Then fail to get the soulcores
				store.EXPECT().
					GetCharacterSoulcores(gomock.Any(), characterID).
					Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to get character soulcores",
		},
		{
			name: "No Soulcores",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// First verify that the character belongs to the user
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Then return empty list of soulcores
				store.EXPECT().
					GetCharacterSoulcores(gomock.Any(), characterID).
					Return([]db.GetCharacterSoulcoresRow{}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []db.GetCharacterSoulcoresRow) {
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

			characterID := uuid.New()
			userID := uuid.New()

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/api/characters/"+characterID.String()+"/soulcores", nil)
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
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, characterID, userID)

			// Execute handler
			h := handlers.NewUsersHandler(store, emailService)
			err := h.GetCharacterSoulcores(c)

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
				var response []db.GetCharacterSoulcoresRow
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response)
			}
		})
	}
}

func TestRemoveCharacterSoulcore(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
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

				// Remove soulcore from character
				store.EXPECT().
					RemoveCharacterSoulcore(gomock.Any(), db.RemoveCharacterSoulcoreParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid Character ID",
			setupRequest: func(c echo.Context) {
				c.SetParamNames("id", "creature_id")
				c.SetParamValues("invalid-uuid", "11111111-1111-1111-1111-111111111111")
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid character ID",
		},
		{
			name: "Invalid Creature ID",
			setupRequest: func(c echo.Context) {
				c.SetParamNames("id", "creature_id")
				c.SetParamValues(uuid.New().String(), "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid creature ID",
		},
		{
			name: "Invalid User ID in Context",
			setupRequest: func(c echo.Context) {
				c.Set("user_id", "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user ID format",
		},
		{
			name: "No User ID in Context",
			setupRequest: func(c echo.Context) {
				// Remove user_id from context
				c.Set("user_id", nil)
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user authentication",
		},
		{
			name: "Character Not Found",
			setupRequest: func(c echo.Context) {
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
			setupRequest: func(c echo.Context) {
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
		{
			name: "Database Error - GetCharacter",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{}, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to get character",
		},
		{
			name: "Database Error - RemoveCharacterSoulcore",
			setupRequest: func(c echo.Context) {
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

				// Fail to remove the soulcore
				store.EXPECT().
					RemoveCharacterSoulcore(gomock.Any(), db.RemoveCharacterSoulcoreParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to remove soul core",
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
			creatureID := uuid.New()
			userID := uuid.New()

			// Create HTTP request
			req := httptest.NewRequest(http.MethodDelete, "/api/characters/"+characterID.String()+"/soulcores/"+creatureID.String(), nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/characters/:id/soulcores/:creature_id")
			c.SetParamNames("id", "creature_id")
			c.SetParamValues(characterID.String(), creatureID.String())
			c.Set("user_id", userID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, characterID, creatureID, userID)

			// Execute handler
			h := handlers.NewUsersHandler(store, emailService)
			err := h.RemoveCharacterSoulcore(c)

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

func TestLogin(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context, reqBody *bytes.Buffer)
		setupMocks    func(store *mockdb.MockStore)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response map[string]interface{}, headers http.Header)
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString(`{"email":"test@example.com","password":"password123"}`)
			},
			setupMocks: func(store *mockdb.MockStore) {
				email := pgtype.Text{String: "test@example.com", Valid: true}
				userID := uuid.New()

				store.EXPECT().
					GetUserByEmail(gomock.Any(), email).
					Return(db.User{
						ID:            userID,
						Email:         email,
						Password:      pgtype.Text{String: MustHashPassword("password123"), Valid: true},
						EmailVerified: true,
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}, headers http.Header) {
				require.NotEmpty(t, response["id"])
				require.Equal(t, true, response["has_email"])
				require.NotEmpty(t, headers.Get("X-Auth-Token"))
			},
		},
		{
			name: "Invalid Request Body",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString(`{invalid json`)
			},
			setupMocks: func(store *mockdb.MockStore) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
		},
		{
			name: "Missing Credentials",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString(`{"email":"","password":""}`)
			},
			setupMocks: func(store *mockdb.MockStore) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Email and password are required",
		},
		{
			name: "User Not Found",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString(`{"email":"nonexistent@example.com","password":"password123"}`)
			},
			setupMocks: func(store *mockdb.MockStore) {
				email := pgtype.Text{String: "nonexistent@example.com", Valid: true}

				store.EXPECT().
					GetUserByEmail(gomock.Any(), email).
					Return(db.User{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid email or password",
		},
		{
			name: "Incorrect Password",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString(`{"email":"test@example.com","password":"wrong_password"}`)
			},
			setupMocks: func(store *mockdb.MockStore) {
				email := pgtype.Text{String: "test@example.com", Valid: true}
				userID := uuid.New()

				store.EXPECT().
					GetUserByEmail(gomock.Any(), email).
					Return(db.User{
						ID:            userID,
						Email:         email,
						Password:      pgtype.Text{String: MustHashPassword("correct_password"), Valid: true},
						EmailVerified: true,
					}, nil)
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid email or password",
		},
		{
			name: "Database Error",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString(`{"email":"test@example.com","password":"password123"}`)
			},
			setupMocks: func(store *mockdb.MockStore) {
				email := pgtype.Text{String: "test@example.com", Valid: true}

				store.EXPECT().
					GetUserByEmail(gomock.Any(), email).
					Return(db.User{}, errors.New("database error"))
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid email or password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			emailService := newMockEmailService(ctrl)

			// Create HTTP request
			reqBody := bytes.NewBuffer([]byte(`{"email":"test@example.com","password":"password123"}`))
			req := httptest.NewRequest(http.MethodPost, "/api/login", reqBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c, reqBody)
			}

			// Setup mock expectations
			tc.setupMocks(store)

			// Execute handler
			h := handlers.NewUsersHandler(store, emailService)
			err := h.Login(c)

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

			// Check response body and headers
			if tc.checkResponse != nil {
				var response map[string]interface{}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response, rec.Header())
			}
		})
	}
}

func TestSignup(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context, reqBody *bytes.Buffer)
		setupMocks    func(store *mockdb.MockStore, emailService *mock.MockEmailServiceInterface)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response map[string]interface{}, headers http.Header)
	}{
		{
			name: "Success - New User",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString(`{"email":"new@example.com","password":"password123"}`)
			},
			setupMocks: func(store *mockdb.MockStore, emailService *mock.MockEmailServiceInterface) {
				email := pgtype.Text{String: "new@example.com", Valid: true}
				userID := uuid.New()

				store.EXPECT().
					GetUserByEmail(gomock.Any(), email).
					Return(db.User{}, sql.ErrNoRows)

				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.CreateUserParams) (db.User, error) {
						require.Equal(t, email, params.Email)
						require.True(t, params.Password.Valid)
						require.True(t, params.EmailVerificationToken.String() != "")
						require.True(t, params.EmailVerificationExpiresAt.Valid)

						return db.User{
							ID:            userID,
							Email:         email,
							Password:      params.Password,
							EmailVerified: false,
						}, nil
					})

				emailService.EXPECT().
					SendVerificationEmail(gomock.Any(), email.String, gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}, headers http.Header) {
				require.NotEmpty(t, response["id"])
				require.Equal(t, true, response["has_email"])
				require.NotEmpty(t, headers.Get("X-Auth-Token"))
			},
		},
		{name: "Success - User With ID in Request",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString(`{"email":"existing@example.com","password":"password123", "user_id":"b5a9b0c3-1d2e-3f4a-5b6c-7d8e9f0a1b2c"}`)
			},
			setupMocks: func(store *mockdb.MockStore, emailService *mock.MockEmailServiceInterface) {
				email := pgtype.Text{String: "existing@example.com", Valid: true}
				userID, _ := uuid.Parse("b5a9b0c3-1d2e-3f4a-5b6c-7d8e9f0a1b2c")

				// Check if user exists with this email
				store.EXPECT().
					GetUserByEmail(gomock.Any(), email).
					Return(db.User{}, sql.ErrNoRows)

				// Create new user with the provided user_id
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.CreateUserParams) (db.User, error) {
						require.Equal(t, email, params.Email)
						require.True(t, params.Password.Valid)
						require.True(t, params.EmailVerificationToken.String() != "")
						require.True(t, params.EmailVerificationExpiresAt.Valid)

						return db.User{
							ID:            userID,
							Email:         email,
							Password:      params.Password,
							EmailVerified: false,
						}, nil
					})

				emailService.EXPECT().
					SendVerificationEmail(gomock.Any(), email.String, gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}, headers http.Header) {
				require.NotEmpty(t, response["id"])
				require.Equal(t, true, response["has_email"])
				require.NotEmpty(t, headers.Get("X-Auth-Token"))
			},
		},
		{
			name: "Success - Anonymous User Found By Email",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString(`{"email":"anonymous@example.com","password":"password123"}`)
			},
			setupMocks: func(store *mockdb.MockStore, emailService *mock.MockEmailServiceInterface) {
				email := pgtype.Text{String: "anonymous@example.com", Valid: true}
				userID := uuid.New()

				// Anonymous user already exists with this email
				store.EXPECT().
					GetUserByEmail(gomock.Any(), email).
					Return(db.User{
						ID:          userID,
						Email:       email,
						IsAnonymous: true,
					}, nil)

				store.EXPECT().
					MigrateAnonymousUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, params db.MigrateAnonymousUserParams) (db.User, error) {
						require.Equal(t, email, params.Email)
						require.True(t, params.Password.Valid)

						return db.User{
							ID:            userID,
							Email:         email,
							Password:      params.Password,
							EmailVerified: false,
						}, nil
					})

				emailService.EXPECT().
					SendVerificationEmail(gomock.Any(), email.String, gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}, headers http.Header) {
				require.NotEmpty(t, response["id"])
				require.Equal(t, true, response["has_email"])
				require.NotEmpty(t, headers.Get("X-Auth-Token"))
			},
		},
		{
			name: "Invalid Request Body",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString(`{invalid json`)
			},
			setupMocks: func(store *mockdb.MockStore, emailService *mock.MockEmailServiceInterface) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
		},
		{
			name: "Missing Credentials",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString(`{"email":"","password":""}`)
			},
			setupMocks: func(store *mockdb.MockStore, emailService *mock.MockEmailServiceInterface) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Email and password are required",
		},
		{
			name: "Email Already In Use",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString(`{"email":"registered@example.com","password":"password123"}`)
			},
			setupMocks: func(store *mockdb.MockStore, emailService *mock.MockEmailServiceInterface) {
				email := pgtype.Text{String: "registered@example.com", Valid: true}
				userID := uuid.New()

				store.EXPECT().
					GetUserByEmail(gomock.Any(), email).
					Return(db.User{
						ID:          userID,
						Email:       email,
						IsAnonymous: false,
					}, nil)
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Email already in use",
		},
		{
			name: "Database Error - CreateUser",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString(`{"email":"error@example.com","password":"password123"}`)
			},
			setupMocks: func(store *mockdb.MockStore, emailService *mock.MockEmailServiceInterface) {
				email := pgtype.Text{String: "error@example.com", Valid: true}

				store.EXPECT().
					GetUserByEmail(gomock.Any(), email).
					Return(db.User{}, sql.ErrNoRows)

				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(db.User{}, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to create user",
		},
		{name: "Database Error - CreateUser With UserID",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString(`{"email":"existing@example.com","password":"password123", "user_id":"b5a9b0c3-1d2e-3f4a-5b6c-7d8e9f0a1b2c"}`)
			},
			setupMocks: func(store *mockdb.MockStore, emailService *mock.MockEmailServiceInterface) {
				email := pgtype.Text{String: "existing@example.com", Valid: true}

				// Check if user exists with this email
				store.EXPECT().
					GetUserByEmail(gomock.Any(), email).
					Return(db.User{}, sql.ErrNoRows)

				// CreateUser fails with database error
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(db.User{}, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to create user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			emailService := newMockEmailService(ctrl)

			// Create HTTP request
			reqBody := bytes.NewBuffer([]byte(`{"email":"test@example.com","password":"password123"}`))
			req := httptest.NewRequest(http.MethodPost, "/api/signup", reqBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c, reqBody)
			}

			// Setup mock expectations
			tc.setupMocks(store, emailService)

			// Execute handler
			h := handlers.NewUsersHandler(store, emailService)
			err := h.Signup(c)

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

			// Check response body and headers
			if tc.checkResponse != nil {
				var response map[string]interface{}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response, rec.Header())
			}
		})
	}
}
